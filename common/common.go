package common

//go:generate stringer -type=MsgType
//go:generate stringer -type=MutexState
//go:generate stringer -type=JobStatus

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
	GSUpMsg
	RMUpMsg
	GetCapacityMsg
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

type JobStatus int

const (
	Waiting JobStatus = iota
	Submitted
	Running
	Finished
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

type IntClient struct {
	Client *rpc.Client
	ID     int64
}

type SyncedSet struct {
	sync.RWMutex
	S map[string]IntClient
}

func (s *SyncedSet) Set(k string, v IntClient) {
	s.Lock()
	defer s.Unlock()
	s.S[k] = v
}

func (s *SyncedSet) SetInt(k string, v int64) {
	s.Lock()
	defer s.Unlock()
	s.S[k] = IntClient{s.S[k].Client, v}
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

func (s *SyncedSet) Get(k string) (v IntClient, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok = s.S[k]
	return
}

func (s *SyncedSet) GetInt(k string) (int64, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.S[k]
	return v.ID, ok
}

func (s *SyncedSet) GetAll() map[string]IntClient {
	s.RLock()
	defer s.RUnlock()
	return s.S
}

func SliceToMap(ss []string) map[string]IntClient {
	m := make(map[string]IntClient)
	for _, s := range ss {
		m[s] = IntClient{}
	}
	return m
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

// TODO is there a way to make the reply generic?
func DialAndCallNoFail(addr string, fn string, args interface{}) (int, error) {
	// var reply interface{}
	reply := -1
	remote, e1 := rpc.DialHTTP("tcp", addr)
	if e1 != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
		return reply, e1
	}
	defer remote.Close()
	e2 := RemoteCallNoFail(remote, fn, args, &reply)
	return reply, e2
}

func SliceFromMap(mymap map[string]IntClient) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}

func EmptyIntChan(c <-chan int) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}

// TakeAllInt64Chan returns a list with all the values in the buffered channel
func TakeAllInt64Chan(c <-chan int64) []int64 {
	res := make([]int64, 0)
loop:
	for {
		select {
		case v := <-c:
			res = append(res, v)
		default:
			break loop
		}
	}
	return res
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
