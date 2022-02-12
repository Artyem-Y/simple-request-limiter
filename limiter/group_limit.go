package limiter

import "time"

type groupLimit struct {
	reqCount int
	resetAt  time.Time
}

func newGroupLimit(interval time.Duration) *groupLimit {
	return &groupLimit{
		reqCount: 0,
		resetAt:  time.Now().Add(interval),
	}
}

func (m *groupLimit) isStale() bool {
	return m.resetAt.Before(time.Now())
}

func (m *groupLimit) refresh(interval time.Duration) {
	m.reqCount = 0
	m.resetAt = time.Now().Add(interval)
}
