package model

type Job struct {
	id       int
	duration int
	history  []string
	status   int // may change
}
