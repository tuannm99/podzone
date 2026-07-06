# Store Onboarding Pipeline

## Purpose

This document defines the business rules for creating and activating a store inside a tenant workspace.

The goal is to remove ambiguity around:

- when a store is considered created
- when it is usable in Backoffice
- when DB placement exists
- when admin approval is required
- which service owns the provisioning pipeline

This is a requirements document, not an implementation plan.

## Business framing

### Workspace

A workspace is the merchant's tenant-level account.

It owns:

- memberships
- permissions
- one or more stores
- onboarding requests for infrastructure

### Store

A store is an operational unit inside a workspace.

A store:

- belongs to exactly one workspace
- is not usable until onboarding has completed
- requires a resolved runtime/data placement before store-scoped operations can run

### Store creation request

Creating a store is not a single instant action.

It is a request that may require:

- capacity checks
- database or schema allocation
- connection selection
- placement publication
- approval

The product must distinguish between:

- `store requested`
- `request queued`
- `pending approval`
- `store being provisioned`
- `store ready`
- `store active`

## Core resource model

Onboarding must treat platform resources as finite inventory, not as unlimited shared defaults.

The resource model must distinguish:

- tenant workspace
- store
- database cluster
- database or schema placement
- Kubernetes cluster
- Kubernetes namespace
- application runtime pool
- connection route
- secret reference
- resource quota
- placement policy

The Mongo-backed routing KV store is not the resource inventory.

The source of truth for resource inventory and placement allocation belongs to the onboarding/platform resource manager.
Routing KV entries are projections that can be rebuilt.

## Functional requirements

### FR-ONB-001 Store request intake

The system must let an authorized workspace owner or operator submit a store onboarding request.

The request must capture:

- workspace ID
- requested store name and slug
- requester identity (`requested_by`), representing the actor that submits the request
- store owner identity (`owner_id`), representing the principal that owns the resulting store
- requested environment
- requested isolation level, if provided
- requested runtime preferences, if provided
- requested data placement preferences, if provided

The request must be persisted before any long-running provisioning work starts.

For a normal tenant-owner request, `owner_id` defaults to `requested_by`. Setting a different owner is an
administrative operation and requires explicit authorization. Provisioning workers and migration service accounts must
remain the requester and must never become the store owner implicitly.

### FR-ONB-002 Request lifecycle tracking

The system must track the lifecycle of every onboarding request from submission to terminal state.

Minimum lifecycle states:

- `requested`
- `validating`
- `planning`
- `pending_approval`
- `approved`
- `provisioning`
- `ready`
- `failed`
- `cancelled`

Each state transition must record:

- actor or system component
- timestamp
- reason
- correlation ID
- previous state
- next state

### FR-ONB-003 Capacity discovery

Before selecting placement, onboarding must discover current capacity for:

- database clusters
- database instances or databases
- schema count per placement database
- database connection limits
- Kubernetes clusters
- Kubernetes namespaces
- namespace CPU and memory quota
- application runtime pools
- existing tenant and store assignments

Capacity data may come from a resource manager database, Kubernetes API, cloud API, or local Docker provider,
depending on environment.

### FR-ONB-004 Placement planning

Onboarding must produce a placement plan before provisioning.

The placement plan must answer:

- which DB cluster will host tenant data
- which database or schema will host tenant data
- whether the tenant uses shared schema mode, database-per-tenant, or cluster-per-tenant
- which Kubernetes cluster will host runtime resources
- which namespace will host runtime resources
- whether an existing namespace can be reused
- whether a new namespace must be created
- which application runtime pool will serve the tenant/store
- which secret reference will be used for runtime connection credentials

The plan must be persisted and auditable before execution.

### FR-ONB-005 Placement policy evaluation

Onboarding must evaluate placement policy before approval or provisioning.

Policies must be able to express:

- max stores per workspace
- max tenants per DB cluster
- max schemas per placement database
- max connections per DB cluster
- max tenants per Kubernetes namespace
- CPU and memory quota thresholds
- reserved capacity for platform or VIP tenants
- isolation level requirements
- environment-specific rules for local, staging, and production

### FR-ONB-006 Approval routing

Onboarding must support both automatic approval and manual platform approval.

Automatic approval is allowed only when:

- requester has permission
- workspace quota passes
- placement policy passes
- selected DB target has capacity
- selected Kubernetes target has capacity
- requested isolation level is allowed
- no high-risk or expensive resource creation is required

Manual approval is required when policy cannot auto-approve the plan.

### FR-ONB-007 Provisioning execution

After approval, onboarding must execute the placement plan through the selected provider.

The provider must support environment-specific execution:

- local Docker for development
- Kubernetes for shared and isolated runtime placement
- Terraform or cloud provider integration in the future

Provisioning must create or bind:

- tenant database or schema
- DB role or connection identity, when required
- Kubernetes namespace, when required
- Kubernetes ResourceQuota and LimitRange, when required
- app runtime assignment
- secret reference
- router projection payload

### FR-ONB-008 Runtime route publication

After provisioning succeeds, onboarding must publish runtime placement metadata for tenant resolution.

The published route must include:

- tenant ID
- store ID when store-specific routing is required
- cluster name
- placement mode
- database name
- schema name or database target
- runtime pool
- Kubernetes cluster and namespace, when applicable
- secret reference
- version

The route must not contain plaintext DB passwords.

### FR-ONB-009 Readiness verification

A store can become selectable only after onboarding verifies:

- placement allocation exists
- DB target is reachable
- schema or database exists
- required migrations or bootstrap data have completed
- runtime namespace or runtime pool is available
- route projection has been published
- Backoffice can resolve the placement

### FR-ONB-010 Failure and retry handling

Onboarding must persist provisioning failures with enough detail for operators to act.

The system must support:

- retryable failure classification
- non-retryable failure classification
- retry count
- last error message
- failed step
- manual retry
- cancellation before provisioning
- operator-visible failure reason

### FR-ONB-011 Reconciliation

Onboarding must support reconciliation between source-of-truth placement allocation and runtime projections.

Reconciliation must detect:

- missing runtime KV entry
- stale router version
- missing DB schema or database
- missing Kubernetes namespace
- quota drift
- placement marked ready while runtime checks fail
- IAM tenants that have no onboarding request or ready placement allocation

Legacy tenants must be enrolled through the normal onboarding command path with an explicit owner. Reconciliation must
not create placement projections directly because a routing entry without a request, allocation, and provisioned
resource is not a valid placement.

### FR-ONB-012 Store selection integration

Backoffice must not show a store as openable until onboarding marks it ready and placement resolution succeeds.

Backoffice may show pending or failed store requests, but those entries must be clearly labelled and must not enter
store-scoped workspace mode.

## Non-functional requirements

### NFR-ONB-001 Source-of-truth integrity

Placement allocation must have one source of truth.

Runtime KV systems are projections only.

### NFR-ONB-002 Capacity safety

Onboarding must fail closed when capacity cannot be determined.

The system must not place a new tenant into a DB cluster, namespace, or runtime pool when capacity state is unknown.

### NFR-ONB-003 Idempotency

Request processing and provisioning steps must be idempotent.

Retrying the same request must not create duplicate stores, duplicate schemas, duplicate namespaces, or conflicting
route entries.

### NFR-ONB-004 Auditability

Every approval, placement decision, provisioning action, failure, retry, and readiness transition must be auditable.

### NFR-ONB-005 Security

Routing projections must not contain plaintext credentials.

Secrets must be represented as references and resolved through the appropriate secret provider at runtime.

### NFR-ONB-006 Availability

Tenant placement resolution must tolerate projection-store outages for already resolved tenants where a valid local cache
exists.

New tenant provisioning must not rely on stale routing projection data as source of truth.

### NFR-ONB-007 Observability

Onboarding must expose operational visibility for:

- queue depth
- request state distribution
- approval backlog
- provisioning duration
- failure rate by provider
- capacity usage by DB cluster
- capacity usage by Kubernetes cluster and namespace
- route projection lag

### NFR-ONB-008 Environment portability

The same onboarding domain flow must work across local Docker, Kubernetes, and future Terraform/cloud providers.

Provider differences must stay behind provider-specific adapters.

### NFR-ONB-009 Bounded resource growth

The system must prevent unbounded tenant growth in a single DB cluster, namespace, runtime pool, or application
deployment.

Placement policy must be configurable per environment.

### NFR-ONB-010 Operator recovery

Operators must be able to inspect, retry, cancel, or manually approve blocked requests without direct database edits.

## Request tracking model

Every store provisioning request must be traceable.

Minimum fields:

- request ID
- workspace ID
- requested store name and slug
- requested by actor
- request timestamp
- current processing stage
- approval state
- selected connection or target pool
- selected placement mode
- last error, if any
- retry count

The UI and admin surfaces must be able to show:

- pending request
- approved but not yet provisioned
- provisioning in progress
- failed with reason
- ready to open

The user must never be left with an unlabelled failure state.

## Required workflow

### Store creation

1. A workspace owner or authorized operator submits a store creation request.
2. The request is persisted as a trackable onboarding item.
3. The system evaluates whether the request can be auto-approved.
4. If approval is required, the request enters a pending approval queue.
5. If approved, onboarding builds a placement allocation from the infrastructure manager.
6. Onboarding provisions the required runtime through the selected provider and persists the allocation.
7. A background worker or queue consumer performs the long-running provisioning work.
8. Onboarding publishes the Mongo runtime KV projection consumed by `pdtenantdb`.
9. Onboarding finalizes the Backoffice store with the request's explicit `owner_id`.
10. Backoffice can open the store only after placement is resolvable and the store is ready.

### Approval paths

The system must support both:

- automatic approval
- admin approval

Automatic approval is allowed only when policy, quota, and infrastructure checks pass.

Admin approval is required when:

- tenant quota is exceeded
- database capacity is constrained
- a higher-risk configuration is requested
- platform policy requires manual review

Approval must be recorded as part of the request lifecycle, not as an implicit side effect.

## Provisioning checks

The onboarding pipeline must validate:

- database capacity
- cluster selection
- schema or database allocation mode
- connection availability
- tenant placement readiness
- store readiness state after publish
- connection health before accepting the request
- whether the request can be placed on an existing tenant pool or needs a new one
- whether the requested store violates tenant quotas or platform policy

The product should not assume a store can be opened just because the tenant exists.

The product should not assume a store can be provisioned synchronously from the UI request path.

## Placement requirements

A store becomes operational only when the tenant has a resolvable placement.

The placement source of truth is the onboarding/platform infrastructure manager allocation.
Mongo `runtime_kv` is only a router projection for runtime lookup and can be rebuilt from that allocation.

Requirements:

- placement allocation must be persisted before publishing router metadata
- each tenant has at most one canonical ready placement allocation; additional stores reuse that allocation
- store and request IDs on an allocation are provisioning provenance, not placement identity
- placement provider must be explicit: local Docker, Kubernetes, Terraform/cloud, or another runtime
- connection endpoint and secret reference must be produced by the selected provider after provisioning
- service config may seed runtime policy or local defaults, but it must not be treated as the source of truth for connection routing
- placement must exist before store-scoped runtime can be entered
- missing placement should surface as a provisioning or readiness problem, not as a generic UI failure
- placement publication is part of onboarding, not a backoffice-side fallback or source of truth
- placement resolution must point to the selected tenant DB, schema, or equivalent storage target
- placement must be refreshed or republished when the storage target changes

## Store lifecycle

The product should treat store lifecycle as:

- `requested`
- `queued`
- `pending_approval`
- `provisioning`
- `failed`
- `ready`
- `active`
- `suspended`
- `archived`

The exact technical status names can differ, but the product must preserve the lifecycle meaning.

## Visibility rules

- Requested stores may appear in admin surfaces as pending items.
- A store that is not ready must not appear as selectable for store-scoped operations.
- Backoffice should only open stores that are ready and resolvable.
- Admin surfaces must expose the reason a store is blocked, failed, or waiting.
- A store request may be visible even if the store itself is not yet selectable.

## Provisioning administration

The onboarding administration surface must expose the same operational facts used by the planner and workers:

- a paginated, searchable pipeline view backed by persisted request transitions
- explicit stages for request, approval, planning, provisioning, route publication, store finalization, and readiness
- CRUD for global database clusters, Kubernetes clusters/namespaces, and runtime pools
- tenant-scoped CRUD for connection routes and secret references
- resource configuration editing with validation; plaintext credentials must never be returned to or stored by the UI
- IAM guards on every read and mutation API; hiding a frontend control is not authorization

Resource deletion archives capacity so historical provisioning facts remain explainable. Archived or unhealthy resources
must not be selected for new placement.

## Responsibilities by service

### IAM

IAM owns:

- workspace creation
- membership and permission assignment
- approval actor identity
- approval policy metadata if policy is defined at the IAM layer

IAM does not own placement publication.
IAM does not execute the long-running provisioning steps.

### Onboarding

Onboarding owns:

- store provisioning requests
- queueing or dispatching the provisioning work
- connection and placement publication
- infra readiness tracking
- approval-driven provisioning flow
- resource checks and placement selection

For the local Docker and Kubernetes shared-Postgres strategy, onboarding provisions tenant backoffice schemas in
the dedicated placement database `podzone_tenants`. The default `postgres` database is treated as an admin/default
database only; service tables must live in service-owned databases such as `auth`, `iam`, and `partner`.

### Queue or worker runtime

A queue consumer or worker may execute the actual provisioning steps.

That runtime owns:

- polling or consuming pending requests
- calling the appropriate infra services
- recording success or failure
- republishing placement after successful changes

### Backoffice

Backoffice owns:

- displaying store readiness
- opening only ready stores
- reporting provisioning/readiness state to the user

## Product requirement summary

The user flow should feel like this:

1. Sign in.
2. Choose workspace.
3. Request or select a store.
4. If the store is pending, wait for onboarding or approval.
5. If the store is ready, enter Backoffice.

The user should never be forced to infer whether a store exists, whether placement exists, or whether the store is usable.
