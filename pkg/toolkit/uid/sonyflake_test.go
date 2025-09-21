package uid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ReturnsUint64NonZero(t *testing.T) {
	id, err := New()
	require.NoError(t, err)
	assert.IsType(t, uint64(0), id)
	assert.NotZero(t, id)
}
