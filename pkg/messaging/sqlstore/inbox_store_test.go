package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInboxStore_DefaultTableName(t *testing.T) {
	store, err := NewInboxStore(nil, "")
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, "message_inbox", store.tableName)
}

func TestNewInboxStore_RejectsInvalidIdentifier(t *testing.T) {
	store, err := NewInboxStore(nil, "message_inbox;drop table users")
	require.Error(t, err)
	assert.Nil(t, store)
}

func TestNewInboxStore_AllowsQualifiedIdentifier(t *testing.T) {
	store, err := NewInboxStore(nil, "public.message_inbox")
	require.NoError(t, err)
	require.NotNil(t, store)
	assert.Equal(t, "public.message_inbox", store.tableName)
}
