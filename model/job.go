package model

type Job struct {
	ID       int64 // must be unique
	Duration int64
	History  []string // possibly improve the type?
}

func (j *Job) appendHistory(x string) {
	j.History = append(j.History, x)
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
