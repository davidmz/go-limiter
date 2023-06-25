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
	// Due to the nature of Go, there is no way to determine the state of the
	// goroutine from the outside. But for successful time mocking, we need to
	// know when the goroutine started, when it went to sleep, and when it
	// ended. To achieve this, we need to run goroutines in a controlled
	// environment. So if your goroutine uses Sleep, use `Go(fn)` to run it
	// instead of `go fn()`.
	Go(func())
	// The following are time management methods. They are not thread-safe and
	// MUST be called from the single management goroutine.
	Set(time.Time)
	Add(time.Duration)
	WaitForAll()
}
