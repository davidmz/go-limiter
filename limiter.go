package limiter

import (
	"time"

	"github.com/davidmz/go-ratelimiter/internal"
	"github.com/davidmz/go-ratelimiter/internal/clock"
	"github.com/davidmz/go-ratelimiter/internal/engine"
)

// Limiter is an interface that defines the contract for a resource limiter.
type Limiter interface {
	// Take tries to acquire access to the resource within the specified timeout.
	// It returns true if access is granted and false if the waiting time exceeds
	// the timeout. Take may be invoked concurrently by multiple clients (goroutines).
	Take(timeout time.Duration) bool
}

// New creates a new Limiter instance with the specified interval.
func New(interval time.Duration) Limiter {
	return &internal.Limiter{
		Interval: interval,
		Clock:    clock.New(),
		Engine:   engine.NewCASBased(),
	}
}
