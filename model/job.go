package model

import "time"

// Job are entities can be executed by worker nodes
type Job struct {
	ID         int64 // must be unique
	Duration   time.Duration
	ResMan     string
	StartTime  time.Time
	FinishTime time.Time
}

func filterJobs(s []Job, fn func(Job) bool) []Job {
	var p []Job
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

// takeJobs will take at most n jobs from channel `c`
func takeJobs(n int, c <-chan Job) []Job {
	jobs := make([]Job, 0)
loop:
	for i := 0; i < n; i++ {
		select {
		case job := <-c:
			jobs = append(jobs, job)
		default:
			break loop
		}
	}
	return jobs
}
