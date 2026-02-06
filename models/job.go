package models

type JobStatus int

const (
	Completed JobStatus = iota
	Postponed
)

type Job struct {
	Infobase *Infobase
	Next     *Job
}

type JobResponse struct {
	Job    *Job
	Status JobStatus
}
