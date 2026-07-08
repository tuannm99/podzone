// Package kafka adapts pkg/messaging contracts to Kafka through pkg/pdkafka.
//
// This package owns envelope serialization, consumer-group handling, retry topic
// publication, dead-letter publication, and the bounded outbox relay used for
// local development or recovery paths. It intentionally depends on pdkafka for
// low-level Sarama producer/consumer lifecycle so service modules can wire Kafka
// once and expose messaging.Publisher/messaging.Consumer to application code.
//
// High-volume outbox publishing should prefer CDC from service-owned outbox
// tables into Kafka. Relay remains useful for local development, smoke tests, and
// low-volume recovery operations.
package kafka
