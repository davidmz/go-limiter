package limiter_test

import (
	"sync"
	"testing"
	"time"

	limiter "github.com/davidmz/go-limiter"
	"github.com/davidmz/go-limiter/internal"
	"github.com/davidmz/go-limiter/internal/clock"
	"github.com/davidmz/go-limiter/internal/engine"
	"github.com/stretchr/testify/suite"
)

type LimiterTestSuite struct {
	suite.Suite

	limiter limiter.Limiter
	clock   clock.Mock
	results *Results
}

func (s *LimiterTestSuite) AddClock(duration time.Duration) {
	s.clock.Add(duration)
	// Must really sleep to complete all async operations
	time.Sleep(10 * time.Millisecond)
}

func (s *LimiterTestSuite) RunCLient(timeout time.Duration) {
	s.results.Add(s.limiter.Take(timeout))
}

func (s *LimiterTestSuite) ExpectResults(ok, fail int) {
	s.results.lk.Lock()
	defer s.results.lk.Unlock()
	s.Equal(ok, s.results.OkCount, "Expected %d 'true' results", ok)
	s.Equal(fail, s.results.FailCount, "Expected %d 'false' results", fail)
}

func TestLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(LimiterTestSuite))
}

func (s *LimiterTestSuite) SetupTest() {
	s.results = new(Results)
	s.clock = clock.NewMock()
	s.limiter = &internal.Limiter{
		Interval: time.Second,
		Clock:    s.clock,
		Engine:   engine.NewCASBased(),
	}
}

func (s *LimiterTestSuite) TestOneClient() {
	go s.RunCLient(1500 * time.Millisecond)

	// Should serve first client immediately
	s.AddClock(0)
	s.ExpectResults(1, 0)
}

func (s *LimiterTestSuite) TestTwoClients() {

	// T = 0s
	// First client got 'true'.
	go s.RunCLient(1500 * time.Millisecond)
	go s.RunCLient(1500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(1, 0)

	// T = 0.5s
	// Still no result from second client
	s.AddClock(500 * time.Millisecond)
	s.ExpectResults(1, 0)

	// T = 1s
	// Second result appear
	s.AddClock(500 * time.Millisecond)
	s.ExpectResults(2, 0)
}

func (s *LimiterTestSuite) TestThreeClients() {
	go s.RunCLient(1500 * time.Millisecond)
	go s.RunCLient(1500 * time.Millisecond)
	go s.RunCLient(1500 * time.Millisecond)

	// T = 0s
	// First client got 'true'.
	s.AddClock(0)
	s.ExpectResults(1, 0)

	// T = 1s
	// Second client got 'true', third client understands that the next
	// checkpoint will be after timeout so it got 'false'.
	s.AddClock(time.Second)
	s.ExpectResults(2, 1)
}

func (s *LimiterTestSuite) TestSeveralClientsWithDelay() {
	// T = 0s
	// First client got 'true'.
	go s.RunCLient(1500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(1, 0)

	// T = 0.5s
	// Run second client, it waiting
	s.AddClock(500 * time.Millisecond)
	go s.RunCLient(1500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(1, 0)

	// T = 1s
	// Second result got result
	s.AddClock(500 * time.Millisecond)
	s.ExpectResults(2, 0)

	// T = 1.5s
	// Run another client, it waiting
	s.AddClock(500 * time.Millisecond)
	go s.RunCLient(1500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(2, 0)

	// T = 2s
	// Third result got result
	s.AddClock(500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(3, 0)

	// T = 4s
	// After long time, run another client. It should receive 'true'
	// immediately.
	s.AddClock(2 * time.Second)
	go s.RunCLient(1500 * time.Millisecond)
	s.AddClock(0)
	s.ExpectResults(4, 0)
}

func (s *LimiterTestSuite) TestClientsWaves() {
	// T = 0s
	// First wave. First client should got 'true' immediately.
	for i := 0; i < 5; i++ {
		go s.RunCLient(2500 * time.Millisecond)
	}
	s.AddClock(0)
	s.ExpectResults(1, 0)

	// T = 1s
	// Second client got 'true'
	s.AddClock(time.Second)
	s.ExpectResults(2, 0)

	// T = 1.5s
	// Second wave, no changes in result
	s.AddClock(500 * time.Millisecond)
	for i := 0; i < 5; i++ {
		go s.RunCLient(2500 * time.Millisecond)
	}
	s.AddClock(0)
	s.ExpectResults(2, 0)

	// T = 2s
	// Third client (in the test it is first wave client) got 'true', the rest
	// of first wave clients gave up.
	s.AddClock(500 * time.Millisecond)
	s.ExpectResults(3, 2)

	// T = 3s
	// Another client got 'true'
	s.AddClock(time.Second)
	s.ExpectResults(4, 2)

	// T = 4s
	// Another client got 'true', other gave up.
	s.AddClock(time.Second)
	s.ExpectResults(5, 5)
}

type Results struct {
	lk        sync.Mutex
	OkCount   int
	FailCount int
}

func (r *Results) Add(item bool) {
	r.lk.Lock()
	defer r.lk.Unlock()
	if item {
		r.OkCount++
	} else {
		r.FailCount++
	}
}
