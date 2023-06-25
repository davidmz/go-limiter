package clock_test

import (
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

	s.clock.Go(func() {
		s.clock.Sleep(time.Second)
		end = s.clock.Now()
	})

	s.clock.Add(time.Second)

	s.Equal(time.Second, end.Sub(s.startAt))
}

func (s *ClockTestSuite) TestSleep0() {
	end := time.Time{}

	s.clock.Go(func() {
		s.clock.Sleep(0)
		end = s.clock.Now()
	})

	s.clock.Add(0)

	s.True(end.Equal(s.startAt))
}

func (s *ClockTestSuite) TestSleepTwoGoroutines() {
	var results []time.Duration
	lk := &sync.Mutex{}

	for i := 0; i < 2; i++ {
		s.clock.Go(func() {
			s.clock.Sleep(time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		})
	}

	s.clock.Add(time.Second)

	s.Equal([]time.Duration{time.Second, time.Second}, results)
}

func (s *ClockTestSuite) TestSequentialSleep() {
	var results []time.Duration
	lk := &sync.Mutex{}

	for i := 0; i < 5; i++ {
		i := i
		s.clock.Go(func() {
			s.clock.Sleep(time.Duration(i) * time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		})
	}

	s.clock.WaitForAll()

	s.Equal([]time.Duration{
		0 * time.Second,
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
	}, results)
}

func (s *ClockTestSuite) TestSequentialWaves() {
	var results []time.Duration
	lk := &sync.Mutex{}

	for i := 0; i < 3; i++ {
		i := i
		s.clock.Go(func() {
			s.clock.Sleep(time.Duration(i) * time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		})
	}

	s.clock.Add(2 * time.Second)

	for i := 0; i < 3; i++ {
		i := i
		s.clock.Go(func() {
			s.clock.Sleep(time.Duration(i) * time.Second)

			lk.Lock()
			defer lk.Unlock()
			results = append(results, s.clock.Now().Sub(s.startAt))
		})
	}

	s.clock.WaitForAll()

	s.Equal([]time.Duration{
		0 * time.Second,
		1 * time.Second,
		2 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
	}, results)
}
