package model

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type MsgType int

const (
	ElectionMsg MsgType = iota
	CoordinateMsg
	MutexReq
	MutexResp
)

type MutexState int

const (
	StateReleased MutexState = iota
	StateWanted
	StateHeld
)

// Node is a generic node
type Node struct {
	id       int
	addr     string
	nodeType NodeType
}

type NodeType int

const (
	GSNode NodeType = iota
	RMNode
)

type Task func() (interface{}, error)

type SyncedVal struct {
	sync.RWMutex
	val interface{}
}

func (v *SyncedVal) set(x interface{}) {
	defer v.Unlock()
	v.Lock()
	v.val = x
}

func (v *SyncedVal) get() interface{} {
	defer v.RUnlock()
	v.RLock()
	return v.val
}

func (v *SyncedVal) geti64() int64 {
	return v.get().(int64)
}

func (t *SyncedVal) tick() {
	x := t.get().(int64) + 1
	t.set(x)
}

func max64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// runRPC registers and runs the RPC server.
func RunRPC(s interface{}, addr string) {
	log.Printf("Initialising RPC on addr %v\n", addr)
	rpc.Register(s)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Panic("runRPC failed", e)
	}
	// the Serve function runs until death
	http.Serve(l, nil)
}

func RemoteCallNoFail(remote *rpc.Client, fn string, args interface{}, reply *int) {
	if e := remote.Call(fn, args, &reply); e != nil {
		log.Printf("Remote call %v on %v failed, %v\n", fn, args, e.Error())
	}
}

func imAlive(args DiscoSrvArgs, discosrv string) (int, error) {
	remote, e := rpc.DialHTTP("tcp", discosrv)
	reply := -1
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", discosrv)
		return reply, e
	}
	defer remote.Close()

	for {
		time.Sleep(10 * time.Second)
		// TODO check whether discosrv is still online, otherwise redail
		RemoteCallNoFail(remote, "DiscoSrv.ImAlive", &args, &reply)
	}
}
