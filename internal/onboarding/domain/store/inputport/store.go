package inputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/collection"
)

type RequestStatus string

const (
	RequestStatusRequested            RequestStatus = "requested"
	RequestStatusPlanning             RequestStatus = "planning"
	RequestStatusPlanned              RequestStatus = "planned"
	RequestStatusPendingApproval      RequestStatus = "pending_approval"
	RequestStatusQueued               RequestStatus = "queued"
	RequestStatusProvisioning         RequestStatus = "provisioning"
	RequestStatusReady                RequestStatus = "ready"
	RequestStatusFailed               RequestStatus = "failed"
	RequestStatusFailedRetryable      RequestStatus = "failed_retryable"
	RequestStatusFailedNonRetryable   RequestStatus = "failed_non_retryable"
	RequestStatusPendingPlatformSetup RequestStatus = "pending_platform_setup"
	RequestStatusRejected             RequestStatus = "rejected"
	RequestStatusSuspended            RequestStatus = "suspended"
	RequestStatusArchived             RequestStatus = "archived"
	RequestStatusCancelled            RequestStatus = "cancelled"
)

type Request struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	Name        string        `json:"name"`
	Subdomain   string        `json:"subdomain"`
	RequestedBy string        `json:"requested_by"`
	OwnerID     string        `json:"owner_id"`
	Status      RequestStatus `json:"status"`
	StoreID     string        `json:"store_id,omitempty"`
	LastError   string        `json:"last_error,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ApprovedAt  *time.Time    `json:"approved_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

type RequestTransition struct {
	ID        string            `json:"id"`
	RequestID string            `json:"request_id"`
	From      RequestStatus     `json:"from"`
	To        RequestStatus     `json:"to"`
	Actor     map[string]string `json:"actor"`
	Step      string            `json:"step"`
	Reason    string            `json:"reason"`
	ErrorCode string            `json:"error_code"`
	CreatedAt time.Time         `json:"created_at"`
}

type CreateStoreRequestCommand struct {
	Name      string
	Subdomain string
	OwnerID   string
}

type Usecase interface {
	CreateStoreRequest(ctx context.Context, cmd CreateStoreRequestCommand) (*Request, error)
	GetStoreRequest(ctx context.Context, id string) (*Request, error)
	ListStoreRequests(
		ctx context.Context,
		workspaceID string,
		query collection.Query,
	) (collection.Page[*Request], error)
	ListStoreRequestTransitions(
		ctx context.Context,
		id string,
		query collection.Query,
	) (collection.Page[RequestTransition], error)
	RetryStoreRequest(ctx context.Context, id string) error
	ApproveStoreRequest(ctx context.Context, id string) error
	RejectStoreRequest(ctx context.Context, id string) error
	UpdateStoreRequestStatus(ctx context.Context, id string, status RequestStatus) error
	ProcessNextStoreRequest(ctx context.Context, workerID string) (*Request, error)
	FinalizeNextStoreRequest(ctx context.Context, workerID string) (*Request, error)
}

type (
	StoreStatus = RequestStatus
	Store       = Request
)
