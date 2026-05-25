package iamprojection

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func TestHandler_HandleUnknownEvent_NoOp(t *testing.T) {
	repo := mocks.NewMockIAMProjectionRepository(t)
	handler := NewHandler(repo)

	err := handler.Handle(context.Background(), messaging.Envelope{Type: "unknown.event"})
	require.NoError(t, err)
}

func TestHandler_HandleTenantCreated_InvalidPayloadDeadLetters(t *testing.T) {
	repo := mocks.NewMockIAMProjectionRepository(t)
	handler := NewHandler(repo)

	err := handler.Handle(context.Background(), messaging.Envelope{
		Type:    "tenant.created",
		Payload: []byte("{"),
	})
	require.Error(t, err)

	classification := messaging.DefaultErrorClassifier().Classify(context.Background(), messaging.Envelope{}, err)
	assert.Equal(t, messaging.FailureActionDeadLetter, classification.Action)
	assert.Equal(t, "invalid tenant.created payload", classification.Reason)
}

func TestHandler_HandleTenantCreated_RepositoryFailureRetries(t *testing.T) {
	repo := mocks.NewMockIAMProjectionRepository(t)
	handler := NewHandler(repo)

	repo.EXPECT().
		UpsertTenant(mock.Anything, "tenant-1", "tenant-1", "Tenant One").
		Return(errors.New("db down")).
		Once()

	payload, _ := json.Marshal(map[string]any{
		"tenant_id":   "tenant-1",
		"tenant_slug": "tenant-1",
		"tenant_name": "Tenant One",
	})
	err := handler.Handle(context.Background(), messaging.Envelope{
		Type:    "tenant.created",
		Payload: payload,
	})
	require.Error(t, err)

	classification := messaging.DefaultErrorClassifier().Classify(context.Background(), messaging.Envelope{}, err)
	assert.Equal(t, messaging.FailureActionRetry, classification.Action)
	assert.Equal(t, "projection store unavailable", classification.Reason)
}
