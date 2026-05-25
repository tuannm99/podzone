package messaging

import (
	"math"
	"time"
)

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

func (p RetryPolicy) Normalize() RetryPolicy {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 3
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = time.Second
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = time.Minute
	}
	if p.Multiplier <= 1 {
		p.Multiplier = 2
	}
	return p
}

func (p RetryPolicy) NextDelay(attempt int) time.Duration {
	p = p.Normalize()
	if attempt <= 1 {
		return p.BaseDelay
	}
	factor := math.Pow(p.Multiplier, float64(attempt-1))
	delay := time.Duration(float64(p.BaseDelay) * factor)
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	return delay
}

func (p RetryPolicy) Exhausted(attempt int) bool {
	return attempt >= p.Normalize().MaxAttempts
}
