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
	GetIDMsg
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

type SyncedSet struct {
	sync.RWMutex
	set map[string]int64
}

func (s *SyncedSet) Set(k string, v int64) {
	s.Lock()
	defer s.Unlock()
	s.set[k] = v
}

func (s *SyncedSet) Delete(k string) bool {
	s.Lock()
	defer s.Unlock()
	_, found := s.set[k]
	if !found {
		return false
	}
	delete(s.set, k)
	return true
}

func (s *SyncedSet) Get(k string) (v int64, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok = s.set[k]
	return
}

func (s *SyncedSet) GetAll() map[string]int64 {
	s.RLock()
	defer s.RUnlock()
	return s.set
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

func RemoteCallNoFail(remote *rpc.Client, fn string, args interface{}, reply interface{}) error {
	e := remote.Call(fn, args, reply)
	if e != nil {
		log.Printf("Remote call %v on %v failed, %v\n", fn, args, e.Error())
	}
	return e
}

func sliceFromMap(mymap map[string]int64) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}

// TODO some repeated code in imAliveProbe and imAlivePoll
func imAliveProbe(nodeAddr string, nodeType NodeType, dsAddr string) (DiscoSrvReply, error) {
	remote, e := rpc.DialHTTP("tcp", dsAddr)
	reply := DiscoSrvReply{}
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", dsAddr)
		return reply, e
	}
	defer remote.Close()

	args := DiscoSrvArgs{
		nodeAddr,
		nodeType,
		true}
	e = RemoteCallNoFail(remote, "DiscoSrv.ImAlive", &args, &reply)
	return reply, e
}

func imAlivePoll(nodeAddr string, nodeType NodeType, dsAddr string) (DiscoSrvReply, error) {
	remote, e := rpc.DialHTTP("tcp", dsAddr)
	reply := DiscoSrvReply{}
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", dsAddr)
		return reply, e
	}
	defer remote.Close()

	args := DiscoSrvArgs{
		nodeAddr,
		nodeType,
		false}
	for {
		// TODO check whether discosrv is still online, otherwise redail
		RemoteCallNoFail(remote, "DiscoSrv.ImAlive", &args, &reply)
		time.Sleep(10 * time.Second)
	}
}
