package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithDeliveryMetadata(t *testing.T) {
	env := Envelope{ID: "evt_1"}
	updated := WithDeliveryMetadata(env, DeliveryMetadata{
		Attempt:          2,
		MaxAttempts:      5,
		OriginalTopic:    "podzone.iam.events",
		LastError:        "boom",
		DeadLetterReason: "validation",
		RedriveCount:     1,
		ConsumerName:     "auth.iam-projection",
	})

	assert.Equal(t, 2, ReadDeliveryMetadata(updated).Attempt)
	assert.Equal(t, "podzone.iam.events", ReadDeliveryMetadata(updated).OriginalTopic)
	assert.Equal(t, "auth.iam-projection", ReadDeliveryMetadata(updated).ConsumerName)
}
