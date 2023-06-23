package clock

import (
	"runtime"
	"sync"
	"time"
)

// Mock Clock

type breakpoint struct {
	At time.Time
	S  chan struct{}
}

func newBreakpoint(at time.Time) *breakpoint {
	return &breakpoint{
		At: at,
		S:  make(chan struct{}),
	}
}

type mock struct {
	now         time.Time
	lock        sync.Mutex
	breakpoints []*breakpoint
}

func NewMock() Mock {
	// Start from the Unix epoch. We will modify 'now' only by .Add, so we
	// don't need to do anything special with monotonic clocks and locales.
	return &mock{now: time.Unix(0, 0)}
}

func (m *mock) Now() time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.now
}

func (m *mock) Sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	<-m.breakSignal(d)
}

func (m *mock) breakSignal(d time.Duration) <-chan struct{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	deadline := m.now.Add(d)
	for i, br := range m.breakpoints {
		if br.At.Equal(deadline) {
			return br.S
		}
		if br.At.After(deadline) {
			newBreak := newBreakpoint(deadline)
			// Insert at 'i' index
			m.breakpoints = append(m.breakpoints[:i+1], m.breakpoints[i:]...)
			m.breakpoints[i] = newBreak
			return newBreak.S
		}
	}

	newBreak := newBreakpoint(deadline)
	m.breakpoints = append(m.breakpoints, newBreak)
	return newBreak.S
}

func (m *mock) Set(newTime time.Time) {
	for m.nextBreakpoint(newTime) {
		runtime.Gosched()
	}

	// Clean unused memory in m.breakpoints underlying array
	m.cleanup()
}

func (m *mock) Add(d time.Duration) { m.Set(m.now.Add(d)) }

func (m *mock) WaitForAll() {
	if len(m.breakpoints) == 0 {
		return
	}

	m.Add(m.breakpoints[len(m.breakpoints)-1].At.Sub(m.now))
}

// Move the inner pointer until deadline, stop at first breakpoint. Returns true
// when stop, so caller should call it again until false.
func (m *mock) nextBreakpoint(deadline time.Time) bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.breakpoints) == 0 || m.breakpoints[0].At.After(deadline) {
		m.now = deadline
		return false
	}

	br := m.breakpoints[0]
	m.breakpoints[0] = nil
	m.breakpoints = m.breakpoints[1:]

	m.now = br.At
	close(br.S)
	return true
}

// Clean unused memory in m.breakpoints underlying array
func (m *mock) cleanup() {
	m.lock.Lock()
	defer m.lock.Unlock()

	newBreakpoints := make([]*breakpoint, len(m.breakpoints))
	copy(newBreakpoints, m.breakpoints)
	m.breakpoints = newBreakpoints
}
