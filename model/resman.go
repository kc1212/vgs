package model

import (
	"log"
	"time"
)

import "github.com/kc1212/vgs/common"
import "github.com/kc1212/vgs/discosrv"

type ResMan struct {
	common.Node
	workers      []Worker
	gsNodes      *common.SyncedSet
	jobsChan     chan Job
	discosrvAddr string
}

func InitResMan(n int, id int, addr string, dsAddr string) ResMan {
	return ResMan{
		common.Node{ID: id, Addr: addr, Type: common.RMNode},
		make([]Worker, n),
		&common.SyncedSet{S: make(map[string]common.IntClient)},
		make(chan Job, 1000),
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
	go rm.schedule()
	rm.startMainLoop()
}

// AddJob RPC call
func (rm *ResMan) AddJob(jobs *[]Job, reply *int) error {
	log.Printf("Jobs received %v\n", *jobs)

	// make a channel of jobs, and then schedule them
	for _, j := range *jobs {
		rm.jobsChan <- j
	}
	*reply = 1
	return nil
}

// RecvMsg PRC call
func (rm *ResMan) RecvMsg(args *RPCArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
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

func (rm *ResMan) notifyAndPopulateGSs(nodes []string) {
	// NOTE: does RM doesn't use a clock, hence the zero
	arg := RPCArgs{rm.ID, rm.Addr, common.RMUpMsg, 0}
	for _, node := range nodes {
		id, e := rpcSendMsgToGS(node, &arg)
		if e == nil {
			rm.gsNodes.SetInt(node, int64(id))
		}
	}
}

func (rm *ResMan) startMainLoop() {
	log.Printf("Startin main loop\n")
	for {
		time.Sleep(time.Second)
		// TODO receive messages from grid scheduler
		// TODO update status of workers
		// TODO schedule jobs
	}
}

// computeCapacity computes the collective remaining capacity of the worker nodes.
func (rm *ResMan) computeCapacity() int {
	cnt := 0
	for _, w := range rm.workers {
		if !w.isRunning() {
			cnt++
		}
	}
	return cnt
}

// greedy scheduler
func (rm *ResMan) schedule() {
	for {
		select {
		case j := <-rm.jobsChan:
			i := rm.nextFreeNode()
			if i == -1 {
				log.Panic("Couldn't find free node")
			}
			rm.workers[i].startJob(j)
		}
	}
}

func (rm *ResMan) nextFreeNode() int {
	// TODO looping over all workers is inefficient
	// because the low idx workers are always assigned first
	for i := range rm.workers {
		if !rm.workers[i].isRunning() {
			return i
		}
	}
	return -1
}
