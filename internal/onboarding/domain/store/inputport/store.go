package inputport

import (
	"context"
	"time"
)

type RequestStatus string

const (
	RequestStatusRequested       RequestStatus = "requested"
	RequestStatusPendingApproval RequestStatus = "pending_approval"
	RequestStatusQueued          RequestStatus = "queued"
	RequestStatusProvisioning    RequestStatus = "provisioning"
	RequestStatusReady           RequestStatus = "ready"
	RequestStatusFailed          RequestStatus = "failed"
	RequestStatusRejected        RequestStatus = "rejected"
	RequestStatusSuspended       RequestStatus = "suspended"
	RequestStatusArchived        RequestStatus = "archived"
)

type Request struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	Name        string        `json:"name"`
	Subdomain   string        `json:"subdomain"`
	RequestedBy string        `json:"requested_by"`
	Status      RequestStatus `json:"status"`
	StoreID     string        `json:"store_id,omitempty"`
	LastError   string        `json:"last_error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ApprovedAt  *time.Time    `json:"approved_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

type CreateStoreRequestCommand struct {
	Name      string
	Subdomain string
}

type Usecase interface {
	CreateStoreRequest(ctx context.Context, cmd CreateStoreRequestCommand) (*Request, error)
	GetStoreRequest(ctx context.Context, id string) (*Request, error)
	ListStoreRequests(ctx context.Context, workspaceID string) ([]*Request, error)
	RetryStoreRequest(ctx context.Context, id string) error
	ApproveStoreRequest(ctx context.Context, id string) error
	RejectStoreRequest(ctx context.Context, id string) error
	UpdateStoreRequestStatus(ctx context.Context, id string, status RequestStatus) error
	ProcessNextStoreRequest(ctx context.Context) (*Request, error)
	FinalizeNextStoreRequest(ctx context.Context) (*Request, error)
}

type (
	StoreStatus = RequestStatus
	Store       = Request
)
