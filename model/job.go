package model

//go:generate stringer -type=JobStatus

import "fmt"

type Job struct {
	ID       int64 // must be unique
	Duration int64
	History  []string // possibly improve the type?
	Status   JobStatus
}

type JobStatus int

const (
	Waiting JobStatus = iota
	Submitted
	Running
	Finished
)

func (j *Job) appendHistory(x string) {
	j.History = append(j.History, x)
}

// remember to use stringer -type=JobStatus
func (j Job) String() string {
	return fmt.Sprintf("id: %v, duration: %v, status: %v, history: %v",
		j.ID, j.Duration, j.Status.String(), j.History)
}
