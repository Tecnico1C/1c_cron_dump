package models

import "sync"

type SharedLock struct {
	limit   int
	counter int
	lock    sync.Mutex
}

func NewSharedLock(concurrencyLevel int) SharedLock {
	return SharedLock{
		limit:   concurrencyLevel,
		counter: 0,
	}
}

func (sl *SharedLock) CanStart() bool {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	if sl.counter == sl.limit {
		return false
	}
	sl.counter += 1
	return true
}

func (sl *SharedLock) WorkDone() {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	sl.counter -= 1
}
