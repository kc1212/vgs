package model

import (
	"fmt"
	"log"
	"time"
)

import "github.com/kc1212/vgs/common"

// Worker is a compute nodes that does the actual processing
type Worker struct {
	running   bool
	startTime int64
	job       Job // should this be nil if job does not exist?
}

// note we're just copying the job
func (n *Worker) startJob(job Job) {
	// this condition should not happen
	if n.running || job.Status != common.Waiting {
		log.Fatal(fmt.Sprintf("Cannot start job %v on worker %v!\n", job, *n))
	}
	n.startTime = time.Now().Unix()
	n.job = job
	n.job.Status = common.Running
	n.running = true
}

func (n *Worker) poll() {
	now := time.Now().Unix()
	if n.running && (now-n.startTime) > n.job.Duration {
		n.job.Status = common.Finished
		n.running = false
	}
}
