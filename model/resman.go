package model

import (
	"log"
	"net/rpc"
	"sync"
	"time"
)

import "github.com/kc1212/virtual-grid/common"
import "github.com/kc1212/virtual-grid/discosrv"

type ResMan struct {
	common.Node
	n             int // number of workers
	gsNodes       *common.SyncedSet
	tasksChan     chan WorkerTask
	completedChan chan int64
	capReq        chan int
	capResp       chan int
	discosrvAddr  string
}

func InitResMan(n int, id int, addr string, dsAddr string) ResMan {
	return ResMan{
		common.Node{ID: id, Addr: addr, Type: common.RMNode},
		n,
		&common.SyncedSet{S: make(map[string]common.IntClient)},
		make(chan WorkerTask, 1000),
		make(chan int64),
		make(chan int),
		make(chan int),
		dsAddr}
}

// RunResMan is the main function, it starts all its services.
func (rm *ResMan) Run() {
	reply, e := discosrv.ImAliveProbe(rm.Addr, common.RMNode, rm.discosrvAddr)
	if e != nil {
		log.Panicf("Discosrv on %v not online\n", rm.discosrvAddr)
	}
	rm.notifyAndPopulateGSs(reply.GSs)

	go discosrv.ImAlivePoll(rm.Addr, common.RMNode, rm.discosrvAddr)
	go common.RunRPC(rm, rm.Addr)
	go runWorkers(rm.n, rm.tasksChan, rm.capReq, rm.capResp, rm.completedChan)
	rm.handleCompletionMsg()
}

func (rm *ResMan) forwardJobs(jobs *[]Job) int {
	log.Printf("Forwarding %v jobs to GS\n", len(*jobs))
	// range over map is random
	reply := -1
	for k := range rm.gsNodes.GetAll() {
		remote, e := rpc.DialHTTP("tcp", k)
		if e != nil {
			log.Printf("Node %v is not online, make sure to use the correct address?\n", k)
			continue
		}
		defer remote.Close()

		if e := remote.Call("GridSdr.AddJobsViaUser", jobs, &reply); e != nil {
			log.Printf("Remote call GridSdr.AddJobsViaUser failed on %v, %v\n", k, e.Error())
		} else {
			return reply
		}
	}
	// unreachable
	log.Panic("At least one GS should be online!")
	return -1
}

// AddJob RPC call, only used by CLI
func (rm *ResMan) AddJobsViaUser(jobs *[]Job, reply *int) error {
	log.Printf("%v jobs received from user \n", len(*jobs))

	// forward the jobs to a random GS if I have no capacity, otherwise schedule them
	if rm.computeCapacity() == 0 {
		rm.forwardJobs(jobs)
	} else {
		// TODO put the scheduled jobs into GS's scheduled jobs queue
		rm.scheduleJobs(jobs)
	}
	*reply = 0
	return nil
}

// AddJob RPC call, only used by GridSdr
func (rm *ResMan) AddJob(jobs *[]Job, reply *int) error {
	log.Printf("%v jobs received \n", len(*jobs))

	rm.scheduleJobs(jobs)
	*reply = 0
	return nil
}

func (rm *ResMan) scheduleJobs(jobs *[]Job) {
	// make a channel of jobs, and then schedule them
	for _, j := range *jobs {
		// in theory the task can be arbitrary, here we just run Sleep
		task := func() (interface{}, error) {
			time.Sleep(time.Duration(j.Duration) * time.Second)
			return 0, nil
		}
		rm.tasksChan <- WorkerTask{task, j.ID}
	}
}

// RecvMsg PRC call
func (rm *ResMan) RecvMsg(args *RPCArgs, reply *int) error {
	// log.Printf("Msg received %v\n", *args)
	*reply = -1
	if args.Type == common.RMUpMsg {
		*reply = rm.ID
		rm.gsNodes.SetInt(args.Addr, int64(args.ID))

	} else if args.Type == common.GetCapacityMsg {
		*reply = rm.computeCapacity()

	} else {
		log.Panic("Invalid message!", args)
	}
	return nil
}

func (rm *ResMan) computeCapacity() int {
	rm.capReq <- 0
	cap := <-rm.capResp
	return cap
}

func (rm *ResMan) notifyAndPopulateGSs(nodes []string) {
	// NOTE: does RM doesn't use a clock, hence the zero
	arg := RPCArgs{rm.ID, rm.Addr, common.RMUpMsg, 0}
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			id, e := rpcSendMsgToGS(addr, &arg)
			if e == nil {
				rm.gsNodes.SetInt(addr, int64(id))
			}
		}(node)
	}
	wg.Wait()
}

// handleCompletionMsg runs forever to notify GSs about job completion
func (rm *ResMan) handleCompletionMsg() {
	ids := make([]int64, 0)
	mutex := sync.Mutex{}

	// update the ids array when something arrives in completedChan
	go func() {
		for {
			for id := range rm.completedChan {
				mutex.Lock()
				ids = append(ids, id)
				mutex.Unlock()
			}
		}
	}()

	// send the ids to GS every 100ms
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ids) == 0 {
			continue
		}

		mutex.Lock()
		log.Printf("Completed %v jobs.\n", len(ids))

		// range over map is random so this is ok
		for k := range rm.gsNodes.GetAll() {
			_, e := rpcSyncCompletedJobs(k, &ids)
			if e == nil {
				break
			}
		}
		ids = make([]int64, 0)
		mutex.Unlock()
	}
}
