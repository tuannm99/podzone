package core

import "time"

type InfraType string

const (
	InfraMongo    InfraType = "mongo"
	InfraRedis    InfraType = "redis"
	InfraPostgres InfraType = "postgres"
	InfraElastic  InfraType = "elasticsearch"
	InfraKafka    InfraType = "kafka"
)

type ProvisionInput struct {
	// ID can be used as request id / legacy id, but "connection identity"
	// is (TenantID, InfraType, Name).
	ID        string
	TenantID  string
	Name      string // default, analytics, ...
	InfraType InfraType

	Metadata map[string]string      // cluster, namespace, pod, labels...
	Config   map[string]interface{} // version, size, pool mode...
}

type ProvisionResult struct {
	Endpoint  string
	SecretRef string
	Status    string
}

type ConnectionInfo struct {
	TenantID  string
	Name      string
	InfraType InfraType

	Endpoint  string
	SecretRef string

	Status  string
	Version int64

	Meta   map[string]string
	Config map[string]interface{}

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ConnectionEvent struct {
	// ID must be unique for each event record.
	ID string

	// CorrelationID groups multiple events in one operation (create/destroy/publish).
	CorrelationID string

	TenantID  string
	Name      string
	InfraType InfraType

	Action string // create|destroy|manual_upsert|publish_consul|delete_consul...
	Status string // started|succeeded|failed

	Request map[string]interface{}
	Result  map[string]interface{}
	Error   string

	Actor map[string]string // user/service/ip/trace_id...

	CreatedAt time.Time
}

type OutboxMessage struct {
	// EventID is unique for each outbox message (idempotency key).
	EventID string

	// CorrelationID links outbox execution back to operation.
	CorrelationID string

	Topic   string // consul.publish / consul.delete
	Payload map[string]interface{}

	// Useful metadata for logging/history writing in worker.
	TenantID  string
	InfraType InfraType
	Name      string

	Status     string // pending|done|failed
	RetryCount int
	NextRetry  time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type InfraProvisioner interface {
	Create(input ProvisionInput) (*ProvisionResult, error)
	Destroy(input ProvisionInput) error
}

// ConnectionStore = current state + history + outbox
type ConnectionStore interface {
	Upsert(info ConnectionInfo) error
	SoftDelete(tenantID string, infraType InfraType, name string) error
	Get(tenantID string, infraType InfraType, name string) (*ConnectionInfo, error)

	// Optional query methods for APIs.
	ListConnections(
		tenantID string,
		infraType InfraType,
		includeDeleted bool,
		limit, offset int,
	) ([]ConnectionInfo, error)
	ListEvents(
		tenantID string,
		infraType InfraType,
		name string,
		correlationID string,
		limit, offset int,
	) ([]ConnectionEvent, error)

	AppendEvent(ev ConnectionEvent) error

	EnqueueOutbox(msg OutboxMessage) error
	FindDueOutbox(limit int) ([]OutboxMessage, error)
	MarkOutboxDone(eventID string) error
	MarkOutboxFailed(eventID string, nextRetry time.Time) error
}

