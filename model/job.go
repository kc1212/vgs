package model

type Job struct {
	id       int
	duration int64
	history  []string  // possibly improve the type?
	status   JobStatus // is this necessary?
}

type JobStatus int

const (
	Waiting JobStatus = iota
	Running
	Finished
)

func (j *Job) appendHistory(x string) {
	j.history = append(j.history, x)
}
