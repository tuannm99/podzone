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
	// ID can stay for backward compatibility (e.g. request id),
	// but the "connection identity" should be (TenantID, InfraType, Name).
	ID        string
	TenantID  string
	Name      string // optional: "default", "analytics", ...
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
	EventID   string
	TenantID  string
	Name      string
	InfraType InfraType

	Action string // create|destroy|publish_consul...
	Status string // started|succeeded|failed

	Request map[string]interface{}
	Result  map[string]interface{}
	Error   string

	Actor map[string]string // user/service/ip/trace_id...

	CreatedAt time.Time
}

type OutboxMessage struct {
	EventID string
	Topic   string // "consul.publish"
	Payload map[string]interface{}

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

	AppendEvent(ev ConnectionEvent) error

	EnqueueOutbox(msg OutboxMessage) error
	FindDueOutbox(limit int) ([]OutboxMessage, error)
	MarkOutboxDone(eventID string) error
	MarkOutboxFailed(eventID string, nextRetry time.Time) error
}
