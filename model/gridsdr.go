package model

import (
	"log"
	"net/rpc"
	"time"
)

import (
	"github.com/kc1212/vgs/common"
	"github.com/kc1212/vgs/discosrv"
)

// GridSdr describes the properties of one grid scheduler
type GridSdr struct {
	common.Node
	gsNodes       *common.SyncedSet // other grid schedulers, not including myself
	rmNodes       *common.SyncedSet // the resource managers
	leader        string            // the lead grid scheduler
	incomingJobs  chan Job          // when user adds a job, it comes here
	scheduledJobs chan Job          // when GS schedules a job, it gets stored here
	tasks         chan common.Task  // these tasks require critical section (CS)
	inElection    *common.SyncedVal
	mutexRespChan chan int
	mutexReqChan  chan common.Task
	mutexState    *common.SyncedVal
	clock         *common.SyncedVal
	reqClock      int64
	discosrvAddr  string
}

// RPCArgs is the arguments for RPC calls between grid schedulers and/or resource maanagers
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
		make(chan Job, 1000),
		make(chan Job, 1000),
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

	gs.scheduleJobs()
}

func (gs *GridSdr) imLeader() bool {
	return gs.leader == gs.Addr
}

func (gs *GridSdr) scheduleJobs() {
	for {
		if len(gs.incomingJobs) == 0 || !gs.imLeader() {
			time.Sleep(time.Millisecond)
			continue
		}
		capacities := gs.getRMCapacities()
		for k, v := range capacities {
			jobs := takeJobs(int(v), gs.incomingJobs)
			gs.addJobsTask(jobs, k)
			if len(gs.incomingJobs) == 0 {
				break
			}
		}
	}
}

func (gs *GridSdr) addJobsTask(jobs []Job, rmAddr string) {
	gs.tasks <- func() (interface{}, error) {
		// send the job to RM
		reply, e := rpcAddJobsToRM(rmAddr, &jobs)

		// add jobs to the submitted list for all GSs
		for k := range gs.gsNodes.GetAll() {
			rpcAddScheduledJobsToGS(k, &jobs)
		}
		jobsToChan(jobs, gs.scheduledJobs) // do it for myself too

		// remove jobs from the incomingJobs list
		for k := range gs.gsNodes.GetAll() {
			rpcDropJobsInGS(k, len(jobs))
		}
		dropJobs(len(jobs), gs.incomingJobs) // do it for myself too

		return reply, e
	}
}

// rpcArgsForGS sets default values for GS
func (gs *GridSdr) rpcArgsForGS(msgType common.MsgType) RPCArgs {
	return RPCArgs{gs.ID, gs.Addr, msgType, gs.clock.Geti64()}
}

func (gs *GridSdr) getRMCapacities() map[string]int64 {
	capacities := make(map[string]int64)
	args := gs.rpcArgsForGS(common.GetCapacityMsg)
	for k := range gs.rmNodes.GetAll() {
		x, e := rpcSendMsgToRM(k, &args)
		if e == nil {
			capacities[k] = int64(x)
		}
	}
	return capacities
}

func (gs *GridSdr) notifyAndPopulateGSs(nodes []string) {
	args := gs.rpcArgsForGS(common.GSUpMsg)
	for _, node := range nodes {
		id, e := rpcSendMsgToGS(node, &args)
		if e == nil {
			gs.gsNodes.SetInt(node, int64(id))
		}
	}
}

func (gs *GridSdr) notifyAndPopulateRMs(nodes []string) {
	args := gs.rpcArgsForGS(common.RMUpMsg)
	for _, node := range nodes {
		id, e := rpcSendMsgToRM(node, &args)
		if e == nil {
			gs.rmNodes.SetInt(node, int64(id))
		}
	}
}

// TODO we need to generalise those sendMsg/addJobs functions

func rpcSendMsgToRM(addr string, args *RPCArgs) (int, error) {
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

// rpcAddJobsToRM creates an RPC connection with a ResMan and does one remote call on AddJob.
func rpcAddJobsToRM(addr string, args *[]Job) (int, error) {
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
func rpcSendMsgToGS(addr string, args *RPCArgs) (int, error) {
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

// rpcAddJobsToGS is a remote call that calls `RecvJobs`.
// NOTE: this function should only be executed when CS is obtained.
func rpcAddJobsToGS(addr string, jobs *[]Job) (int, error) {
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

func rpcAddScheduledJobsToGS(addr string, jobs *[]Job) (int, error) {
	log.Printf("Sending scheduled jobs %v, to %v\n", *jobs, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvScheduledJobs", jobs, &reply)
	return reply, remote.Close()
}

func rpcDropJobsInGS(addr string, n int) (int, error) {
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.DropJobs", &n, &reply)
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

	if len(gs.mutexRespChan) != 0 {
		log.Panic("Nodes following the protocol shouldn't send more messages")
	}

	gs.clock.Tick()
	successes := 0
	for k := range gs.gsNodes.GetAll() {
		args := gs.rpcArgsForGS(common.MutexReq)
		_, e := rpcSendMsgToGS(k, &args)
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

	// now empty gs.mutexRespChan because we received all the messages
	common.EmptyIntChan(gs.mutexRespChan)

	// here we're in critical section
	gs.mutexState.Set(common.StateHeld)
	log.Println("In CS!", gs.ID)
}

// releaseCritSection sets the mutexState to StateReleased and then runs all the queued requests.
func (gs *GridSdr) releaseCritSection() {
	gs.mutexState.Set(common.StateReleased)
loop:
	for {
		select {
		case resp := <-gs.mutexReqChan:
			_, e := resp()
			if e != nil {
				log.Panic("task failed with", e)
			}
		default:
			break loop
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
		args := gs.rpcArgsForGS(common.ElectionMsg)
		_, e := rpcSendMsgToGS(k, &args)
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
			args := gs.rpcArgsForGS(common.CoordinateMsg)
			rpcSendMsgToGS(k, &args) // NOTE: ok to fail the send, because nodes might be done
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
	log.Printf("new incoming jobs %v\n", *jobs)
	jobsToChan(*jobs, gs.incomingJobs)
	*reply = 0
	return nil
}

// NOTE: this function should not be called directly by the client, it requires CS.
func (gs *GridSdr) RecvScheduledJobs(jobs *[]Job, reply *int) error {
	log.Printf("adding scheduled jobs: %v\n", *jobs)
	jobsToChan(*jobs, gs.scheduledJobs)
	*reply = 0
	return nil
}

func (gs *GridSdr) DropJobs(n *int, reply *int) error {
	dropJobs(*n, gs.incomingJobs)
	*reply = 0
	return nil
}

// AddJobs is called by the client to add job(s) to the tasks queue.
// NOTE: this does not add jobs to the jobs queue, that is done by `runTasks`.
func (gs *GridSdr) AddJobs(jobs *[]Job, reply *int) error {
	gs.tasks <- func() (interface{}, error) {
		// add jobs to the others
		for k := range gs.gsNodes.GetAll() {
			rpcAddJobsToGS(k, jobs) // ok to fail
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
		// check whether there are tasks that needs running every 100ms
		time.Sleep(100 * time.Millisecond)
		if len(gs.tasks) > 0 {
			// acquire CS, run the tasks, run for 1ms at most, then release CS
			gs.obtainCritSection()
			timeout := time.After(time.Millisecond)
		inner_loop:
			for {
				select {
				case task := <-gs.tasks:
					_, e := task()
					if e != nil {
						log.Panic("task failed with", e)
					}
				case <-timeout:
					break inner_loop
				default:
					break inner_loop
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
		// NOTE: use gs.reqClock instead of the normal clock
		rpcSendMsgToGS(args.Addr, &RPCArgs{gs.ID, gs.Addr, common.MutexResp, gs.reqClock})
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
