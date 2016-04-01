package model

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

func anyJobStatus(st common.JobStatus, jobs []Job) bool {
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
