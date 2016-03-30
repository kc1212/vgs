package model

import (
	"errors"
	"log"
	"time"
)

type DiscoSrv struct {
	gsSet *SyncedSet
	rmSet *SyncedSet
}

type DiscoSrvArgs struct {
	Addr     string
	Type     NodeType
	NeedList bool
}

type DiscoSrvReply struct {
	GSs   []string
	RMs   []string
	Reply int
}

func (ds *DiscoSrv) ImAlive(args *DiscoSrvArgs, reply *DiscoSrvReply) error {
	now := time.Now().Unix()
	reply.Reply = 0
	if args.Type == GSNode {
		ds.gsSet.Set(args.Addr, now)
	} else if args.Type == RMNode {
		ds.rmSet.Set(args.Addr, now)
	} else {
		reply.Reply = 1
		return errors.New("Invalid NodeType!")
	}

	if args.NeedList {
		ds.gsSet.RLock()
		reply.GSs = sliceFromMap(ds.gsSet.set)
		ds.gsSet.RUnlock()

		ds.rmSet.RLock()
		reply.RMs = sliceFromMap(ds.rmSet.set)
		ds.rmSet.RUnlock()
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

func (ds *DiscoSrv) RunDiscoSrv(addr string) {
	ds.gsSet = &SyncedSet{set: make(map[string]int64)}
	ds.rmSet = &SyncedSet{set: make(map[string]int64)}
	go RunRPC(ds, addr)
	ds.removeDead()
}
