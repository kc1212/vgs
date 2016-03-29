package model

import (
	"log"
	"sync"
	"time"
)

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

type DiscoSrv struct {
	gsSet SyncedSet
	rmSet SyncedSet
}

type DiscoSrvArgs struct {
	Addr string
	Type NodeType
}

func (ds *DiscoSrv) ImAlive(args *DiscoSrvArgs, reply *int) error {
	now := time.Now().Unix()
	*reply = 0
	if args.Type == GSNode {
		ds.gsSet.Set(args.Addr, now)
	} else if args.Type == RMNode {
		ds.rmSet.Set(args.Addr, now)
	} else {
		*reply = 1
		log.Panic("Invalid NodeType!")
	}
	return nil
}

func (ds *DiscoSrv) removeDead() {
	for {
		time.Sleep(time.Second)
		threshold := int64(20)
		t := time.Now().Unix()
		log.Println("GS: ", ds.gsSet)
		log.Println("RM: ", ds.rmSet)

		// TODO repeated code, loop over the two sets
		ds.gsSet.Lock()
		for k := range ds.gsSet.set {
			if t-ds.gsSet.set[k] > threshold {
				delete(ds.gsSet.set, k)
			}
		}
		ds.gsSet.Unlock()

		ds.rmSet.Lock()
		for k := range ds.rmSet.set {
			if t-ds.rmSet.set[k] > threshold {
				delete(ds.rmSet.set, k)
			}
		}
		ds.rmSet.Unlock()
	}
}

func (ds *DiscoSrv) RunDiscoSrv() {
	go RunRPC(ds, "localhost:3333")
	ds.removeDead()
}
