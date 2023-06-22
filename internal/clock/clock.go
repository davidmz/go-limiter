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
	Add(time.Duration)
	WaitForAll()
}
