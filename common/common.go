package common

//go:generate stringer -type=MsgType
//go:generate stringer -type=MutexState

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
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
	ID   int
	Addr string
	Type NodeType
}

type NodeType int

const (
	GSNode NodeType = iota
	RMNode
	DSNode
)

type Task func() (interface{}, error)

type SyncedVal struct {
	sync.RWMutex
	V interface{}
}

func (v *SyncedVal) Set(x interface{}) {
	defer v.Unlock()
	v.Lock()
	v.V = x
}

func (v *SyncedVal) Get() interface{} {
	defer v.RUnlock()
	v.RLock()
	return v.V
}

func (v *SyncedVal) Geti64() int64 {
	return v.Get().(int64)
}

func (t *SyncedVal) Tick() {
	x := t.Get().(int64) + 1
	t.Set(x)
}

func Max64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

type SyncedSet struct {
	sync.RWMutex
	S map[string]int64
}

func (s *SyncedSet) Set(k string, v int64) {
	s.Lock()
	defer s.Unlock()
	s.S[k] = v
}

func (s *SyncedSet) Delete(k string) bool {
	s.Lock()
	defer s.Unlock()
	_, found := s.S[k]
	if !found {
		return false
	}
	delete(s.S, k)
	return true
}

func (s *SyncedSet) Get(k string) (v int64, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok = s.S[k]
	return
}

func (s *SyncedSet) GetAll() map[string]int64 {
	s.RLock()
	defer s.RUnlock()
	return s.S
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

func SliceFromMap(mymap map[string]int64) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}
