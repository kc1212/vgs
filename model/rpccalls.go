package model

import (
	"log"
	"net/rpc"
)

import "github.com/kc1212/virtual-grid/common"

// TODO a lot of repeated code here, need to be generalised

func rpcSendMsgToRM(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending message %v to %v\n", *args, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "ResMan.RecvMsg", args, &reply)
	return reply, remote.Close()
}

// rpcAddJobsToRM creates an RPC connection with a ResMan and does one remote call on AddJob.
func rpcAddJobsToRM(addr string, args *[]Job) (int, error) {
	log.Printf("Sending job to RM on %v\n", addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "ResMan.AddJob", args, &reply)
	return reply, remote.Close()
}

// sendMsgToGS creates an RPC connection with another GridSdr and does one remote call on RecvMsg.
func rpcSendMsgToGS(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending message %v to GS on %v\n", *args, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvMsg", args, &reply)
	return reply, remote.Close()
}

// rpcAddJobsToGS is a remote call that calls `RecvJobs`.
// NOTE: this function should only be executed when CS is obtained.
func rpcSyncJobs(addr string, jobs *[]Job) (int, error) {
	log.Printf("Syncing jobs %v with GS on %v\n", *jobs, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvJobs", jobs, &reply)
	return reply, remote.Close()
}

func rpcSyncScheduledJobs(addr string, jobs *[]Job) (int, error) {
	log.Printf("Syncing scheduled jobs %v with GS %v\n", *jobs, addr)
	reply := -1
	remote, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e
	}
	common.RemoteCallNoFail(remote, "GridSdr.RecvScheduledJobs", jobs, &reply)
	return reply, remote.Close()
}

func rpcDropJobs(addr string, n int) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.DropJobs", &n)
	return reply, e
}

func rpcSyncCompletedJobs(addr string, js *[]int64) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.SyncCompletedJobs", js)
	return reply, e
}

func rpcRemoveCompletedJobs(addr string, js *[]int64) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.RemoveCompletedJobs", js)
	return reply, e
}
