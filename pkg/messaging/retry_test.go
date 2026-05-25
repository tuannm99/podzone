package messaging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryPolicyNextDelay(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts: 5,
		BaseDelay:   time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2,
	}

	assert.Equal(t, time.Second, policy.NextDelay(1))
	assert.Equal(t, 2*time.Second, policy.NextDelay(2))
	assert.Equal(t, 8*time.Second, policy.NextDelay(4))
	assert.Equal(t, 10*time.Second, policy.NextDelay(8))
}

func TestRetryPolicyExhausted(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 3}
	assert.False(t, policy.Exhausted(2))
	assert.True(t, policy.Exhausted(3))
}
