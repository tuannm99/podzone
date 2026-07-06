package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPlacementPlanUpdateDoesNotWriteCreatedAtTwice(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.July, 5, 13, 30, 0, 0, time.UTC)
	update := placementPlanUpdate(placementPlanDoc{
		RequestID: "request-1",
		CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(time.Minute),
	})

	set, ok := update["$set"].(bson.M)
	require.True(t, ok)
	require.NotContains(t, set, "created_at")
	require.Equal(t, "request-1", set["request_id"])

	setOnInsert, ok := update["$setOnInsert"].(bson.M)
	require.True(t, ok)
	require.Equal(t, createdAt, setOnInsert["created_at"])
}
