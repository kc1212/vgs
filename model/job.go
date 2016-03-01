package model

type Job struct {
	id       int
	duration int64
	history  []string
	status   int // may change
}
