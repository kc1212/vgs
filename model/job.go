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
