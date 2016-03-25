package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"
)

// the grid scheduler
type GridSdr struct {
	id           int
	addr         string
	basePort     int
	others       []string // other GridSdr's
	clusters     []string
	leader       string // the lead GridSdr
	jobs         []Job
	electionFlag ElectionFlag
	syncChan     chan int
}

type ElectionFlag struct {
	sync.RWMutex
	isRunning bool
}

func (ef *ElectionFlag) set(f bool) {
	defer ef.Unlock()
	ef.Lock()
	ef.isRunning = f
}

func (ef *ElectionFlag) get() bool {
	defer ef.RUnlock()
	ef.RLock()
	return ef.isRunning
}

type GSArgs struct {
	Id    int
	Addr  string
	Type  MsgType
	Clock int64
}

func InitGridSdr(id int, n int, basePort int, prefix string) GridSdr {
	addr := prefix + strconv.Itoa(basePort+id)
	// TODO read from config file or have bootstrap/discovery server
	var otherGS []string
	for i := 0; i < n; i++ {
		if i != id {
			otherGS = append(otherGS, prefix+strconv.Itoa(basePort+i))
		}
	}
	// TODO see above
	var clusters []string
	leader := ""
	jobs := InitJobs(0) // start with zero jobs
	return GridSdr{id, addr, basePort, otherGS, clusters, leader, jobs,
		ElectionFlag{}, make(chan int, n-1)}
}

// TODO how should the user submit request
// via REST API or RPC call from a client?

func (gs *GridSdr) Run() {
	go gs.startRPC()
	go gs.pollLeader()

	for {
		// TODO get all the clusters
		// TODO arrange them in loaded order
		// TODO allocate *all* jobs
	}
}

func addSendJobToRM(addr string, args ResManArgs) (int, error) {
	log.Printf("Sending job to %v\n", addr)
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return -1, e
	}
	reply := -1
	err := remote.Call("ResMan.AddJob", args, &reply)
	if err != nil {
		log.Printf("Node %v not online (ResMan.AddJob)\n", addr)
	}
	return reply, remote.Close()
}

func sendMsgToGS(addr string, args GSArgs) (int, error) {
	log.Printf("Sending message to %v\n", addr)
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return -1, e
	}
	reply := -1
	if e := remote.Call("GridSdr.RecvMsg", args, &reply); e != nil {
		log.Printf("Node %v not online (RecvMsg)\n", addr)
	}
	return reply, remote.Close()
}

// send the critical section request and then wait for responses until some timeout
// don't wait for response for nodes that are already offline
func (gs *GridSdr) reqCritSection() {
	successes := 0
	for _, o := range gs.others {
		_, e := sendMsgToGS(o, GSArgs{gs.id, gs.addr, CritSectionMsg, 0}) // TODO add time
		if e == nil {
			successes++
		}
	}

	// wait until others has written to syncChan or time out
	t := time.Now().Add(time.Second)
	for t.After(time.Now()) {
		if len(gs.syncChan) >= successes {
			break
		}
		time.Sleep(time.Microsecond)
	}

	// empty the channel
	// NOTE: nodes following the protocol shouldn't send more messages
	for len(gs.syncChan) > 0 {
		<-gs.syncChan
	}

	// here we're in critical section
}

// send messages to procs with higher id
func (gs *GridSdr) elect() {
	defer func() {
		gs.electionFlag.set(false)
	}()
	gs.electionFlag.set(true)

	oks := 0
	for _, o := range gs.others {
		if idFromAddr(o, gs.basePort) < gs.id {
			continue // do nothing to lower ids
		}
		_, e := sendMsgToGS(o, GSArgs{gs.id, gs.addr, ElectionMsg, 0}) // TODO add time
		if e != nil {
			continue
		}
		oks++
	}

	// if no responses, then set the node itself as leader, and tell others
	gs.leader = gs.addr
	log.Printf("I'm the leader (%v).\n", gs.leader)
	if oks == 0 {
		for i := range gs.others {
			args := GSArgs{gs.id, gs.addr, CoordinateMsg, 0} // TODO add time
			_, e := sendMsgToGS(gs.others[i], args)
			if e != nil {
				// ok to fail the send, because nodes might be done
				continue
			}
		}
	}

	// artificially make the election last longer so that multiple messages
	// requests won't initialise multiple election runs
	time.Sleep(time.Second)
}

func (gs *GridSdr) RecvMsg(args *GSArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
	*reply = 1
	if args.Type == CoordinateMsg {
		gs.leader = args.Addr
		log.Printf("Leader set to %v\n", gs.leader)
	} else if args.Type == ElectionMsg {
		// don't start a new election if one is already running
		if !gs.electionFlag.get() {
			go gs.elect()
		}
	} else if args.Type == CritSectionMsg {
		// TODO start a process that would respond
		// TODO consider to send the response after a short delay
	}
	return nil
}

func (gs *GridSdr) pollLeader() {
	for {
		time.Sleep(time.Second)
		if gs.addr == gs.leader {
			continue
		}
		remote, e := rpc.DialHTTP("tcp", gs.leader) // TODO should we have a mutex on `gs.leader?`
		if e != nil {
			log.Printf("Leader %v not online (DialHTTP), initialising election.\n", gs.leader)
			gs.elect()
		} else {
			remote.Close()
		}
	}
}

func (gs *GridSdr) startRPC() {
	log.Printf("Initialising RPC on addr %v\n", gs.addr)
	rpc.Register(gs)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", gs.addr)
	if e != nil {
		log.Panic("startRPC failed", e)
	}
	// the Serve function runs until death
	http.Serve(l, nil)
}
