package ddd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSystemClockReturnsUTCTime(t *testing.T) {
	t.Parallel()

	now := NewSystemClock().Now()

	require.Equal(t, time.UTC, now.Location())
	require.False(t, now.IsZero())
}

func TestFixedClockReturnsConfiguredUTCTime(t *testing.T) {
	t.Parallel()

	source := time.Date(2026, 6, 5, 10, 30, 0, 0, time.FixedZone("ICT", 7*60*60))
	clock := NewFixedClock(source)

	require.Equal(t, source.UTC(), clock.Now())
	require.Equal(t, time.UTC, clock.Now().Location())
}
