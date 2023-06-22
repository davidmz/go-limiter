package internal

import (
	"time"

	"github.com/davidmz/go-ratelimiter/internal/clock"
	"github.com/davidmz/go-ratelimiter/internal/engine"
)

// Configurable limiter implementation

type Limiter struct {
	Interval time.Duration
	Clock    clock.Clock
	Engine   engine.Engine
}

func (l *Limiter) Take(timeout time.Duration) bool {
	for {
		wait := l.Engine.TryTake(l.Clock.Now(), l.Interval)
		if wait == 0 {
			return true
		}
		if wait > timeout {
			return false
		}
		timeout -= wait
		l.Clock.Sleep(wait)
	}
}
