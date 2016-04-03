package model

import (
	"log"
	"time"
)

// Worker is a compute nodes that does the actual processing
type Worker struct {
	startTime int64
	job       Job // should this be nil if job does not exist?
}

// note we're just copying the job
func (n *Worker) startJob(job Job) {
	// this condition should not happen
	if n.isRunning() {
		log.Panicf("Cannot start job %v on worker %v!\n", job, *n)
	}
	log.Printf("Job %v started\n", job)
	n.startTime = time.Now().Unix()
	n.job = job
}

func (n *Worker) isRunning() bool {
	now := time.Now().Unix()
	return n.job.Duration > (now - n.startTime)
}
