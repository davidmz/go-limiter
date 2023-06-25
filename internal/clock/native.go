package clock

import (
	"time"
)

// Native Clock

type native struct{}

func New() Clock {
	return &native{}
}

func (c *native) Now() time.Time        { return time.Now() }
func (c *native) Sleep(d time.Duration) { time.Sleep(d) }
