package iamprojection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

type TenantMemberAddedHandler struct {
	repo outputport.IAMProjectionRepository
}

var _ messaging.TypedHandler = (*TenantMemberAddedHandler)(nil)

func NewTenantMemberAddedHandler(repo outputport.IAMProjectionRepository) *TenantMemberAddedHandler {
	return &TenantMemberAddedHandler{repo: repo}
}

func (h *TenantMemberAddedHandler) MessageType() string {
	return "tenant.member.added"
}

func (h *TenantMemberAddedHandler) Handle(ctx context.Context, msg messaging.Envelope) error {
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
	if err := h.repo.UpsertTenantMembership(
		ctx, payload.TenantID, payload.UserID, payload.RoleName, payload.Status); err != nil {
		return messaging.RetryableError(
			fmt.Errorf("project tenant.member.added: %w", err),
			"projection store unavailable",
		)
	}
	return nil
}
