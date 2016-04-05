package model

import (
	"log"
	"net/rpc"
	"time"
)

import (
	"github.com/kc1212/virtual-grid/common"
	"github.com/kc1212/virtual-grid/discosrv"
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

// Run is the main function for GridSdr, it starts all its services, do not run it more than once.
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
		time.Sleep(100 * time.Millisecond)

		// try again later if I'm not leader
		if !gs.imLeader() {
			time.Sleep(time.Second)
			continue
		}

		// try again later if there are no jobs
		if len(gs.incomingJobs) == 0 {
			continue
		}

		// try again later if no free RMs
		addr, cap := gs.getNextFreeRM()
		if cap == -1 || addr == "" {
			continue
		}

		// schedule jobs if there are any
		select {
		case job := <-gs.incomingJobs:
			rest := takeJobs(cap-1, gs.incomingJobs)
			jobs := append(rest, job) // TODO this really should be prepend
			jobsToChan(jobs, gs.scheduledJobs)
			gs.runJobsTask(jobs, addr) // this function blocks util the task finishes executing
		case <-time.After(100 * time.Millisecond):
			break
		}
	}
}

// runJobsTask pushes to tasks channel, it blocks until the task is completed
func (gs *GridSdr) runJobsTask(jobs []Job, rmAddr string) {
	c := make(chan int)
	gs.tasks <- func() (interface{}, error) {
		// send the job to RM
		reply, e := rpcAddJobsToRM(rmAddr, &jobs)

		// add jobs to the submitted list for all GSs
		rpcJobsGo(common.SliceFromMap(gs.gsNodes.GetAll()), &jobs, rpcSyncScheduledJobs)

		// remove jobs from the incomingJobs list
		rpcIntGo(common.SliceFromMap(gs.gsNodes.GetAll()), len(jobs), rpcDropJobs)

		c <- 0
		return reply, e
	}
	<-c
}

// rpcArgsForGS sets default values for GS
func (gs *GridSdr) rpcArgsForGS(msgType common.MsgType) RPCArgs {
	return RPCArgs{gs.ID, gs.Addr, msgType, gs.clock.Geti64()}
}

// NOTE: there are various ways to improve this function, i.e. get the the RM with highest number of free workers
func (gs *GridSdr) getNextFreeRM() (string, int) {
	caps := gs.getRMCapacities()
	for k, v := range caps {
		if v > 0 {
			return k, int(v)
		}
	}
	return "", -1
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

// obtainCritSection implements most of the Ricart-Agrawala algorithm, it sends the critical section request and then wait for responses until some timeout.
// Initially we set the mutexState to StateWanted, if the critical section is obtained we set it to StateHeld.
// NOTE: this function isn't designed to be thread safe, it is run periodically in `runTasks`.
func (gs *GridSdr) obtainCritSection() {
	if gs.mutexState.Get().(common.MutexState) != common.StateReleased {
		log.Panicf("Should not be in CS, state: %v\n", gs)
	}

	if len(gs.mutexRespChan) != 0 {
		log.Panic("Nodes following the protocol shouldn't send more messages")
	}

	gs.mutexState.Set(common.StateWanted)

	gs.clock.Tick()
	args := gs.rpcArgsForGS(common.MutexReq)
	addrs := common.SliceFromMap(gs.gsNodes.GetAll())
	successes := rpcGo(addrs, &args, rpcSendMsgToGS)
	gs.reqClock = gs.clock.Geti64()

	// wait until others has written to mutexRespChan or time out (2s)
	cnt := 0
	timeout := time.After(2 * time.Second)
loop:
	for {
		if cnt >= successes {
			break
		}
		select {
		case <-gs.mutexRespChan:
			cnt++
		case <-timeout:
			break loop
		}
	}

	// now empty gs.mutexRespChan because we received all the messages
	// NOTE: probably not necessary
	common.EmptyIntChan(gs.mutexRespChan)

	// here we're in critical section
	gs.mutexState.Set(common.StateHeld)
	log.Println("In CS!", gs.ID)
}

// releaseCritSection sets the mutexState to StateReleased and then runs all the queued requests.
func (gs *GridSdr) releaseCritSection() {
	log.Printf("starting release")
	gs.mutexState.Set(common.StateReleased)
	log.Printf("set released", len(gs.mutexReqChan))
	for {
		select {
		case req := <-gs.mutexReqChan:
			_, e := req()
			if e != nil {
				log.Panic("request failed with", e)
			}
		default:
			log.Println("Out CS!", gs.ID)
			return
		}
	}
}

// elect implements the Bully algorithm.
func (gs *GridSdr) elect() {
	defer func() {
		gs.inElection.Set(false)
	}()
	gs.inElection.Set(true)

	gs.clock.Tick()
	oks := 0
	args := gs.rpcArgsForGS(common.ElectionMsg)
	for k, v := range gs.gsNodes.GetAll() {
		if v.ID < int64(gs.ID) {
			continue // do nothing to lower ids
		}
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

		args := gs.rpcArgsForGS(common.CoordinateMsg)
		addrs := common.SliceFromMap(gs.gsNodes.GetAll())
		rpcGo(addrs, &args, rpcSendMsgToGS) // NOTE: ok to fail the send, because nodes might be done
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
	log.Printf("%v new incoming jobs.\n", len(*jobs))
	jobsToChan(*jobs, gs.incomingJobs)
	*reply = 0
	return nil
}

// NOTE: this function should not be called directly by the client, it requires CS.
func (gs *GridSdr) RecvScheduledJobs(jobs *[]Job, reply *int) error {
	log.Printf("Adding %v scheduled jobs.\n", len(*jobs))
	jobsToChan(*jobs, gs.scheduledJobs)
	*reply = 0
	return nil
}

func (gs *GridSdr) DropJobs(n *int, reply *int) error {
	dropJobs(*n, gs.incomingJobs)
	*reply = 0
	return nil
}

// SyncCompletedJobs is called by the RM when job(s) are completed.
// We acquire a critical section and propogate the change to everybody.
func (gs *GridSdr) SyncCompletedJobs(js *[]int64, reply *int) error {
	c := make(chan int)
	gs.tasks <- func() (interface{}, error) {
		rpcInt64sGo(common.SliceFromMap(gs.gsNodes.GetAll()), js, rpcRemoveCompletedJobs)

		// TODO remove it from myself too
		log.Printf("I'm removing %v jobs.\n", len(*js))

		c <- 0
		return 0, nil
	}
	<-c
	*reply = 0
	return nil
}

// RemoveCompletedJobs is called by another GS to remove job(s) from the scheduledJobs
func (gs *GridSdr) RemoveCompletedJobs(js *[]int, reply *int) error {
	log.Printf("I'm removing %v jobs.\n", len(*js))
	*reply = 0
	return nil
}

// AddJobsTask is called by the client to add job(s) to the tasks queue, it returns when the job is synchronised.
func (gs *GridSdr) AddJobsTask(jobs *[]Job, reply *int) error {
	c := make(chan int)
	gs.tasks <- func() (interface{}, error) {
		// add jobs to the others
		rpcJobsGo(common.SliceFromMap(gs.gsNodes.GetAll()), jobs, rpcSyncJobs)

		// add jobs to myself
		r := -1
		e := gs.RecvJobs(jobs, &r)

		c <- 0
		return r, e
	}
	<-c
	*reply = 0
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
		if gs.inElection.Get().(bool) || gs.imLeader() {
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
