// Package infrasmanager provides infrastructure provisioning & connection registration for onboarding.
//
// # Overview
//
// The onboarding service needs to:
//  1. provision infrastructure (Mongo/Redis/Postgres/Elastic/Kafka, etc.) in a target environment
//  2. store connection information (endpoint, auth secret reference, metadata)
//  3. keep history/audit of provisioning actions for traceability
//  4. publish side-effect requests through Kafka-backed outbox relay instead of writing directly to Consul
//
// Architecture (high-level)
//
//	Controller / API layer
//	       |
//	       v
//	InfraManager (core)
//	  - validates request
//	  - delegates to a provisioner by InfraType via Registry
//	  - persists connection state via repository ports
//	  - writes history events via repository ports
//	       |
//	       v
//	Providers
//	  - InfraProvisioner implementations (K8s/Helm/Terraform/CLI, etc.)
//	  - Connection repository implementations (MongoDB, etc.)
//	  - Outbox adapter for Kafka relay
//	  - History/event repository (Mongo collection)
//
// # Data flow
//
// CreateInfra:
//
//	Input(tenant/service/infraType/config/meta)
//	  -> Registry.Get(infraType)
//	    -> Provisioner.Create(input) => ProvisionResult{Endpoint, SecretRef, Status}
//	      -> state repository persists ConnectionInfo{...}
//	        -> event repository appends Event{action=create, ...}
//	          -> outbox repository enqueues Envelope{type=consul.publish, ...}
//	          -> return ProvisionResult
//
// DestroyInfra:
//
//	Input(id/infraType/meta)
//	  -> Registry.Get(infraType)
//	    -> Provisioner.Destroy(input)
//	      -> state repository soft-deletes connection
//	        -> event repository appends Event{action=destroy, ...} (optional)
//	          -> return nil
//
// # Notes
//
//   - The Registry is a typed map[InfraType]InfraProvisioner to avoid relying on provisioners exposing Type().
//   - Connection state can be backed by MongoDB or Consul KV, while history is queryable storage and outbox
//     is relayed to Kafka.
package infrasmanager
