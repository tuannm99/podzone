package entity

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
