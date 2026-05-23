package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopic(t *testing.T) {
	assert.Equal(t, "podzone.iam.events", Topic("iam", "events"))
}
