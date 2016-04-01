package model

import (
	"log"
	"time"
)

import "github.com/kc1212/vgs/common"
import "github.com/kc1212/vgs/discosrv"

type ResMan struct {
	common.Node
	workers      []Worker
	gsNodes      *common.SyncedSet
	jobs         []Job
	discosrvAddr string
}

func InitResMan(n int, id int, addr string, dsAddr string) ResMan {
	return ResMan{
		common.Node{ID: id, Addr: addr, Type: common.RMNode},
		make([]Worker, n),
		&common.SyncedSet{S: make(map[string]common.IntClient)},
		*new([]Job),
		dsAddr}
}

// RunResMan is the main function, it starts all its services.
func (rm *ResMan) Run() {
	reply, e := discosrv.ImAliveProbe(rm.Addr, common.RMNode, rm.discosrvAddr)
	if e != nil {
		log.Panicf("Discosrv on %v not online\n", rm.discosrvAddr)
	}
	rm.notifyAndPopulateGSs(reply.GSs)

	go discosrv.ImAlivePoll(rm.Addr, common.RMNode, rm.discosrvAddr)
	go common.RunRPC(rm, rm.Addr)
	rm.startMainLoop()
}

// AddJob RPC call
func (rm *ResMan) AddJob(args *[]Job, reply *int) error {
	log.Printf("Message received %v\n", *args)
	// rm.jobs = append(rm.jobs, args.JobArg)
	*reply = 1
	return nil
}

// RecvMsg PRC call
func (rm *ResMan) RecvMsg(args *RPCArgs, reply *int) error {
	log.Printf("Msg received %v\n", *args)
	*reply = -1
	if args.Type == common.RMUpMsg {
		*reply = rm.ID
		rm.gsNodes.SetInt(args.Addr, int64(args.ID))

	} else {
		log.Panic("Invalid message!", args)
	}
	return nil
}

func (rm *ResMan) notifyAndPopulateGSs(nodes []string) {
	// NOTE: does RM doesn't use a clock, hence the zero
	arg := RPCArgs{rm.ID, rm.Addr, common.RMUpMsg, 0}
	for _, node := range nodes {
		id, e := sendMsgToGS(node, &arg)
		if e == nil {
			rm.gsNodes.SetInt(node, int64(id))
		}
	}
}

func (rm *ResMan) startMainLoop() {
	log.Printf("Startin main loop\n")
	for {
		time.Sleep(time.Second)
		// TODO receive messages from grid scheduler
		// TODO update status of workers
		// TODO schedule jobs
	}
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
