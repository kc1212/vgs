package model

import (
<<<<<<< HEAD
=======
	"log"
	"strconv"
	"strings"
>>>>>>> mutex
	"sync"
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
<<<<<<< HEAD
=======
	}
	return b
}

func idFromAddr(addr string, basePort int) int {
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		log.Panic("idFromAddr failed", e)
>>>>>>> mutex
	}
	return b
}
