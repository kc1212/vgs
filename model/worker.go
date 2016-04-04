package model

import (
	"log"
	"time"
)

import "github.com/kc1212/virtual-grid/common"

type WorkerDone struct {
	jobID    int64
	workerID int64
}

type WorkerTask struct {
	task  common.Task
	jobID int64
}

// work represent a single worker, this is a generator function
func work(workerID int64, resultChan chan<- WorkerDone) chan WorkerTask {
	taskChan := make(chan WorkerTask)
	go func() {
		for {
			select {
			case mytask := <-taskChan:
				log.Printf("Worker %v started job %v.\n", workerID, mytask.jobID)
				mytask.task()
				resultChan <- WorkerDone{mytask.jobID, workerID}
				log.Printf("Worker %v finished Job %v.\n", workerID, mytask.jobID)
			}
		}
	}()
	return taskChan
}

// runWorkers receives tasks and schedules them to workers using a greedy algorithm
func runWorkers(n int, tasksChan <-chan WorkerTask, capReq <-chan int, capResp chan<- int) {
	// initialisation
	doneChan := make(chan WorkerDone)
	busyFlags := make([]bool, n) // one flag per worker

	// start all workers, each having their own channel
	workerChans := make([]chan WorkerTask, n)
	for i := 0; i < n; i++ {
		c := work(int64(i), doneChan)
		workerChans[i] = c
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
			if i == -1 {
				log.Print("Warning: nextWorker returned -1, I will try again.")
				time.Sleep(100 * time.Second)
				break
			}
			busyFlags[i] = true
			workerChans[i] <- task
		case done := <-doneChan:
			busyFlags[done.workerID] = false
			// TODO inform GS
		}
	}
}

func countFree(busyFlags []bool) int {
	cnt := 0
	for _, x := range busyFlags {
		if !x {
			cnt++
		}
	}
	return cnt
}

func nextWorker(busyFlags []bool) int {
	for i := range busyFlags {
		if !busyFlags[i] {
			return i
		}
	}
	return -1
}
