package model

import (
	"log"
	"time"
)

import "github.com/kc1212/virtual-grid/common"

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

type WorkDone struct {
	jobID    int
	workerID int
}

type WorkerTask struct {
	task  common.Task
	jobID int
}

// work represent a single worker
func work(workerID int, taskChan <-chan WorkerTask, resultChan chan<- WorkDone) {
	for {
		select {
		case mytask := <-taskChan:
			mytask.task()
			resultChan <- WorkDone{mytask.jobID, workerID}
		}
	}
}

func runWorkers(n int, tasksChan <-chan WorkerTask, capReq <-chan int, capResp chan<- int) {
	// initialisation
	workerChans := make([]chan WorkerTask, n)
	for i := range workerChans {
		workerChans[i] = make(chan WorkerTask)
	}
	doneChan := make(chan WorkDone)
	busyFlags := make([]bool, n)

	// start all workers
	for i := 0; i < n; i++ {
		go work(i, workerChans[i], doneChan)
	}

	// handle new jobs and assign them to workers
	// cap{Req,Resp} is for reporting the current capacity
	// block until there's something to do
	for {
		select {
		case <-capReq:
			capResp <- countFree(busyFlags)
		case task := <-tasksChan:
			i := nextWorker(busyFlags)
			busyFlags[i] = true
			workerChans[i] <- task
		case done := <-doneChan:
			busyFlags[done.workerID] = false
			// TODO inform GS
		}
	}
}

func countFree(running []bool) int {
	cnt := 0
	for _, x := range running {
		if !x {
			cnt++
		}
	}
	return cnt
}

func nextWorker(running []bool) int {
	for i := range running {
		if !running[i] {
			return i
		}
	}
	log.Panic("GS shouldn't send jobs to a busy RM.")
	return -1 // unreachable
}
