package toolkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstNonEmpty(t *testing.T) {
	require.Equal(t, "fallback", FirstNonEmpty("", "fallback"))
	require.Equal(t, "value", FirstNonEmpty("value", "fallback"))
	require.Empty(t, FirstNonEmpty("", ""))
}
