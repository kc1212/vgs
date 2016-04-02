package model

import (
	"log"
	"time"
)

import "github.com/kc1212/vgs/common"

// Worker is a compute nodes that does the actual processing
type Worker struct {
	startTime int64
	job       Job // should this be nil if job does not exist?
}

// note we're just copying the job
func (n *Worker) startJob(job Job) {
	// this condition should not happen
	if job.Status != common.Waiting {
		log.Panicf("Cannot start job %v on worker %v!\n", job, *n)
	}
	n.startTime = time.Now().Unix()
	n.job = job
	n.job.Status = common.Running
}

func (n *Worker) isRunning() bool {
	n.update()
	return n.job.Status == common.Running
}

func (n *Worker) update() {
	now := time.Now().Unix()
	if n.job.Status == common.Running && (now-n.startTime) > n.job.Duration {
		n.job.Status = common.Finished
	}
}
