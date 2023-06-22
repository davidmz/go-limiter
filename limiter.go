package limiter

import (
	"time"

	"github.com/davidmz/go-ratelimiter/clock"
	"github.com/davidmz/go-ratelimiter/internal/engine"
)

// Limiter is a rate limiter that guarantees some resource can be taken only
// once during some interval.
type Limiter interface {
	// Take tries to take the resource. It waits until the resource is available
	// and then returns true. The timeout parameter defines the maximum wait
	// time. If the wait time is too long, it will return false.
	Take(timeout time.Duration) bool
}

func New(interval time.Duration, options ...Option) Limiter {
	l := &limiter{interval: interval}

	append(Options{
		WithClock(clock.New()),
		WithEngine(engine.NewCASBased()),
	}, options...).Apply(l)
	return l
}

// Implementation

type limiter struct {
	clock    clock.Clock
	interval time.Duration
	engine   engine.Engine
}

func (l *limiter) Take(timeout time.Duration) bool {
	for {
		wait := l.engine.TryTake(l.clock.Now(), l.interval)
		if wait == 0 {
			return true
		}
		if wait > timeout {
			return false
		}
		timeout -= wait
		l.clock.Sleep(wait)
	}
}
