package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateStoreValidatesAndEmitsEvent(t *testing.T) {
	t.Parallel()

	store, events, err := CreateStore(
		"Urban Finds",
		"Print-on-demand storefront",
		"user-1",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.NotEmpty(t, store.ID)
	require.Equal(t, StoreStatusDraft, store.Status)
	require.Len(t, events, 1)
	require.Equal(t, "StoreCreated", events[0].EventType())

	pending := store.PullEvents()
	require.Len(t, pending, 1)
	require.Equal(t, "StoreCreated", pending[0].EventType())
	require.Empty(t, store.PullEvents())
}

func TestCreateStoreRequiresNameAndOwner(t *testing.T) {
	t.Parallel()

	_, _, err := CreateStore("", "description", "user-1", time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))
	require.Error(t, err)

	_, _, err = CreateStore("Urban Finds", "description", "", time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))
	require.Error(t, err)
}

func TestStoreActivationEmitsEvents(t *testing.T) {
	t.Parallel()

	store, _, err := CreateStore(
		"Urban Finds",
		"Print-on-demand storefront",
		"user-1",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	store.PullEvents()

	store.Activate(time.Date(2026, 6, 4, 11, 0, 0, 0, time.UTC))
	require.True(t, store.IsActive)
	require.Equal(t, StoreStatusActive, store.Status)
	require.Equal(t, "StoreActivated", store.PullEvents()[0].EventType())

	store.Deactivate(time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC))
	require.False(t, store.IsActive)
	require.Equal(t, StoreStatusInactive, store.Status)
	require.Equal(t, "StoreDeactivated", store.PullEvents()[0].EventType())
}
