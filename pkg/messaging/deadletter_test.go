package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicStrategy(t *testing.T) {
	strategy := DefaultTopicStrategy()
	assert.Equal(t, "podzone.iam.events.retry.2", strategy.RetryTopic("podzone.iam.events", 2))
	assert.Equal(t, "podzone.iam.events.dlt", strategy.DeadLetterTopic("podzone.iam.events"))
}
