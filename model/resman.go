package model

import "fmt"

type ResMan struct {
	nodes []Node
	jobs  []Job
}

func InitResMan(n int) ResMan {
	nodes := InitNodes(n)
	rm := ResMan{nodes, *new([]Job)}
	return rm
}

func (rm *ResMan) run() {
	// TODO receive messages from GS
	// TODO update status of nodes
	// TODO schedule jobs
}

func (rm *ResMan) nextFreeNode() int {
	// TODO looping over all nodes is inefficient
	// because the low idx nodes are always assigned first
	for i := range rm.nodes {
		if !rm.nodes[i].running {
			return i
		}
	}
	return -1
}

// greedy scheduler
func (rm *ResMan) schedule() {
	x := rm.nextFreeNode()
	if x > -1 {
		// assigning the job will also remove it from the job queue
		rm.assign(x)
	}
}

// always assign the very next job in queue?
// the node should already be free before calling this function
func (rm *ResMan) assign(idx int) {
	j := rm.jobs[0]
	rm.nodes[idx].startJob(j)
	rm.jobs = rm.jobs[1:] // remove the very first job
}

func (rm ResMan) PrintStatus() {
	fmt.Printf("ResMan: %v jobs and %v nodes\n", len(rm.jobs), len(rm.nodes))
}
