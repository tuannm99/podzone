// Package infrasmanager provides infrastructure provisioning & connection registration for onboarding.
//
// # Overview
//
// The onboarding service needs to:
//  1. provision infrastructure (Mongo/Redis/Postgres/Elastic/Kafka, etc.) in a target environment
//  2. store connection information (endpoint, auth secret reference, metadata)
//  3. keep history/audit of provisioning actions for traceability
//  4. publish commit-coupled side-effect requests through the transactional messaging path
//
// Architecture (high-level)
//
//	Controller / API layer
//	       |
//	       v
//	Interactor (domain)
//	  - validates request
//	  - persists connection state via repository ports
//	  - writes history events via repository ports
//	       |
//	       v
//	Infrastructure adapters
//	  - Connection repository implementations (MongoDB, etc.)
//	  - Transactional event store for placement/connection side effects
//	  - History/event repository (Mongo collection)
//
// # Data flow
//
// CreateInfra:
//
//	Input(tenant/service/infraType/config/meta)
//	  -> state repository persists ConnectionInfo{...}
//	    -> event repository appends Event{action=create, ...}
//	      -> transactional event path enqueues Envelope{type=kv_store.publish, ...}
//	        -> return response
//
// DestroyInfra:
//
//	Input(id/infraType/meta)
//	  -> state repository soft-deletes connection
//	    -> event repository appends Event{action=destroy, ...} (optional)
//	      -> return nil
//
// # Notes
//
//   - Runtime connection projections are stored through KVStore, backed by MongoDB.
//   - Placement/connection publishes are commit-coupled integration events. They can be drained by CDC
//     when available; bounded polling is a fallback relay, not the target scale design.
//   - Best-effort operational jobs should publish directly through a service output port instead of adding
//     outbox storage dependency by default.
package infrasmanager
