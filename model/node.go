package model

import (
	"fmt"
	"log"
	"time"
)

// the compute nodes that does the actual processing
type Node struct {
	running   bool
	startTime int64
	job       Job // should this be nil if job does not exist?
}

func InitNodes(n int) []Node {
	return make([]Node, n)
}

// note we're just copying the job
func (n *Node) startJob(job Job) {
	// this condition should not happen
	if n.running || job.status != Waiting {
		log.Fatal(fmt.Sprintf("Cannot start job %v on node %v!\n", job, *n))
	}
	n.startTime = time.Now().Unix()
	n.job = job
	n.job.status = Running
	n.running = true
}

func (n *Node) poll() {
	now := time.Now().Unix()
	if n.running && (now-n.startTime) > n.job.duration {
		n.job.status = Finished
		n.running = false
	}
}
