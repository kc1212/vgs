package model

import (
	"log"
	"sync"
)

import "github.com/kc1212/virtual-grid/common"

func rpcSendMsgToRM(addr string, args *RPCArgs) (int, error) {
	// log.Printf("Sending message %v to %v\n", *args, addr)
	reply, e := common.DialAndCallNoFail(addr, "ResMan.RecvMsg", args)
	return reply, e
}

// rpcAddJobsToRM creates an RPC connection with a ResMan and does one remote call on AddJob.
func rpcAddJobsToRM(addr string, args *[]Job) (int, error) {
	log.Printf("Sending job to RM on %v\n", addr)
	reply, e := common.DialAndCallNoFail(addr, "ResMan.AddJob", args)
	return reply, e
}

// sendMsgToGS creates an RPC connection with another GridSdr and does one remote call on RecvMsg.
func rpcSendMsgToGS(addr string, args *RPCArgs) (int, error) {
	log.Printf("Sending message %v to GS on %v\n", *args, addr)
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.RecvMsg", args)
	return reply, e
}

// rpcAddJobsToGS is a remote call that calls `RecvJobs`.
// NOTE: this function should only be executed when CS is obtained.
func rpcSyncJobs(addr string, jobs *[]Job) (int, error) {
	log.Printf("Syncing %v jobs with GS on %v\n", len(*jobs), addr)
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.RecvJobs", jobs)
	return reply, e
}

func rpcSyncScheduledJobs(addr string, jobs *[]Job) (int, error) {
	log.Printf("Syncing %v scheduled jobs with GS on %v\n", len(*jobs), addr)
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.RecvScheduledJobs", jobs)
	return reply, e
}

func rpcDropJobs(addr string, n int) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.DropJobs", &n)
	return reply, e
}

func rpcSyncCompletedJobs(addr string, jobs *[]int64) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.SyncCompletedJobs", jobs)
	return reply, e
}

func rpcRemoveCompletedJobs(addr string, jobs *[]int64) (int, error) {
	reply, e := common.DialAndCallNoFail(addr, "GridSdr.RemoveCompletedJobs", jobs)
	return reply, e
}

// TODO horrible repeated code
func rpcInt64sGo(addrs []string, args *[]int64,
	rpcFn func(string, *[]int64) (int, error)) int {

	wg := sync.WaitGroup{}
	ch := make(chan int, len(addrs))
	for _, addr := range addrs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			_, e := rpcFn(s, args)
			if e == nil {
				ch <- 0
			}
		}(addr)
	}
	wg.Wait()

	close(ch)
	res := 0
	for range ch {
		res++
	}
	return res
}

func rpcIntGo(addrs []string, args int,
	rpcFn func(string, int) (int, error)) int {

	wg := sync.WaitGroup{}
	ch := make(chan int, len(addrs))
	for _, addr := range addrs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			_, e := rpcFn(s, args)
			if e == nil {
				ch <- 0
			}
		}(addr)
	}
	wg.Wait()

	close(ch)
	res := 0
	for range ch {
		res++
	}
	return res
}

func rpcJobsGo(addrs []string, args *[]Job,
	rpcFn func(string, *[]Job) (int, error)) int {

	wg := sync.WaitGroup{}
	ch := make(chan int, len(addrs))
	for _, addr := range addrs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			_, e := rpcFn(s, args)
			if e == nil {
				ch <- 0
			}
		}(addr)
	}
	wg.Wait()

	close(ch)
	res := 0
	for range ch {
		res++
	}
	return res
}

func rpcGo(addrs []string, args *RPCArgs,
	rpcFn func(string, *RPCArgs) (int, error)) int {

	wg := sync.WaitGroup{}
	ch := make(chan int, len(addrs))
	for _, addr := range addrs {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			_, e := rpcFn(s, args)
			if e == nil {
				ch <- 0
			}
		}(addr)
	}
	wg.Wait()

	close(ch)
	res := 0
	for range ch {
		res++
	}
	return res
}
