package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

type StoreRequest struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	WorkspaceID string              `bson:"workspace_id"           json:"workspace_id"`
	Name        string              `bson:"name"                   json:"name"`
	Subdomain   string              `bson:"subdomain"              json:"subdomain"`
	RequestedBy string              `bson:"requested_by"           json:"requested_by"`
	Status      RequestStatus       `bson:"status"                 json:"status"`
	StoreID     *primitive.ObjectID `bson:"store_id,omitempty"     json:"store_id,omitempty"`
	LastError   string              `bson:"last_error,omitempty"   json:"last_error,omitempty"`
	CreatedAt   time.Time           `bson:"created_at"             json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at"             json:"updated_at"`
	ApprovedAt  *time.Time          `bson:"approved_at,omitempty"  json:"approved_at,omitempty"`
	CompletedAt *time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type StoreRequestTransition struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RequestID string             `bson:"request_id"     json:"request_id"`
	From      RequestStatus      `bson:"from"           json:"from"`
	To        RequestStatus      `bson:"to"             json:"to"`
	Actor     map[string]string  `bson:"actor"          json:"actor"`
	Reason    string             `bson:"reason"         json:"reason"`
	ErrorCode string             `bson:"error_code"     json:"error_code"`
	CreatedAt time.Time          `bson:"created_at"     json:"created_at"`
}

type ProvisioningConfig struct {
	Enabled      bool
	AutoApprove  bool
	ClusterName  string
	Mode         string
	DBName       string
	SchemaPrefix string
}
