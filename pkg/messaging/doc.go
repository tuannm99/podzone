// Package messaging defines Podzone's service-to-service messaging contracts.
//
// Use Publisher for normal best-effort async jobs where publishing does not need
// to be committed atomically with a database transaction. Use OutboxStore plus
// TransactionalOutboxPublisher only for commit-coupled integration events where
// losing the event after the aggregate/data commit would break another workflow.
//
// The package is split by responsibility:
//   - Envelope: stable message metadata and JSON payload.
//   - Publisher/Consumer/Handler: runtime contracts.
//   - Registry: routes Envelope.Type to TypedHandler implementations.
//   - Retry, dead-letter, redrive, idempotency: consumer operational behavior.
//   - kafka: Sarama-backed adapter built on pkg/pdkafka.
//   - sqlstore: reusable SQL stores for inbox/outbox tables.
//
// Service-specific event handlers belong under internal/<service>/controller,
// while workers, Kafka wiring, and persistence adapters stay under
// internal/<service>/infrastructure.
package messaging
