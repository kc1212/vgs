package model

import "github.com/kc1212/virtual-grid/common"

type Job struct {
	ID       int64 // must be unique
	Duration int64
	History  []string         // possibly improve the type?
	Status   common.JobStatus // TODO probably don't need this
}

func (j *Job) appendHistory(x string) {
	j.History = append(j.History, x)
}

func anyJobHasStatus(st common.JobStatus, jobs []Job) bool {
	for _, job := range jobs {
		if st == job.Status {
			return true
		}
	}
	return false
}

func jobsFromIDs(ids []int64, jobs []Job) []Job {
	newJobs := make([]Job, len(ids))
	i := 0
	for _, id := range ids {
		for _, job := range jobs {
			if id == job.ID {
				newJobs[i] = job
				i++
				break
			}
		}
	}
	return newJobs
}

func updateJobs(ids []int64, jobs []Job, st common.JobStatus) []Job {
	for _, id := range ids {
		for i := range jobs {
			if id == jobs[i].ID {
				jobs[i].Status = st
				break
			}
		}
	}
	return jobs
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

func dropJobs(n int, c <-chan Job) {
	for i := 0; i < n; i++ {
		select {
		case <-c:
		default:
			return
		}
	}
}

// takeJobs will take at most n jobs from channel `c`
func takeJobs(n int, c <-chan Job) []Job {
	jobs := *new([]Job)
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

func jobsToChan(jobs []Job, c chan<- Job) {
	for _, j := range jobs {
		c <- j
	}
}
