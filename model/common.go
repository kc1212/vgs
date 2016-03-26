package model

import (
	"log"
	"strconv"
	"strings"
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

func (ef *SyncedVal) set(x interface{}) {
	defer ef.Unlock()
	ef.Lock()
	ef.val = x
}

func (ef *SyncedVal) get() interface{} {
	defer ef.RUnlock()
	ef.RLock()
	return ef.val
}

func idFromAddr(addr string, basePort int) int {
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		log.Panic("idFromAddr failed", e)
	}
	return port - basePort
}
