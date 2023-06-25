package internal

import (
	"time"

	"github.com/davidmz/go-limiter/internal/clock"
	"github.com/davidmz/go-limiter/internal/engine"
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
		if wait == 0 || wait > timeout {
			return wait == 0
		}
		timeout -= wait
		l.Clock.Sleep(wait)
	}
}
