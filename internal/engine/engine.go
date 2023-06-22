package engine

import "time"

type Engine interface {
	TryTake(now time.Time, interval time.Duration) time.Duration
}

func NewMutexBased() Engine {
	return &mutexBased{}
}

func NewCASBased() Engine {
	return &casBased{}
}
