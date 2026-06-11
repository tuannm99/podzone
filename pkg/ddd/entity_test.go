package ddd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntityBaseRequiresID(t *testing.T) {
	t.Parallel()

	_, err := NewEntityBase("")

	require.Error(t, err)
}

func TestEntityBaseReturnsID(t *testing.T) {
	t.Parallel()

	id, err := ParseID("store_1")
	require.NoError(t, err)

	entity, err := NewEntityBase(id)

	require.NoError(t, err)
	require.Equal(t, id, entity.EntityID())
}
