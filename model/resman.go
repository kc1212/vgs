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
	workers      []Worker
	gsnodes      []string
	jobs         []Job
	discosrvAddr string
}

type ResManArgs struct {
	Type common.MsgType
}

func InitResMan(n int, id int, addr string, dsAddr string) ResMan {
	reply, e := discosrv.ImAliveProbe(addr, common.RMNode, dsAddr)
	if e != nil {
		log.Panicf("Discosrv on %v not online\n", dsAddr)
	}
	return ResMan{
		common.Node{ID: id, Addr: addr, Type: common.RMNode},
		make([]Worker, n), reply.GSs, *new([]Job), dsAddr}
}

// RunResMan is the main function, it starts all its services.
func (rm *ResMan) Run() {
	go discosrv.ImAlivePoll(rm.Addr, common.RMNode, rm.discosrvAddr)
	go common.RunRPC(rm, rm.Addr)
	rm.startMainLoop()
}

// AddJob RPC call
func (rm *ResMan) AddJob(args *ResManArgs, reply *int) error {
	log.Printf("Message received %v\n", *args)
	// rm.jobs = append(rm.jobs, args.JobArg)
	*reply = 1
	return nil
}

// RecvMsg PRC call
func (rm *ResMan) RecvMsg(args *ResManArgs, reply *int) error {
	*reply = -1
	if args.Type == common.RMUpMsg {

	} else {
	}
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
