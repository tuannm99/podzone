package ddd

import "time"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

var _ Clock = (*SystemClock)(nil)

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type FixedClock struct {
	now time.Time
}

var _ Clock = (*FixedClock)(nil)

func NewFixedClock(now time.Time) *FixedClock {
	return &FixedClock{now: now.UTC()}
}

func (c *FixedClock) Now() time.Time {
	return c.now
}
