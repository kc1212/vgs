package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

// the grid scheduler
type GS struct {
	id       int
	addr     string
	others   []string // other GS's
	clusters []string
	leader   string // the lead GS
	jobs     []Job
}

type GSArgs struct {
	SenderId int
	MsgType  int // TODO
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
	var clusters []string // TODO see above
	leader := ""
	jobs := InitJobs(0) // start with zero jobs
	return GS{id, addr, otherGS, clusters, leader, jobs}
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

// send messages to procs with higher id
func (gs *GS) Elect() {
	// TODO maybe keep tcp connection open?
	for i := range gs.others {
		remote, e := rpc.DialHTTP("tcp", gs.others[i])
		if e != nil {
			log.Fatal("Dialing failed: ", e)
		}
		args := GSArgs{gs.id, 0}
		var reply int
		if e := remote.Call("GS.RecvMsg", args, &reply); e != nil {
			log.Fatal("GS.RecvMsg failed: ", e)
		}
	}
}

func (gs *GS) RecvMsg(args *GSArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
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
