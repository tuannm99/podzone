// Package pdkafka owns low-level Kafka client configuration and Sarama lifecycle.
//
// Use this package when wiring producers, admins, consumer group factories, and
// Fx modules. Business and service code should generally depend on
// pkg/messaging interfaces instead of Sarama directly. The pkg/messaging/kafka
// adapter converts pdkafka producers/runners into messaging.Publisher and
// messaging.Consumer implementations.
package pdkafka
