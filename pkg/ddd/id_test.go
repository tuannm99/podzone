package ddd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseIDRequiresValue(t *testing.T) {
	t.Parallel()

	_, err := ParseID(" ")

	require.Error(t, err)
}

func TestNewIDDelegatesToParseID(t *testing.T) {
	t.Parallel()

	id, err := NewID("store_1")

	require.NoError(t, err)
	require.Equal(t, ID("store_1"), id)
}

func TestUUIDGeneratorCreatesPrefixedID(t *testing.T) {
	t.Parallel()

	id, err := NewUUIDGenerator().NewID("Customer Order")

	require.NoError(t, err)
	require.True(t, strings.HasPrefix(id.String(), "customer_order_"))
	require.False(t, id.IsZero())
}

func TestUUIDGeneratorSupportsUnprefixedID(t *testing.T) {
	t.Parallel()

	id, err := NewUUIDGenerator().NewID("")

	require.NoError(t, err)
	require.NotEmpty(t, id.String())
	require.NotContains(t, id.String(), "__")
}
