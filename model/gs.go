package model

import (
	"log"
	"math/rand"
	"net/rpc"
)

// the grid scheduler
type GS struct {
	id       int
	others   []string // other GS's
	clusters []string
	leader   string // the lead GS
	jobs     []Job
}

func StartGS() {
	id := rand.Int()
	others := *new([]string)   // TODO read from config file or have bootstrapper
	clusters := *new([]string) // TODO see above
	leader := ""
	jobs := InitJobs(0) // start with zero jobs
	gs := GS{id, others, clusters, leader, jobs}
	gs.run()
}

// TODO how should the user submit request
// via REST API or RPC call from a client?

// TODO only one cluster at the moment, need to generalise
// runs the main loop
func (gs *GS) run() {
	// initialise RPC for calling cluster prodecures
	remote, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	args := ResManArgs{42}
	var reply int
	err = remote.Call("ResMan.AddJob", args, &reply)
	if err != nil {
		log.Fatal("rm error:", err)
	}

	// initialise RPC for receiving requests
	for {
		// TODO get all the clusters
		// TODO arrange them in loaded order
		// TODO allocate *all* jobs
	}
}
