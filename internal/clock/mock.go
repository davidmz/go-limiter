package clock

import (
	"sync"
	"time"
)

// Mock Clock

type mock struct {
	now      time.Time
	lock     sync.Mutex
	sleepers map[time.Time]chan struct{}
}

func NewMock() Mock {
	// Start from the current time. We will modify 'now' only by .Add, so we
	// don't need to do anything special with monotonic clocks and locales.
	return &mock{
		now:      time.Now(),
		sleepers: make(map[time.Time]chan struct{}),
	}
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
	<-m.sleeperSignal(d)
}

func (m *mock) sleeperSignal(d time.Duration) <-chan struct{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	deadline := m.now.Add(d)
	signal, exists := m.sleepers[deadline]
	if !exists {
		signal = make(chan struct{})
		m.sleepers[deadline] = signal
	}
	return signal
}

func (m *mock) Add(d time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.now = m.now.Add(d)
	for t, signal := range m.sleepers {
		if m.now.Compare(t) >= 0 {
			close(signal)
			delete(m.sleepers, t)
		}
	}
}

func (m *mock) WaitForAll() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for t, ch := range m.sleepers {
		if t.After(m.now) {
			m.now = t
		}
		close(ch)
		delete(m.sleepers, t)
	}
}
