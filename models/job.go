package models

import "time"

type JobStatus interface {
	IsCompleted() bool
	GetNextTick() time.Time
	GetInfobase() *Infobase
}

type CompletedJob struct{}

func (cj *CompletedJob) IsCompleted() bool {
	return true
}

func (cj *CompletedJob) GetNextTick() time.Time {
	return time.Time{}
}

func (cj *CompletedJob) GetInfobase() *Infobase {
	return nil
}

type FailedJob struct {
	Infobase *Infobase
	NextTick time.Time
}

func (fj *FailedJob) IsCompleted() bool {
	return false
}

func (fj *FailedJob) GetNextTick() time.Time {
	return fj.NextTick
}

func (fj *FailedJob) GetInfobase() *Infobase {
	return fj.Infobase
}
