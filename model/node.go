package model

type Node struct {
	id      int
	running bool
	job     Job
}

func InitNodes(n int) []Node {
	nodes := make([]Node, n)
	for i, _ := range nodes {
		nodes[i].id = i
	}
	return nodes
}

func (n *Node) startJob(j Job) bool {
	// cannot accept job if it's already running
	if n.running {
		return false
	}
	n.job = j
	n.running = true
	return true
}

func (n *Node) updateStatus() {
	// TODO run in go routine and update running flag
}
