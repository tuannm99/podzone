package pdkafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/messaging"
)

func TestExpandTopics_GroupedOrder(t *testing.T) {
	topics := expandTopics(messaging.TopicBootstrapConfig{
		Enabled:          true,
		MainTopics:       []string{"podzone.iam.events"},
		RetryAttempts:    []int{1, 2},
		CreateDeadLetter: true,
	})

	assert.Equal(t, []string{
		"podzone.iam.events",
		"podzone.iam.events.retry.1",
		"podzone.iam.events.retry.2",
		"podzone.iam.events.dlt",
	}, topics)
}

func TestBootstrapTopics_CreatesMissingTopics(t *testing.T) {
	admin := &fakeAdmin{
		topics: map[string]sarama.TopicDetail{
			"podzone.iam.events": {},
		},
	}
	cfg := &Config{
		Topics: messaging.TopicBootstrapConfig{
			Enabled:            true,
			MainTopics:         []string{"podzone.iam.events"},
			RetryAttempts:      []int{1},
			CreateDeadLetter:   true,
			DefaultPartitions:  3,
			DefaultReplication: 1,
		},
	}

	err := BootstrapTopics(admin, cfg)
	require.NoError(t, err)
	assert.Contains(t, admin.topics, "podzone.iam.events.retry.1")
	assert.Contains(t, admin.topics, "podzone.iam.events.dlt")
}
