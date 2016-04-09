package model

type Job struct {
	ID       int64 // must be unique
	Duration int64
	ResMan   string // possibly improve the type?
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

// chanToJobs attempts to copy all the jobs from a buffered channel
// without modifying that channel
// TODO NOT SURE IF THIS IS THE RIGHT WAY TO THIS
func chanToJobs(c *chan Job, cap int) []Job {
	jobs := make([]Job, 0)
	c2 := *c
	*c = make(chan Job, cap)
	close(c2)

	for job := range c2 {
		*c <- job
		jobs = append(jobs, job)
	}
	return jobs
}
