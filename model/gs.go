package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"time"
)

// the grid scheduler
type GS struct {
	id              int
	addr            string
	basePort        int
	others          []string // other GS's
	clusters        []string
	leader          string // the lead GS
	jobs            []Job
	runningElection bool
	electionRWLock  sync.RWMutex
}

type MsgType int

const (
	ElectionMsg MsgType = iota
	CoordinateMsg
)

type GSArgs struct {
	SenderId   int
	SenderAddr string
	SenderType MsgType
}

func InitGS(id int, n int, basePort int, prefix string) GS {
	addr := prefix + strconv.Itoa(basePort+id)
	// TODO read from config file or have bootstrapper
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
	return GS{id, addr, basePort, otherGS, clusters, leader, jobs, false, sync.RWMutex{}}
}

// TODO how should the user submit request
// via REST API or RPC call from a client?

// TODO only one cluster at the moment, need to generalise
// runs the main loop
func (gs *GS) Run() {
	// initialise GS as RPC server
	gs.startRPC()

	// initialise RPC for calling cluster prodecures
	// remote, err := rpc.DialHTTP("tcp", "localhost:1234")
	// if err != nil {
	// 	log.Fatal("dialing:", err)
	// }

	// args := ResManArgs{42}
	// var reply int
	// err = remote.Call("ResMan.AddJob", args, &reply)
	// if err != nil {
	// 	log.Fatal("rm error:", err)
	// }

	// // initialise RPC for receiving requests
	// for {
	// 	// TODO get all the clusters
	// 	// TODO arrange them in loaded order
	// 	// TODO allocate *all* jobs
	// }
}

func (gs *GS) sendMsgToGS(otherId int, args GSArgs) error {
	log.Printf("Sending message to %v\n", gs.others[otherId])
	remote, e := rpc.DialHTTP("tcp", gs.others[otherId])
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", gs.others[otherId])
		return e
	}
	var reply int
	if e := remote.Call("GS.RecvMsg", args, &reply); e != nil {
		log.Printf("Node %v not online (RecvMsg)\n", gs.others[otherId])
		return e
	}
	return nil
}

// send messages to procs with higher id
func (gs *GS) Elect() {
	defer func() {
		gs.runningElection = false
		gs.electionRWLock.Unlock()
	}()
	gs.electionRWLock.Lock()
	gs.runningElection = true
	oks := 0
	for i := range gs.others {
		if idFromAddr(gs.others[i], gs.basePort) < gs.id {
			continue // do nothing to lower ids
		}
		e := gs.sendMsgToGS(i, GSArgs{gs.id, gs.addr, ElectionMsg})
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
			e := gs.sendMsgToGS(i, args)
			if e != nil {
				// ok to fail the send, because nodes might be done
				continue
			}
		}
	}

	time.Sleep(time.Second * 3)
}

func idFromAddr(addr string, basePort int) int {
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		log.Fatal("idFromAddr failed", e)
	}
	return port - basePort
}

func (gs *GS) RecvMsg(args *GSArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
	if args.SenderType == CoordinateMsg {
		gs.leader = args.SenderAddr
		log.Printf("Leader set to %v\n", gs.leader)
	} else if args.SenderType == ElectionMsg {
		// don't start a new election if one is already running
		gs.electionRWLock.RLock()
		if !gs.runningElection {
			go gs.Elect()
		}
		gs.electionRWLock.RUnlock()
	}
	*reply = 1
	return nil
}

func (gs *GS) startRPC() {
	// initialise RPC
	log.Printf("Initialising RPC on addr %v\n", gs.addr)
	rpc.Register(gs)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", gs.addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}
