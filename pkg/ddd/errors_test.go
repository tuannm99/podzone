package ddd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainErrorFormatsCodeAndMessage(t *testing.T) {
	t.Parallel()

	err := NewDomainError("ORDER_ALREADY_CANCELLED", "order is already cancelled")

	require.Equal(t, "[ORDER_ALREADY_CANCELLED] order is already cancelled", err.Error())
	require.True(t, IsDomainError(err))
	require.True(t, IsDomainError(fmt.Errorf("wrap: %w", err)))
}

func TestDomainErrorHandlesEmptyFields(t *testing.T) {
	t.Parallel()

	require.Equal(t, "ORDER_ALREADY_CANCELLED", NewDomainError("ORDER_ALREADY_CANCELLED", "").Error())
	require.Equal(t, "order is already cancelled", NewDomainError("", "order is already cancelled").Error())
}
