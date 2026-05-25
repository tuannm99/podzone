package iamprojection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type Handler struct {
	repo outputport.IAMProjectionRepository
}

var _ messaging.Handler = (*Handler)(nil)

func NewHandler(repo outputport.IAMProjectionRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, msg messaging.Envelope) error {
	switch msg.Type {
	case "tenant.created":
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
		if err := h.repo.UpsertTenant(ctx, payload.TenantID, payload.TenantSlug, payload.TenantName); err != nil {
			return messaging.RetryableError(
				fmt.Errorf("project tenant.created: %w", err),
				"projection store unavailable",
			)
		}
		return nil
	case "tenant.member.added":
		var payload struct {
			TenantID string `json:"tenant_id"`
			UserID   uint   `json:"user_id"`
			RoleName string `json:"role_name"`
			Status   string `json:"status"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return messaging.DeadLetterError(
				fmt.Errorf("decode tenant.member.added payload: %w", err),
				"invalid tenant.member.added payload",
			)
		}
		if err := h.repo.UpsertTenantMembership(ctx, payload.TenantID, payload.UserID, payload.RoleName, payload.Status); err != nil {
			return messaging.RetryableError(
				fmt.Errorf("project tenant.member.added: %w", err),
				"projection store unavailable",
			)
		}
		return nil
	default:
		return nil
	}
}
