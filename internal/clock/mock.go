package clock

import (
	"sync"
	"time"
)

// Mock Clock

type breakpoint struct {
	sync.WaitGroup
	At time.Time
	S  chan struct{}
}

func newBreakpoint(at time.Time) *breakpoint {
	br := &breakpoint{
		At: at,
		S:  make(chan struct{}),
	}
	br.Add(1)
	return br
}

func (br *breakpoint) release() {
	close(br.S)
	// Wait for all sleepers to start
	br.Wait()
}

type mock struct {
	now         time.Time
	lock        sync.Mutex
	breakpoints []*breakpoint
	threads     *sync.WaitGroup
}

func NewMock() Mock {
	// Start from the Unix epoch. We will modify 'now' only by .Add, so we
	// don't need to do anything special with monotonic clocks and locales.
	return &mock{
		now:     time.Unix(0, 0),
		threads: new(sync.WaitGroup),
	}
}

func (m *mock) Now() time.Time {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.now
}

func (m *mock) Go(fn func()) {
	m.threads.Add(1)
	go func() {
		defer m.threads.Done()
		fn()
	}()
}

func (m *mock) Sleep(d time.Duration) {
	if d < 0 {
		panic("attempt to sleep with negative duration")
	}

	if d == 0 {
		return
	}
	br := m.addBreakpointAt(d)
	go m.threads.Done()
	<-br.S
	m.threads.Add(1)
	br.Done()
}

func (m *mock) addBreakpointAt(d time.Duration) *breakpoint {
	m.lock.Lock()
	defer m.lock.Unlock()

	deadline := m.now.Add(d)
	for i, br := range m.breakpoints {
		if br.At.Equal(deadline) {
			br.Add(1)
			return br
		}
		if br.At.After(deadline) {
			newBreak := newBreakpoint(deadline)
			// Insert at 'i' index
			m.breakpoints = append(m.breakpoints[:i+1], m.breakpoints[i:]...)
			m.breakpoints[i] = newBreak
			return newBreak
		}
	}

	newBreak := newBreakpoint(deadline)
	m.breakpoints = append(m.breakpoints, newBreak)
	return newBreak
}

func (m *mock) Set(newTime time.Time) {
	if newTime.Before(m.now) {
		panic("attempt to set time before current")
	}

	// Before moving the time forward, we need to make sure that all
	// sleepers are blocked on their Sleep calls.
	m.threads.Wait()
	for {
		br := m.moveToNextBreakpoint(newTime)
		if br == nil {
			break
		}
		br.release()
		// Wait for all sleepers to sleep again
		m.threads.Wait()
	}

	// Clean unused memory in m.breakpoints underlying array
	m.cleanup()
}

func (m *mock) Add(d time.Duration) { m.Set(m.now.Add(d)) }

func (m *mock) WaitForAll() {
	for {
		m.threads.Wait()
		if len(m.breakpoints) == 0 {
			return
		}
		br := m.moveToNextBreakpoint(m.breakpoints[0].At)
		br.release()
	}
}

// Move the inner pointer until deadline, stop at first breakpoint. Returns
// found breakpoint when stop, so caller should call it again until nil.
func (m *mock) moveToNextBreakpoint(deadline time.Time) *breakpoint {
	if len(m.breakpoints) == 0 || m.breakpoints[0].At.After(deadline) {
		m.now = deadline
		return nil
	}

	br := m.breakpoints[0]
	m.breakpoints[0] = nil
	m.breakpoints = m.breakpoints[1:]

	m.now = br.At
	return br
}

// Clean unused memory in m.breakpoints underlying array
func (m *mock) cleanup() {
	m.lock.Lock()
	defer m.lock.Unlock()

	newBreakpoints := make([]*breakpoint, len(m.breakpoints))
	copy(newBreakpoints, m.breakpoints)
	m.breakpoints = newBreakpoints
}
