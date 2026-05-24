package iam

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	outputportmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/messaging"
)

func TestHandlerHandleTenantCreated(t *testing.T) {
	repo := outputportmocks.NewMockIAMProjectionRepository(t)
	repo.EXPECT().UpsertTenant(mock.Anything, "t1", "tenant-one", "Tenant One").Return(nil).Once()

	payload, err := json.Marshal(map[string]any{
		"tenant_id":   "t1",
		"tenant_slug": "tenant-one",
		"tenant_name": "Tenant One",
	})
	require.NoError(t, err)

	h := NewHandler(repo)
	require.NoError(t, h.Handle(context.Background(), messaging.Envelope{
		ID:            "e1",
		Type:          "tenant.created",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       payload,
	}))
}

func TestHandlerHandleTenantMemberAdded(t *testing.T) {
	repo := outputportmocks.NewMockIAMProjectionRepository(t)
	repo.EXPECT().UpsertTenantMembership(mock.Anything, "t1", uint(11), "tenant_viewer", "active").Return(nil).Once()

	payload, err := json.Marshal(map[string]any{
		"tenant_id": "t1",
		"user_id":   11,
		"role_name": "tenant_viewer",
		"status":    "active",
	})
	require.NoError(t, err)

	h := NewHandler(repo)
	require.NoError(t, h.Handle(context.Background(), messaging.Envelope{
		ID:            "e2",
		Type:          "tenant.member.added",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       payload,
	}))
}

func TestHandlerHandleUnknownEventNoop(t *testing.T) {
	repo := outputportmocks.NewMockIAMProjectionRepository(t)
	h := NewHandler(repo)
	require.NoError(t, h.Handle(context.Background(), messaging.Envelope{
		ID:            "e3",
		Type:          "policy.attached",
		Source:        "iam",
		OccurredAt:    time.Now().UTC(),
		SchemaVersion: 1,
		Payload:       []byte(`{}`),
	}))
}
