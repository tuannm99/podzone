// Package infrasmanager provides infrastructure provisioning & connection registration for onboarding.
//
// Overview
//
// The onboarding service needs to:
//   1) provision infrastructure (Mongo/Redis/Postgres/Elastic/Kafka, etc.) in a target environment
//   2) store connection information (endpoint, auth secret reference, metadata)
//   3) keep history/audit of provisioning actions for traceability
//
// Architecture (high-level)
//
//   Controller / API layer
//          |
//          v
//   InfraManager (core)
//     - validates request
//     - delegates to a provisioner by InfraType via Registry
//     - persists connection info via ConnectionStore
//     - writes History events (if enabled)
//          |
//          v
//   Providers
//     - InfraProvisioner implementations (K8s/Helm/Terraform/CLI, etc.)
//     - ConnectionStore implementations (Consul KV, MongoDB, etc.)
//     - History store (Consul KV or Mongo collection)
//
// Data flow
//
// CreateInfra:
//   Input(tenant/service/infraType/config/meta)
//     -> Registry.Get(infraType)
//       -> Provisioner.Create(input) => ProvisionResult{Endpoint, SecretRef, Status}
//         -> ConnectionStore.Save(ConnectionInfo{...})
//           -> History.Append(Event{action=create, ...})  (optional)
//             -> return ProvisionResult
//
// DestroyInfra:
//   Input(id/infraType/meta)
//     -> Registry.Get(infraType)
//       -> Provisioner.Destroy(input)
//         -> ConnectionStore.Delete(id)
//           -> History.Append(Event{action=destroy, ...}) (optional)
//             -> return nil
//
// Notes
//
// - The Registry is a typed map[InfraType]InfraProvisioner to avoid relying on provisioners exposing Type().
// - ConnectionStore can be backed by Consul KV (fast lookup, centralized config) while History can be stored
//   in MongoDB (queryable audit log) or Consul (simple append-only key scheme).
package infrasmanager
