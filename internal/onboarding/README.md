# Onboarding Implementation Guide

## Purpose

Onboarding owns the store provisioning pipeline and the source-of-truth placement allocation for tenant runtime/data
resources.

The service must not treat tenant placement as a static config lookup. A new tenant or store must go through a placement
decision before it can become usable in Backoffice.

## Implementation priority

Build the placement backbone before expanding UI or secret-routing work:

1. resource inventory
2. capacity checks
3. placement policy evaluation
4. placement planning
5. provisioning execution
6. placement allocation persistence
7. route projection publication
8. readiness verification

Mongo `runtime_kv` is only a router projection. It is not the placement source of truth.

Resource inventory must be declared before provisioning. The planner must select from persisted inventory, not invent a
DB cluster, namespace, or runtime pool from request data.

## Clean architecture shape

Follow the service-local clean architecture rules from `agent/SKILL.md`.

Recommended package ownership:

- `domain/store`
  - store request lifecycle
  - approval state
  - user-facing onboarding status
  - orchestration into infrastructure placement usecases
- `domain/infrasmanager`
  - resource inventory concepts
  - capacity snapshots
  - placement policy contracts
  - placement plan and allocation contracts
  - provisioning workflow usecases
- `domain/*/inputport`
  - command/query usecase contracts consumed by controllers and workers
- `domain/*/outputport`
  - repository, planner, policy, provisioner, route writer, and runtime-check ports
- `controller/httphandler`
  - request parsing, response mapping, auth context extraction
  - no placement business rules
- `controller/eventhandler`
  - inbound event handling and route projection handlers
- `infrastructure/repository`
  - Mongo-backed request, inventory, plan, allocation, outbox, and audit persistence
- `infrastructure/provisioning`
  - Docker, Kubernetes, and future Terraform/cloud provider adapters
- `infrastructure/messaging`
  - worker runtime, CDC/fallback outbox relay, queue consumers

Controllers must not call provider adapters directly. Interactors orchestrate through output ports.

## Placement backbone

The first implementation slice should introduce explicit models and ports for finite resources.

Core concepts:

- `ResourceInventory`
- `DatabaseCluster`
- `PlacementDatabase`
- `KubernetesCluster`
- `KubernetesNamespace`
- `RuntimePool`
- `CapacitySnapshot`
- `PlacementPolicy`
- `PlacementPlan`
- `PlacementAllocation`
- `ReadinessCheck`

Core output ports:

- `ResourceInventoryRepository`
- `CapacityChecker`
- `PlacementPolicyEvaluator`
- `PlacementPlanner`
- `PlacementPlanRepository`
- `PlacementAllocationRepository`
- `StorageProvisioner`
- `RuntimeProvisioner`
- `PlacementRouteWriter`
- `ReadinessChecker`

The planner must fail closed when inventory or capacity is unknown.

## Resource inventory persistence

The first inventory source-of-truth lives in Mongo.

Initial collections:

- `resource_db_clusters`
- `resource_k8s_clusters`
- `resource_runtime_pools`

Provisioning must read these collections before creating any tenant placement.

Config can seed a local/dev inventory record when the collections are empty, but config is not the runtime source of
truth once inventory exists in DB.

If inventory is partially configured, onboarding must not silently fill the missing pieces from config. Operators must
fix the inventory declaration.

## Required flow

The target flow is:

```text
Create store request
-> persist request
-> validate permissions and workspace quota
-> load resource inventory
-> calculate capacity snapshot
-> evaluate placement policy
-> create placement plan
-> approve automatically or wait for platform approval
-> provision DB/schema/runtime namespace through provider adapters
-> persist placement allocation
-> publish route projection
-> verify readiness
-> mark store ready
```

No store should become selectable until placement allocation exists and runtime resolution succeeds.

## Store request lifecycle

Store request processing must be visible through explicit state transitions.

Current lifecycle states:

- `requested`
- `queued`
- `planning`
- `planned`
- `pending_approval`
- `provisioning`
- `ready`
- `pending_platform_setup`
- `failed_retryable`
- `failed_non_retryable`
- `cancelled`

Every transition should be appended to `store_request_transitions` with actor, reason, error code, and timestamp.

Provider or platform setup gaps, such as unavailable Kubernetes/Terraform execution adapters, should move the request to
`pending_platform_setup` instead of pretending the store is ready or returning an unlabeled failure.

## Provider boundaries

Provider adapters are implementation details behind domain-owned output ports.

Supported provider directions:

- local Docker
  - create or bind local placement database/schema
  - seed development route projection
- Kubernetes
  - select or create namespace
  - apply ResourceQuota and LimitRange
  - bind runtime pool and secret reference
  - currently planning-only until a real Kubernetes adapter and test cluster exist
- Terraform/cloud
  - future adapter
  - must implement the same domain ports
  - currently planning-only and must fail clearly during provisioning

Provider-specific SDK clients must stay under `infrastructure/provisioning`.

## Resource placement rules

Placement planning must consider at least:

- max tenants per DB cluster
- max schemas per placement database
- DB connection limit
- DB health
- max tenants per Kubernetes namespace
- namespace CPU and memory quota
- runtime pool capacity
- requested isolation level
- reserved capacity
- environment policy

The planner must not place new tenants into a shared database, namespace, or runtime pool indefinitely.

## Persistence rules

Persist these as source-of-truth onboarding state:

- store request
- request state transitions
- resource inventory snapshot or inventory references
- placement plan
- approval decision
- placement allocation
- readiness result
- outbox or CDC-backed projection event

The Mongo runtime KV projection can be rebuilt from placement allocation state.

## Secrets and routes

Route projection payloads must not contain plaintext credentials.

Routes should contain a `secret_ref` and enough non-secret placement metadata for runtime resolution:

- tenant ID
- store ID when needed
- cluster name
- placement mode
- database name
- schema name or database target
- Kubernetes cluster and namespace when applicable
- runtime pool
- route version
- secret reference

Secret resolution belongs to a separate runtime concern and should not be mixed into placement planning.

## Testing expectations

Add domain/interactor tests before provider-specific tests.

Minimum planner tests:

- selects DB and namespace when capacity exists
- fails when DB capacity is unknown
- fails when namespace capacity is unknown
- rejects placement when policy threshold is exceeded
- requires approval for high-risk or expensive placement
- keeps retry/idempotency from creating duplicate allocations

Provider tests should verify adapter behavior, not business placement rules.

## Current migration note

The current local Docker and Kubernetes shared-Postgres strategy places tenant schemas in `podzone_tenants`.

The `postgres` database is only an admin/default connection database and must not become a tenant placement target or
service-owned public schema again.
