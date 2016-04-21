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

// MsgType includes all messages types other than job list manipulation
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

// MutexState are all the possible states for Ricart-Agrawala algorithm
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

// NodeType can be either for GS, RM or DS (discosrv)
type NodeType int

const (
	GSNode NodeType = iota
	RMNode
	DSNode
)

// Task for running in CS or by a worker node
type Task func() (interface{}, error)

// SyncedVal is an interface{} with RWMutex
type SyncedVal struct {
	sync.RWMutex
	V interface{}
}

// Set sets the value
func (v *SyncedVal) Set(x interface{}) {
	defer v.Unlock()
	v.Lock()
	v.V = x
}

// Get gets the value
func (v *SyncedVal) Get() interface{} {
	defer v.RUnlock()
	v.RLock()
	return v.V
}

// Geti64 is Get for int64
func (v *SyncedVal) Geti64() int64 {
	return v.Get().(int64)
}

// Tick increments int64 value
func (v *SyncedVal) Tick() {
	x := v.Get().(int64) + 1
	v.Set(x)
}

// MaxInt64 returns the maximum of a and b
func MaxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the minimum of a and b
func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// IntClient stores an int64 ID and a Client pointer
type IntClient struct {
	Client *rpc.Client
	ID     int64
}

// SyncedSet is a concurrent map
type SyncedSet struct {
	sync.RWMutex
	S map[string]IntClient
}

// Set sets the value IntClient on key k
func (s *SyncedSet) Set(k string, v IntClient) {
	s.Lock()
	defer s.Unlock()
	s.S[k] = v
}

// SetInt is Set but only sets the int64 part of IntClient
func (s *SyncedSet) SetInt(k string, v int64) {
	s.Lock()
	defer s.Unlock()
	s.S[k] = IntClient{s.S[k].Client, v}
}

// Delete deletes entry at key k
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

// Get gets value at key k
func (s *SyncedSet) Get(k string) (v IntClient, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok = s.S[k]
	return
}

// GetInt gets the int part of the value at key k
func (s *SyncedSet) GetInt(k string) (int64, bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.S[k]
	return v.ID, ok
}

// GetAll does unwrapping
func (s *SyncedSet) GetAll() map[string]IntClient {
	s.RLock()
	defer s.RUnlock()
	return s.S
}

// RunRPC registers and runs the RPC server.
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

// RemoteCallNoFail does a remote.Call and logs failure
func RemoteCallNoFail(remote *rpc.Client, fn string, args interface{}, reply interface{}) error {
	e := remote.Call(fn, args, reply)
	if e != nil {
		log.Printf("Remote call %v failed, %v\n", fn, e.Error())
	}
	return e
}

// DialAndCallNoFail dial and then do the remote call
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

// SliceFromMap does what it says
func SliceFromMap(mymap map[string]IntClient) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}

// SliceToMap is like the opposite of SliceFromMap
func SliceToMap(ss []string) map[string]IntClient {
	m := make(map[string]IntClient)
	for _, s := range ss {
		m[s] = IntClient{}
	}
	return m
}

// EmptyIntChan drops everything in channel c
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
	var res []int64
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
