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

type Task func() (interface{}, error)

type SyncedFlag struct {
	sync.RWMutex
	isRunning bool
}

func (ef *SyncedFlag) set(f bool) {
	defer ef.Unlock()
	ef.Lock()
	ef.isRunning = f
}

func (ef *SyncedFlag) get() bool {
	defer ef.RUnlock()
	ef.RLock()
	return ef.isRunning
}

func idFromAddr(addr string, basePort int) int {
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		log.Panic("idFromAddr failed", e)
	}
	return port - basePort
}
