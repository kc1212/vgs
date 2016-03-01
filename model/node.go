package model

import "fmt"
import "time"

// the compute nodes that does the actual processing
type Node struct {
	running   bool
	startTime int64
	job       Job
}

func InitNodes(n int) []Node {
	return make([]Node, n)
}

func (n *Node) startJob(job Job) {
	// this condition should not happen
	if n.running {
		// TODO panic or return error?
		panic(fmt.Sprintf("Node is already running!\n"))
	}
	n.startTime = time.Now().Unix()
	n.job = job
	n.running = true
}

func (n *Node) poll() {
	now := time.Now().Unix()
	if n.running && (now-n.startTime) > n.job.duration {
		n.running = false
	}
}
