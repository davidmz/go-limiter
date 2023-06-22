package engine

import (
	"sync"
	"time"
)

type mutexBased struct {
	sync.Mutex
	nextTake time.Time
}

func (e *mutexBased) TryTake(now time.Time, interval time.Duration) time.Duration {
	e.Lock()
	defer e.Unlock()

	if wait := e.nextTake.Sub(now); wait > 0 {
		return wait
	}
	e.nextTake = now.Add(interval)
	return 0
}
