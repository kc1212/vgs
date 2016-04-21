package discosrv

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"time"
)

import "github.com/kc1212/virtual-grid/common"

// Srv represents the discovery server
type Srv struct {
	gsSet *common.SyncedSet
	rmSet *common.SyncedSet
}

// Args is for RPC argument
type Args struct {
	Addr     string
	Type     common.NodeType
	NeedList bool
}

// Reply is for RPC responses
type Reply struct {
	GSs   []string
	RMs   []string
	Reply int
}

// Run runs the DiscoSrv
func (ds *Srv) Run(addr string) {
	ds.gsSet = &common.SyncedSet{S: make(map[string]common.IntClient)}
	ds.rmSet = &common.SyncedSet{S: make(map[string]common.IntClient)}
	go common.RunRPC(ds, addr)
	go ds.runRemoveDead()
	http.HandleFunc("/", ds.hello)
	http.ListenAndServe(":8333", nil)
}

// ImAlive RPC, called by GS or RM to update their status
func (ds *Srv) ImAlive(args *Args, reply *Reply) error {
	now := time.Now().Unix()
	reply.Reply = 0
	if args.Type == common.GSNode {
		ds.gsSet.SetInt(args.Addr, now)
	} else if args.Type == common.RMNode {
		ds.rmSet.SetInt(args.Addr, now)
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

// TODO some repeated code in "ImAliveProbe" and "ImAlivePoll"

// ImAliveProbe sends a probe message to discosrv, discosrv should return a list of RMs and GSs.
func ImAliveProbe(nodeAddr string, nodeType common.NodeType, dsAddr string) (Reply, error) {
	remote, e := rpc.DialHTTP("tcp", dsAddr)
	reply := Reply{}
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", dsAddr)
		return reply, e
	}
	defer remote.Close()

	args := Args{
		nodeAddr,
		nodeType,
		true}
	e = common.RemoteCallNoFail(remote, "Srv.ImAlive", &args, &reply)
	return reply, e
}

// ImAlivePoll polls the discosrv to inform it that the node on `nodeAddr` is online.
func ImAlivePoll(nodeAddr string, nodeType common.NodeType, dsAddr string) (Reply, error) {
	remote, e := rpc.DialHTTP("tcp", dsAddr)
	reply := Reply{}
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", dsAddr)
		return reply, e
	}
	defer remote.Close()

	args := Args{
		nodeAddr,
		nodeType,
		false}
	for {
		// TODO check whether discosrv is still online, otherwise redail
		common.RemoteCallNoFail(remote, "Srv.ImAlive", &args, &reply)
		time.Sleep(10 * time.Second)
	}
}

func (ds *Srv) runRemoveDead() {
	for {
		time.Sleep(time.Second)

		threshold := int64(20)
		t := time.Now().Unix()
		log.Printf("%v GSs, %v RMs\n", len(ds.gsSet.GetAll()), len(ds.rmSet.GetAll()))

		// TODO repeated code, loop over the two sets
		ds.gsSet.Lock()
		for k := range ds.gsSet.S {
			if t-ds.gsSet.S[k].ID > threshold {
				delete(ds.gsSet.S, k)
			}
		}
		ds.gsSet.Unlock()

		ds.rmSet.Lock()
		for k := range ds.rmSet.S {
			if t-ds.rmSet.S[k].ID > threshold {
				delete(ds.rmSet.S, k)
			}
		}
		ds.rmSet.Unlock()
	}
}

func (ds *Srv) hello(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v GSs, %v RMs\n", len(ds.gsSet.GetAll()), len(ds.rmSet.GetAll()))
	io.WriteString(w, msg)
}
