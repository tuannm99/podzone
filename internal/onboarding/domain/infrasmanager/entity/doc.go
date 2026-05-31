// Package entity contains domain entities and pure helpers for infrastructure provisioning.
//
// # Key types
//
// - InfraType: identifies an infrastructure kind (mongo/redis/postgres/...)
// - ProvisionInput: request payload for provisioning/destroying
// - ProvisionResult: response from provisioner
// - ConnectionInfo: stored connection metadata for later consumption
//
// # Main components
//
// - BuildConsulKey: pure helper for connection snapshot keys.
//
// # Responsibilities
//
// - ConnectionInfo, ConnectionEvent, OutboxMessage: persisted domain records.
//
// This package should remain framework-agnostic and easy to unit test.
package entity
