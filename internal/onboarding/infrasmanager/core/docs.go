// Package core contains the domain contracts and orchestration logic for infrastructure provisioning.
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
// - Registry: typed container that resolves InfraProvisioner by InfraType.
// - InfraManager: orchestrates create/destroy and persists ConnectionInfo.
//
// # Responsibilities
//
// - InfraProvisioner: performs actual provisioning work (K8s, Helm, Terraform, etc.).
// - ConnectionStore: persists ConnectionInfo (Consul KV / MongoDB / etc.).
//
// This package should remain framework-agnostic and easy to unit test.
package core
