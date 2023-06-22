package limiter

import (
	"github.com/davidmz/go-ratelimiter/clock"
	"github.com/davidmz/go-ratelimiter/internal/engine"
)

func WithClock(clock clock.Clock) OptionFn {
	return func(l *limiter) { l.clock = clock }
}

func WithEngine(engine engine.Engine) OptionFn {
	return func(l *limiter) { l.engine = engine }
}

type Option interface {
	Apply(*limiter)
}

type Options []Option

func (o Options) Apply(l *limiter) {
	for _, o := range o {
		o.Apply(l)
	}
}

type OptionFn func(*limiter)

func (f OptionFn) Apply(l *limiter) { f(l) }
