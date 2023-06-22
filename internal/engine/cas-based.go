package engine

import (
	"sync/atomic"
	"time"
)

type casBased struct {
	nextTake atomic.Int64
}

func (e *casBased) TryTake(nowTime time.Time, interval time.Duration) time.Duration {
	now := nowTime.UnixNano()
	newNext := now + int64(interval)
	for {
		next := e.nextTake.Load()
		if wait := next - now; wait > 0 {
			return time.Duration(wait)
		}
		if e.nextTake.CompareAndSwap(next, newNext) {
			return 0
		}
	}
}
