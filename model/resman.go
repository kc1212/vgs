package model

import "fmt"

type ResMan struct {
	nodes []Node // the cluster
	jobs  []Job
}

func InitResMan(n int) ResMan {
	nodes := make([]Node, n)
	for i, _ := range nodes {
		nodes[i].id = i
	}

	rm := ResMan{nodes, *new([]Job)}
	return rm
}

func (rm *ResMan) schedule() {
}

// always assign the very next job in queue?
// the node should already be free before calling this function
func (rm *ResMan) assign(n Node) {
	if n.running {
		panic(fmt.Sprintf("Node %v is already running!\n", n.id))
	}
	j := rm.jobs[0]
	n.startJob(j)
	rm.jobs = rm.jobs[1:] // remove the very first job
}

func (rm ResMan) PrintStatus() {
	fmt.Printf("ResMan: %v jobs and %v nodes\n", len(rm.jobs), len(rm.nodes))
}
