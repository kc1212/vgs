package model

type Job struct {
	ID       int64 // must be unique
	Duration int64
	History  []string  // possibly improve the type?
	Status   JobStatus // is this necessary?
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
