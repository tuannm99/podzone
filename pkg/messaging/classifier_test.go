package messaging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyError(t *testing.T) {
	retry := ClassifyError(RetryableError(errors.New("transient"), "retry later"))
	assert.Equal(t, FailureActionRetry, retry.Action)
	assert.Equal(t, "retry later", retry.Reason)

	dlt := ClassifyError(DeadLetterError(errors.New("invalid"), "bad payload"))
	assert.Equal(t, FailureActionDeadLetter, dlt.Action)

	drop := ClassifyError(DropError(errors.New("duplicate"), "already applied"))
	assert.Equal(t, FailureActionDrop, drop.Action)

	unknown := ClassifyError(errors.New("boom"))
	assert.Equal(t, FailureActionReturn, unknown.Action)
}
