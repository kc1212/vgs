package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

import "github.com/kc1212/vgs/common"

type ResMan struct {
	common.Node
	nodes []Worker
	jobs  []Job
}

type ResManArgs struct {
	Test int // TODO change this an actual job
}

func RunResMan(n int, port string) {
	nodes := make([]Worker, n)
	// TODO make proper Node
	rm := ResMan{common.Node{}, nodes, *new([]Job)}
	rm.startRPC(port)
	rm.startMainLoop()
}

// the RPC function
func (rm *ResMan) AddJob(args *ResManArgs, reply *int) error {
	log.Printf("Message received %v\n", *args)
	// rm.jobs = append(rm.jobs, args.JobArg)
	*reply = 1
	return nil
}

func (rm *ResMan) startRPC(port string) {
	// initialise RPC
	log.Printf("Initialising RPC on port %v\n", port)
	rpc.Register(rm)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", port)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (rm *ResMan) startMainLoop() {
	log.Printf("Startin main loop\n")
	for {
	}
	// TODO receive messages from grid scheduler
	// TODO update status of nodes
	// TODO schedule jobs
}

func (rm *ResMan) nextFreeNode() int {
	// TODO looping over all nodes is inefficient
	// because the low idx nodes are always assigned first
	for i := range rm.nodes {
		if !rm.nodes[i].running {
			return i
		}
	}
	return -1
}

// greedy scheduler
func (rm *ResMan) schedule() {
	x := rm.nextFreeNode()
	if x > -1 {
		// assigning the job will also remove it from the job queue
		rm.assign(x)
	}
}

// always assign the very next job in queue?
// the node should already be free before calling this function
func (rm *ResMan) assign(idx int) {
	j := rm.jobs[0]
	rm.nodes[idx].startJob(j)
	rm.jobs = rm.jobs[1:] // remove the very first job
}

func (rm ResMan) logStatus() {
	log.Printf("ResMan: %v jobs and %v nodes\n", len(rm.jobs), len(rm.nodes))
}
