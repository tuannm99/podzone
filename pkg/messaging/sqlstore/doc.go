// Package sqlstore provides SQL implementations of messaging durability stores.
//
// InboxStore implements idempotent consumer tracking. OutboxStore implements a
// reusable PostgreSQL-style transactional outbox store. Table names are supplied
// by each service so tables can follow the owner pattern, for example
// order_outbox, iam_outbox, or message_outbox for legacy shared usage.
//
// The stores do not create schema automatically. Services own migrations and
// should keep table columns compatible with the methods in this package.
package sqlstore
