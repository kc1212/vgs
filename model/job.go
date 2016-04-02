package model

import "math"

import "github.com/kc1212/vgs/common"

type Job struct {
	ID       int64 // must be unique
	Duration int64
	History  []string // possibly improve the type?
	Status   common.JobStatus
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

// idsOfNextNJobs returns the ids of the next `n` jobs that matches st
func idsOfNextNJobs(jobs []Job, n int64, st common.JobStatus) []int64 {
	newJobs := filterJobs(jobs, func(j Job) bool { return j.Status == st })
	ids := make([]int64, len(newJobs))
	for i := range ids {
		ids[i] = newJobs[i].ID
	}
	last := math.Min(float64(n-1), float64(len(newJobs)-1))
	return ids[0:int64(last)]
}
