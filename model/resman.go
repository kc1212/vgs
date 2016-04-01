package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

import "github.com/kc1212/vgs/common"
import "github.com/kc1212/vgs/discosrv"

type ResMan struct {
	common.Node
	workers []Worker
	gsnodes []string
	jobs    []Job
}

type ResManArgs struct {
	Test int // TODO change this an actual job
}

// RunResMan is the main function, it starts all its services.
func RunResMan(n int, id int, addr string, dsAddr string) {
	workers := make([]Worker, n)
	// TODO make proper Node
	reply, e := discosrv.ImAliveProbe(addr, common.RMNode, dsAddr)
	if e != nil {
		log.Panicf("Discosrv on %v not online\n", dsAddr)
	}
	rm := ResMan{common.Node{ID: id, Addr: addr, Type: common.RMNode}, workers, reply.GSs, *new([]Job)}

	go discosrv.ImAlivePoll(addr, common.RMNode, dsAddr)
	go common.RunRPC(rm, addr)
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
	// TODO update status of workers
	// TODO schedule jobs
}

func (rm *ResMan) nextFreeNode() int {
	// TODO looping over all workers is inefficient
	// because the low idx workers are always assigned first
	for i := range rm.workers {
		if !rm.workers[i].running {
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
	rm.workers[idx].startJob(j)
	rm.jobs = rm.jobs[1:] // remove the very first job
}

func (rm ResMan) logStatus() {
	log.Printf("ResMan: %v jobs and %v workers\n", len(rm.jobs), len(rm.workers))
}
