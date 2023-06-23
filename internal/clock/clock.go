package clock

import (
	"time"
)

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type Mock interface {
	Clock
	// The following are time management methods. They are not thread-safe and
	// MUST be called from the single management goroutine.
	Set(time.Time)
	Add(time.Duration)
	WaitForAll()
}
