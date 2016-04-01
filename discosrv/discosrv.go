package discosrv

import (
	"errors"
	"log"
	"net/rpc"
	"time"
)

import "github.com/kc1212/vgs/common"

type DiscoSrv struct {
	gsSet *common.SyncedSet
	rmSet *common.SyncedSet
}

type DiscoSrvArgs struct {
	Addr     string
	Type     common.NodeType
	NeedList bool
}

type DiscoSrvReply struct {
	GSs   []string
	RMs   []string
	Reply int
}

func (ds *DiscoSrv) Run(addr string) {
	ds.gsSet = &common.SyncedSet{S: make(map[string]int64)}
	ds.rmSet = &common.SyncedSet{S: make(map[string]int64)}
	go common.RunRPC(ds, addr)
	ds.removeDead()
}

func (ds *DiscoSrv) ImAlive(args *DiscoSrvArgs, reply *DiscoSrvReply) error {
	now := time.Now().Unix()
	reply.Reply = 0
	if args.Type == common.GSNode {
		ds.gsSet.Set(args.Addr, now)
	} else if args.Type == common.RMNode {
		ds.rmSet.Set(args.Addr, now)
	} else {
		reply.Reply = 1
		return errors.New("Invalid NodeType!")
	}

	if args.NeedList {
		ds.gsSet.RLock()
		reply.GSs = common.SliceFromMap(ds.gsSet.S)
		ds.gsSet.RUnlock()

		ds.rmSet.RLock()
		reply.RMs = common.SliceFromMap(ds.rmSet.S)
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
		for k := range ds.gsSet.S {
			if t-ds.gsSet.S[k] > threshold {
				delete(ds.gsSet.S, k)
			}
		}
		ds.gsSet.Unlock()

		ds.rmSet.Lock()
		for k := range ds.rmSet.S {
			if t-ds.rmSet.S[k] > threshold {
				delete(ds.rmSet.S, k)
			}
		}
		ds.rmSet.Unlock()
	}
}

// TODO some repeated code in "ImAliveProbe" and "ImAlivePoll"

// ImAliveProbe sends Node's info to Discosrv
func ImAliveProbe(nodeAddr string, nodeType common.NodeType, dsAddr string) (DiscoSrvReply, error) {
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
	e = common.RemoteCallNoFail(remote, "DiscoSrv.ImAlive", &args, &reply)
	return reply, e
}

// ImAlivePoll sends which Nodes are online
func ImAlivePoll(nodeAddr string, nodeType common.NodeType, dsAddr string) (DiscoSrvReply, error) {
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
		common.RemoteCallNoFail(remote, "DiscoSrv.ImAlive", &args, &reply)
		time.Sleep(10 * time.Second)
	}
}
