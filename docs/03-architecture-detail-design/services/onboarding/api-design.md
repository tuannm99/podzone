# Onboarding Service — API Design

Parent: [Services Index](../README.md) · [Onboarding README](./README.md) · [DB Design](./db-design.md)

## C3: Component View

```mermaid
flowchart TB
    subgraph Controller["controller/httphandler"]
        StoreCtrl["store.Controller"]
        InfrasCtrl["infrasmanager.Controller"]
        Auth["Authentication (JWT + service-token guard)"]
    end

    subgraph Domain["domain"]
        StoreUC["store.StoreInteractor"]
        InfrasUC["infrasmanager.Interactor"]
        ResourceUC["infrasmanager.ResourceInteractor"]
    end

    subgraph Infra["infrastructure"]
        StoreRepo["repository/store (Mongo)"]
        InfrasRepo["repository/infrasmanager (Mongo)"]
        Provisioning["provisioning/router, provisioning/provider"]
        Worker["messaging/worker.StoreProvisioningWorker"]
        BOClient["backofficeclient.StoreFinalizer"]
    end

    StoreCtrl --> Auth
    InfrasCtrl --> Auth
    StoreCtrl --> StoreUC
    InfrasCtrl --> InfrasUC
    InfrasCtrl --> ResourceUC
    StoreUC --> StoreRepo
    StoreUC -->|"gRPC"| IAM["IAM Service"]
    InfrasUC --> InfrasRepo
    ResourceUC --> InfrasRepo
    Worker --> StoreUC
    Worker --> Provisioning
    Provisioning -->|"publish route"| KV["Redis/Valkey"]
    Worker --> BOClient
    BOClient -->|"HTTP, internal service token"| BO["Backoffice Service"]
```

Matches the module boundaries in
[`01-modules.md`](../../01-modules.md) and the layering rule in
[`03-ddd-clean-architecture.md`](../../03-ddd-clean-architecture.md):
controllers call usecases only, usecases depend on ports
(`domain/store/inputport`, `domain/infrasmanager/*`), infrastructure
implements those ports.

## HTTP API Surface

All routes are also mirrored under a `legacy` prefix with identical
handlers (kept for backward compatibility — do not add new endpoints only
to the legacy group). Table below lists the canonical path only.

### Store Requests (`/requests`)

| Method + Path | Request | Response | Errors | Notes |
|---|---|---|---|---|
| `POST /requests` | `{name, subdomain, owner_id}` | `201` + `Request` | `400` invalid body | `CreateStoreRequest` |
| `GET /requests` | query: `collection.*` (pagination/filter) | `200` `{items: []Request, pageInfo}` | — | workspace-scoped |
| `GET /requests/:id` | — | `200` `Request` | `404` not found/workspace mismatch | |
| `GET /requests/:id/readiness` | — | `200` `{request_id, request_status, readiness, failure_reason, ui_state}` | `404`, `401` | Full contract: [`05-transport-contracts.md`](../../05-transport-contracts.md) Slice 0.3, [PZEP-0001](../../../09-pzep/PZEP-0001-onboarding-store-readiness-endpoint.md) |
| `GET /requests/:id/transitions` | query: `collection.*` | `200` `{items: []RequestTransition, pageInfo}` | — | Audit trail of status changes |
| `POST /requests/:id/retry` | — | `204` | error mapped via `writeStoreError` | Requeues a failed/blocked request |
| `POST /requests/:id/approve` | — | `204` | requires `authorizeApproval` | For `pending_approval` requests |
| `POST /requests/:id/reject` | — | `204` | requires `authorizeApproval` | |

`Controller.UpdateStoreRequestStatus` (`{status}` body,
`StoreInteractor.UpdateStoreRequestStatus`) is implemented but **not
registered on any route** — verified via full-repo grep, only referenced
internally by `ApproveStoreRequest`/`RejectStoreRequest`. Dead HTTP
handler; do not assume it's reachable.

### Infrastructure Admin (`/infras`, `requireInfrastructureRead`/`Manage` guarded)

| Method + Path | Purpose |
|---|---|
| `GET /infras/connections` | List backing-service connections (`infra_type`: mongo/redis/postgres/elasticsearch/kafka) |
| `GET /infras/connections/:infraType/:name` | Get one connection |
| `POST /infras/connections` | Upsert connection |
| `DELETE /infras/connections/:infraType/:name` | Delete connection |
| `GET /infras/events` | List connection health-check events |
| `GET /infras/placements/:tenantId/status` | `GetTenantPlacementStatus` — `{allocation_ready, route_ready}`, admin-only, not the same as the user-facing readiness endpoint above |
| `POST /infras/placements/:tenantId/reconcile` | `ReconcileTenantPlacement` — re-derive and republish the KV route projection from `placement_allocations` when it's out of sync |
| `GET`/`PUT`/`DELETE /infras/resources/database-clusters/:name` | DB cluster inventory (capacity, health) |
| `GET`/`PUT`/`DELETE /infras/resources/kubernetes-clusters/:name` | K8s cluster inventory |
| `GET`/`PUT`/`DELETE /infras/resources/runtime-pools/:name` | Runtime pool inventory |

None of the `/infras/*` surface is called by the frontend today — see
[README.md](./README.md) "Frontend Surface", `AdminProvisioningPage.tsx`
covers only a subset via its own panels, not confirmed to hit every route
above.

## C4: Sequences Per Usecase

### Onboarding Placement Publish to Runtime KV

```mermaid
sequenceDiagram
    participant Client as Operator / Provisioning Worker
    participant Onboarding as Onboarding Service
    participant Mongo as Mongo Onboarding Store
    participant EventRelay as Onboarding CDC / Fallback Relay
    participant RuntimeKV as Redis/Valkey Route Projection

    Client->>Onboarding: save placement allocation and router publish record
    Onboarding->>Mongo: persist allocation as source of truth
    Onboarding->>Mongo: append publish record for router projection
    Onboarding-->>Client: accepted

    Mongo->>EventRelay: stream publish record or bounded fallback read
    EventRelay->>RuntimeKV: upsert placement projection for pdtenantdb
    EventRelay->>Mongo: mark published when using fallback relay
```

Label corrected from the source doc this was moved from
(`docs/02-architecture-overall/05-sequences.md`, now deleted) — "Mongo
runtime_kv" was wrong, it's Redis/Valkey. See
[db-design.md](./db-design.md) "Not A Database Table: KV Route Projection".

### Store Onboarding Request to Ready

```mermaid
sequenceDiagram
    participant UI as Admin / Workspace Owner UI
    participant Onboarding as Onboarding Service
    participant IAM as IAM Service
    participant Queue as Provisioning Queue (leased claim)
    participant Worker as StoreProvisioningWorker
    participant Provider as Placement Provider
    participant Mongo as Mongo Onboarding Store
    participant RuntimeKV as Redis/Valkey Route Projection
    participant BO as Backoffice Runtime
    participant DB as Tenant DB / Schema
    participant PDT as pdtenantdb

    UI->>Onboarding: POST /requests (name, subdomain, owner_id)
    Onboarding->>Mongo: insert store_requests (status=requested)
    Onboarding-->>UI: 201 created

    loop Worker tick
        Worker->>Onboarding: ProcessNextStoreRequest (claim next queued/planning, lease)
        Onboarding->>IAM: verify requester membership / approval actor
        IAM-->>Onboarding: membership + policy context
        Worker->>Provider: plan runtime placement
        Provider-->>Worker: cluster, db, schema, metadata
        Worker->>Mongo: write placement_plans, placement_allocations
        Worker->>Onboarding: FinalizeNextStoreRequest
        Onboarding->>Mongo: append router publish record
        Mongo->>RuntimeKV: publish tenant placement projection
        Onboarding->>Mongo: update store_requests.status = ready
    end

    BO->>PDT: resolve placement for tenant/store scope
    PDT->>RuntimeKV: get tenant placement
    RuntimeKV-->>PDT: cluster/db/schema placement
    BO->>DB: execute store-scoped operations only after placement resolves
```

Label corrected the same way as above. This is the highest-risk flow in
the system per
[`backbone-flow-refactor.md`](../../../06-recovery/backbone-flow-refactor.md)
and is **unverified end-to-end in Docker dev** — see
[`docs/STATUS_CURRENT.md`](../../../STATUS_CURRENT.md). The worker-loop
detail (`ProcessNextStoreRequest`/`FinalizeNextStoreRequest`, lease-based
claim) was added here from `domain/store/interactor.go` — the original
diagram in the deleted `05-sequences.md` described this at a higher level
via a generic "Worker" participant with a "request placement plan"
round-trip; actual code combines claim+process+finalize as three
interactor calls around one worker tick, not a plan-then-provision
round-trip through the same RPC. Adjusted accordingly.

### Reconcile Tenant Placement (repair path)

Not in the original sequence set — added because
[db-design.md](./db-design.md) "Failure Modes" names this as the recovery
path when the KV publish fails after the Mongo write succeeds.

```mermaid
sequenceDiagram
    participant Admin as Operator (admin UI/CLI)
    participant Onboarding as Onboarding Service
    participant Mongo as Mongo Onboarding Store
    participant RuntimeKV as Redis/Valkey Route Projection

    Admin->>Onboarding: POST /infras/placements/:tenantId/reconcile
    Onboarding->>Mongo: read latest ready placement_allocation for tenant
    Onboarding->>Onboarding: re-derive route projection from allocation
    Onboarding->>RuntimeKV: republish route projection
    Onboarding-->>Admin: 200 (reconciled status)
```

## Cross-Service Dependencies

**Onboarding calls:**
- IAM (gRPC) — workspace membership/approval-actor checks in
  `CreateStoreRequest`/`ProcessNextStoreRequest` and `authorizeApproval`/
  `authorizeRead`.
- Backoffice (HTTP, internal service token) — `backofficeclient.
  StoreFinalizer` calls back into Backoffice when finalizing a store;
  see `BACKOFFICE_INTERNAL_SERVICE_TOKEN` in
  `deployments/docker/services.yml`. Not covered by an existing sequence
  above — out of scope for this pass, flagged for a future addition.

**Calls onboarding:**
- Frontend (`frontend/apps/onboarding`) via APISIX — store request CRUD,
  readiness (once wired, see PZEP-0001 open question).
- `authentication.go`'s service-token path accepts calls authenticated
  with `ONBOARDING_SERVICE_TOKEN`/`BACKOFFICE_INTERNAL_SERVICE_TOKEN`
  instead of a user JWT — used for service-to-service and
  `dev-bootstrap` seeding, not by end users.

## Links Back To Delivery

- [Onboarding README](./README.md)
- [DB Design](./db-design.md)
- [Backbone Flow Refactor](../../../06-recovery/backbone-flow-refactor.md)
- [PZEP-0001](../../../09-pzep/PZEP-0001-onboarding-store-readiness-endpoint.md)
