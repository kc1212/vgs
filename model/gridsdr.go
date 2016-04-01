package model

import (
	"log"
	"math/rand"
	"net/rpc"
	"time"
)

import "github.com/kc1212/vgs/common"
import "github.com/kc1212/vgs/discosrv"

// GridSdr describes the properties of one grid scheduler
type GridSdr struct {
	common.Node
	gsNodes       *common.SyncedSet // other GridSdr's, not including myself
	rmNodes       *common.SyncedSet
	leader        string // the lead GridSdr
	jobs          []Job
	tasks         chan common.Task // these tasks require CS
	inElection    *common.SyncedVal
	mutexRespChan chan int
	mutexReqChan  chan common.Task
	mutexState    *common.SyncedVal
	clock         *common.SyncedVal
	reqClock      int64
	discosrvAddr  string
}

// RPCArgs is the arguments for RPC calls between grid schedulers
type RPCArgs struct {
	ID    int
	Addr  string
	Type  common.MsgType
	Clock int64
}

// InitGridSdr creates a grid scheduler.
func InitGridSdr(id int, addr string, dsAddr string) GridSdr {
	// NOTE: the following three values are initiated in `Run`
	gsNodes := &common.SyncedSet{S: make(map[string]common.IntClient)}
	rmNodes := &common.SyncedSet{S: make(map[string]common.IntClient)}
	var leader string

	return GridSdr{
		common.Node{id, addr, common.GSNode},
		gsNodes, rmNodes, leader,
		make([]Job, 0),
		make(chan common.Task, 100),
		&common.SyncedVal{V: false},
		make(chan int, 100),
		make(chan common.Task, 100),
		&common.SyncedVal{V: common.StateReleased},
		&common.SyncedVal{V: int64(0)},
		0,
		dsAddr,
	}
}

// Run is the main function for GridSdr, it starts all its services.
func (gs *GridSdr) Run() {
	rand.Seed(time.Now().UTC().UnixNano())

	reply, e := discosrv.ImAliveProbe(gs.Addr, gs.Type, gs.discosrvAddr)
	if e != nil {
		log.Panicf("Discosrv on %v not online\n", gs.discosrvAddr)
	}
	gs.notifyAndPopulateGSs(reply.GSs)
	gs.notifyAndPopulateRMs(reply.RMs)

	go discosrv.ImAlivePoll(gs.Addr, gs.Type, gs.discosrvAddr)
	go common.RunRPC(gs, gs.Addr)
	go gs.pollLeader()
	go gs.runTasks()

	for {
		// TODO get all the rmNodes
		// TODO arrange them in loaded order
		// TODO allocate *all* jobs
		time.Sleep(time.Second)
	}
}

func (gs *GridSdr) notifyAndPopulateGSs(nodes []string) {
	arg := RPCArgs{gs.ID, gs.Addr, common.GSUpMsg, gs.clock.Geti64()}
	for _, node := range nodes {
		id, e := sendMsgToGS(node, &arg)
		if e == nil {
			gs.gsNodes.SetInt(node, int64(id))
		}
	}
}

func (gs *GridSdr) notifyAndPopulateRMs(nodes []string) {
	args := RPCArgs{gs.ID, gs.Addr, common.RMUpMsg, gs.clock.Geti64()}
	for _, node := range nodes {
		id, e := sendMsgToRM(node, &args)
		if e == nil {
			gs.rmNodes.SetInt(node, int64(id))
		}
	}
}

// TODO we need to generalise those sendMsg/addJobs functions

func sendMsgToRM(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending message %v to %v\n", *args, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "ResMan.RecvMsg", args, &reply)
	return reply, remote.Close()
}

// addJobsToRM creates an RPC connection with a ResMan and does one remote call on AddJob.
func addJobsToRM(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending job to %v\n", addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "ResMan.AddJob", args, &reply)
	return reply, remote.Close()
}

// sendMsgToGS creates an RPC connection with another GridSdr and does one remote call on RecvMsg.
func sendMsgToGS(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending message %v to %v\n", *args, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvMsg", args, &reply)
	return reply, remote.Close()
}

// addJobsToGS is a remote call that calls `RecvJobs`.
// NOTE: this function should only be executed when CS is obtained.
func addJobsToGS(addr string, jobs *[]Job) (int, error) {
	log.Printf("Sending jobs %v, to %v\n", *jobs, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvJobs", jobs, &reply)
	return reply, remote.Close()
}

// obtainCritSection implements most of the Ricart-Agrawala algorithm, it sends the critical section request and then wait for responses until some timeout.
// Initially we set the mutexState to StateWanted, if the critical section is obtained we set it to StateHeld.
// NOTE: this function isn't designed to be thread safe, it is run periodically in `runTasks`.
func (gs *GridSdr) obtainCritSection() {
	if gs.mutexState.Get().(common.MutexState) != common.StateReleased {
		log.Panicf("Should not be in CS, state: %v\n", gs)
	}

	gs.mutexState.Set(common.StateWanted)

	// empty the channel before starting just in case
	for len(gs.mutexRespChan) > 0 {
		<-gs.mutexRespChan
	}

	gs.clock.Tick()
	successes := 0
	for k, _ := range gs.gsNodes.GetAll() {
		_, e := sendMsgToGS(k, &RPCArgs{gs.ID, gs.Addr, common.MutexReq, gs.clock.Geti64()})
		if e == nil {
			successes++
		}
	}
	gs.reqClock = gs.clock.Geti64()

	// wait until others has written to mutexRespChan or time out (2s)
	t := time.Now().Add(2 * time.Second)
	for t.After(time.Now()) {
		if len(gs.mutexRespChan) >= successes {
			break
		}
		time.Sleep(time.Microsecond)
	}

	// empty the channel
	// NOTE: nodes following the protocol shouldn't send more messages
	for len(gs.mutexRespChan) > 0 {
		<-gs.mutexRespChan
	}

	// here we're in critical section
	gs.mutexState.Set(common.StateHeld)
	log.Println("In CS!", gs.ID)
}

// releaseCritSection sets the mutexState to StateReleased and then runs all the queued requests.
func (gs *GridSdr) releaseCritSection() {
	gs.mutexState.Set(common.StateReleased)
	for len(gs.mutexReqChan) > 0 {
		resp := <-gs.mutexReqChan
		_, e := resp()
		if e != nil {
			log.Panic("task failed with", e)
		}
	}
	log.Println("Out CS!", gs.ID)
}

// elect implements the Bully algorithm.
func (gs *GridSdr) elect() {
	defer func() {
		gs.inElection.Set(false)
	}()
	gs.inElection.Set(true)

	gs.clock.Tick()
	oks := 0
	for k, v := range gs.gsNodes.GetAll() {
		if v.ID < int64(gs.ID) {
			continue // do nothing to lower ids
		}
		_, e := sendMsgToGS(k, &RPCArgs{gs.ID, gs.Addr, common.ElectionMsg, gs.clock.Geti64()})
		if e == nil {
			oks++
		}
	}

	// if no responses, then set the node itself as leader, and tell the others
	if oks == 0 {
		gs.clock.Tick()
		gs.leader = gs.Addr
		log.Printf("I'm the leader (%v).\n", gs.leader)
		for k, _ := range gs.gsNodes.GetAll() {
			args := RPCArgs{gs.ID, gs.Addr, common.CoordinateMsg, gs.clock.Geti64()}
			sendMsgToGS(k, &args) // NOTE: ok to fail the send, because nodes might be done
		}
	}

	// artificially make the election last longer so that multiple messages
	// requests won't initialise multiple election runs
	time.Sleep(time.Second)
}

// RecvMsg is called remotely, it updates the Lamport clock first and then performs tasks depending on the message type.
func (gs *GridSdr) RecvMsg(args *RPCArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
	*reply = 1
	gs.clock.Set(common.Max64(gs.clock.Geti64(), args.Clock) + 1) // update Lamport clock
	if args.Type == common.CoordinateMsg {
		gs.leader = args.Addr
		log.Printf("Leader set to %v\n", gs.leader)

	} else if args.Type == common.ElectionMsg {
		// don't start a new election if one is already running
		if !gs.inElection.Get().(bool) {
			go gs.elect()
		}
	} else if args.Type == common.MutexReq {
		go gs.respCritSection(*args)

	} else if args.Type == common.MutexResp {
		gs.mutexRespChan <- 0

	} else if args.Type == common.GSUpMsg {
		*reply = gs.ID
		gs.gsNodes.SetInt(args.Addr, int64(args.ID))

	} else if args.Type == common.RMUpMsg {
		*reply = gs.ID
		gs.rmNodes.SetInt(args.Addr, int64(args.ID))

	} else {
		log.Panic("Invalid message!", args)
	}
	return nil
}

// RecvJobs appends new jobs into the jobs queue.
// NOTE: this function should not be called directly by the client, it requires CS.
func (gs *GridSdr) RecvJobs(jobs *[]Job, reply *int) error {
	log.Printf("adding %v\n", *jobs)
	gs.jobs = append(gs.jobs, (*jobs)...)
	*reply = 0
	return nil
}

// AddJobs is called by the client to add job(s) to the tasks queue.
// NOTE: this does not add jobs to the jobs queue, that is done by `runTasks`.
func (gs *GridSdr) AddJobs(jobs *[]Job, reply *int) error {
	gs.tasks <- func() (interface{}, error) {
		// add jobs to the others
		for k, _ := range gs.gsNodes.GetAll() {
			addJobsToGS(k, jobs) // ok to fail
		}
		// add jobs to myself
		reply := -1
		e := gs.RecvJobs(jobs, &reply)
		return reply, e
	}
	return nil
}

// runTasks queries the tasks queue and if there are outstanding tasks it will request for critical and run the tasks.
func (gs *GridSdr) runTasks() {
	for {
		// sleep for 100ms
		time.Sleep(100 * time.Millisecond)

		// acquire CS, run the tasks, run for 1ms at most, then release CS
		if len(gs.tasks) > 0 {
			gs.obtainCritSection()
			t := time.Now().Add(time.Millisecond)
			for t.After(time.Now()) && len(gs.tasks) > 0 {
				task := <-gs.tasks
				_, e := task()
				if e != nil {
					log.Panic("task failed with", e)
				}
			}
			gs.releaseCritSection()
		}
	}
}

// argsIsLater checks whether args has a later Lamport clock, tie break using node ID.
func (gs *GridSdr) argsIsLater(args RPCArgs) bool {
	return gs.clock.Geti64() < args.Clock || (gs.clock.Geti64() == args.Clock && gs.ID < args.ID)
}

// respCritSection puts the critical section response into the response queue when it can't respond straight away.
func (gs *GridSdr) respCritSection(args RPCArgs) {
	resp := func() (interface{}, error) {
		sendMsgToGS(args.Addr, &RPCArgs{gs.ID, gs.Addr, common.MutexResp, gs.reqClock})
		return 0, nil
	}

	st := gs.mutexState.Get().(common.MutexState)
	if st == common.StateHeld || (st == common.StateWanted && gs.argsIsLater(args)) {
		gs.mutexReqChan <- resp
	} else {
		resp()
	}
}

// pollLeader polls the leader node and initiates the election algorithm is the leader goes offline.
func (gs *GridSdr) pollLeader() {
	for {
		time.Sleep(time.Second)

		// don't do anything if election is running or I'm leader
		if gs.inElection.Get().(bool) || gs.Addr == gs.leader {
			continue
		}

		remote, e := rpc.DialHTTP("tcp", gs.leader)
		if e != nil {
			log.Printf("Leader %v not online (DialHTTP), initialising election.\n", gs.leader)
			gs.elect()
		} else {
			remote.Close()
		}
	}
}
