package clock_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/davidmz/go-limiter/internal/clock"
	"github.com/stretchr/testify/suite"
)

type ClockTestSuite struct {
	suite.Suite
	startAt time.Time
	clock   clock.Mock
}

func (s *ClockTestSuite) SetupTest() {
	s.clock = clock.NewMock()
	s.startAt = s.clock.Now()
}

func TestClockTestSuite(t *testing.T) {
	suite.Run(t, new(ClockTestSuite))
}

func (s *ClockTestSuite) TestNow() {
	s.Equal(s.startAt, s.clock.Now())

	time.Sleep(10 * time.Millisecond)
	s.Equal(s.startAt, s.clock.Now())
}

func (s *ClockTestSuite) TestSleep() {
	end := time.Time{}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.clock.Sleep(time.Second)
		end = s.clock.Now()
	}()

	// Wait for goroutine to start
	runtime.Gosched()
	s.clock.Add(time.Second)

	// Wait for goroutine to finish
	wg.Wait()

	s.Equal(time.Second, end.Sub(s.startAt))
}

func (s *ClockTestSuite) TestSleepTwoGoroutines() {
	var results []time.Duration
	lk := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			s.clock.Sleep(time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		}()
	}

	// Wait for goroutines to start
	runtime.Gosched()
	s.clock.Add(time.Second)

	// Wait for goroutines to finish
	wg.Wait()

	s.Equal([]time.Duration{time.Second, time.Second}, results)
}

func (s *ClockTestSuite) TestSequentialSleep() {
	var results []time.Duration
	lk := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			defer wg.Done()
			s.clock.Sleep(time.Duration(n) * time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		}(i)
	}

	// Wait for goroutines to start
	runtime.Gosched()
	s.clock.WaitForAll()

	// Wait for goroutines to finish
	wg.Wait()

	s.Equal([]time.Duration{
		0 * time.Second,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
	}, results)
}
