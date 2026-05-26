package iamprojection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type TenantCreatedHandler struct {
	repo outputport.IAMProjectionRepository
}

var _ messaging.TypedHandler = (*TenantCreatedHandler)(nil)

func NewTenantCreatedHandler(repo outputport.IAMProjectionRepository) *TenantCreatedHandler {
	return &TenantCreatedHandler{repo: repo}
}

func (h *TenantCreatedHandler) MessageType() string {
	return "tenant.created"
}

func (h *TenantCreatedHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
	var payload struct {
		TenantID   string `json:"tenant_id"`
		TenantSlug string `json:"tenant_slug"`
		TenantName string `json:"tenant_name"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return messaging.DeadLetterError(
			fmt.Errorf("decode tenant.created payload: %w", err),
			"invalid tenant.created payload",
		)
	}
	if err := h.repo.UpsertTenant(
		ctx, payload.TenantID, payload.TenantSlug, payload.TenantName); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("project tenant.created: %w", err),
			"projection store unavailable",
		)
	}
	return nil
}
