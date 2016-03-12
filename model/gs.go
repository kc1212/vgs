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
type GS struct {
	id          int
	addr        string
	basePort    int
	others      []string // other GS's
	clusters    []string
	leader      string // the lead GS
	jobs        []Job
	runningElec ElectionFlag
}

type ElectionFlag struct {
	isRunning bool
	mutex     sync.RWMutex
}

type GSArgs struct {
	SenderId   int
	SenderAddr string
	SenderType MsgType
}

func (ef *ElectionFlag) set(f bool) {
	defer ef.mutex.Unlock()
	ef.mutex.Lock()
	ef.isRunning = f
}

func (ef *ElectionFlag) get() bool {
	defer ef.mutex.RUnlock()
	ef.mutex.RLock()
	return ef.isRunning
}

func InitGS(id int, n int, basePort int, prefix string) GS {
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
	return GS{id, addr, basePort, otherGS, clusters, leader, jobs, ElectionFlag{}}
}

// TODO how should the user submit request
// via REST API or RPC call from a client?

func (gs *GS) Run() {
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
	if e := remote.Call("GS.RecvMsg", args, &reply); e != nil {
		log.Printf("Node %v not online (RecvMsg)\n", addr)
	}
	return reply, remote.Close()
}

// send messages to procs with higher id
func (gs *GS) Elect() {
	defer func() {
		gs.runningElec.set(false)
	}()
	gs.runningElec.set(true)
	oks := 0
	for i := range gs.others {
		if idFromAddr(gs.others[i], gs.basePort) < gs.id {
			continue // do nothing to lower ids
		}
		_, e := sendMsgToGS(gs.others[i], GSArgs{gs.id, gs.addr, ElectionMsg})
		if e != nil {
			continue
		}
		oks++
	}

	// if no responses, then set the node itself as leader, and tell others
	gs.leader = gs.addr
	if oks == 0 {
		log.Printf("I'm the leader (%v).\n", gs.addr)
		for i := range gs.others {
			args := GSArgs{gs.id, gs.addr, CoordinateMsg}
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

func (gs *GS) RecvMsg(args *GSArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
	if args.SenderType == CoordinateMsg {
		gs.leader = args.SenderAddr
		log.Printf("Leader set to %v\n", gs.leader)
	} else if args.SenderType == ElectionMsg {
		// don't start a new election if one is already running
		if !gs.runningElec.get() {
			go gs.Elect()
		}
	}
	*reply = 1
	return nil
}

func (gs *GS) pollLeader() {
	for {
		time.Sleep(time.Second)
		if gs.addr == gs.leader {
			continue
		}
		remote, e := rpc.DialHTTP("tcp", gs.leader) // TODO should we have a mutex on `gs.leader?`
		if e != nil {
			log.Printf("Leader %v not online (DialHTTP), initialising election.\n", gs.leader)
			gs.Elect()
		} else {
			remote.Close()
		}
	}
}

func (gs *GS) startRPC() {
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
