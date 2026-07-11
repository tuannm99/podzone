# Backbone Flow Refactor

Status: updated 2026-07-11.

This document describes the first runtime flow to stabilize before expanding
Podzone.

## Flow

```text
User
  -> sign in
  -> choose workspace
  -> request or select store
  -> onboarding verifies placement/readiness
  -> Backoffice opens store-scoped workspace
  -> UI calls one protected Backoffice API
```

## SRS Links

- SRS-AUTH-001
- SRS-IAM-001
- SRS-ONB-001
- SRS-ONB-002
- SRS-ONB-003
- SRS-BO-001

## Current Status (2026-07-10)

Investigated from actual code in `internal/`. Status key:
- ✅ Verified working  
- ⚠️ Exists, unverified in Docker  
- ❌ Missing  
- 🔍 Needs investigation

| Capability | Owner | Code Status | Notes |
| --- | --- | --- | --- |
| Validate session / JWT | auth | ⚠️ Exists, unverified | `pkg/pdauthn` verifier used at all service inbound boundaries. Auth gRPC handler in `internal/auth/controller/grpchandler/`. Needs Docker run to verify. |
| Resolve workspace membership | iam | ⚠️ Exists, unverified | gRPC-only. Used by onboarding and backoffice guards via `iam-service:50053`. Bootstrap membership created by `dev-bootstrap` script. Needs verification. |
| List store requests with status | onboarding | ⚠️ Exists, unverified | `GET /requests` returns `collection.Page[*Request]` with `status` field. 15 status values defined in `domain/store/entity`. FE needs to map to pending/failed/ready. |
| Get single store request status | onboarding | ⚠️ Exists, unverified | `GET /requests/:id` — returns full `Request` including `Status`, `LastError`, `StoreID`. IAM workspace guard enforced. |
| Placement allocation + route ready | onboarding | ⚠️ Exists, unverified | `GetTenantPlacementStatus` in `domain/infrasmanager` returns `PlacementStatus{AllocationReady, RouteReady}`. Admin HTTP at `/infras/placements/:tenantId/status`. Not user-facing. |
| **Combined store readiness for FE** | onboarding | ⚠️ Implemented, not wired to FE | `GET /requests/:id/readiness` exists (`internal/onboarding/controller/httphandler/store/controller.go:145`, `Controller.GetStoreReadiness`). Contract documented in `transport-contracts.md` (Slice 0.3) and task spec at `docs/04-sprints/tasks/onboarding-readiness-api.md`. No frontend code calls this endpoint yet, and it is unverified end-to-end in Docker. |
| Resolve tenant DB route (KV) | pdtenantdb + onboarding | ⚠️ Exists, unverified | Route projection published to Redis/Valkey by onboarding worker after provisioning. `pkg/pdtenantdb` reads it. Must exist before backoffice opens. |
| Enforce protected API permission | backoffice → IAM gRPC | ⚠️ Exists, unverified | Backoffice has `authz.go` and IAM gRPC client. GraphQL resolvers gate per tenant. Needs end-to-end test with a ready store. |

## Known Gaps

The following gaps block the backbone from running reliably in Docker dev:

1. **Combined readiness endpoint exists but is not wired to FE**: `GET /requests/:id/readiness` already combines store status with placement allocation/route readiness (see capability table above). The remaining gap is frontend integration and end-to-end verification in Docker, not endpoint design.

2. **Placement bootstrap sequence is unclear**: `dev-bootstrap` seeds a store request and calls onboarding, but whether the full provisioning pipeline (planning → queued → provisioning → ready + placement write + KV projection publish) completes automatically in Docker dev is unverified.

3. **IAM permissions for backoffice guard**: The backoffice graphql guard calls IAM to check permission. Whether the bootstrapped user has the correct permission rows for a fresh store is unverified.

4. **FE store chooser maps 15 statuses to 3 UI states**: The frontend store chooser must map `requested | planning | planned | pending_approval | queued | provisioning` → pending, `failed | failed_retryable | failed_non_retryable | rejected | suspended` → failed, `ready` → ready. This mapping is not yet documented in a UI contract.

5. **MFE remote load errors**: If an MFE remote (backoffice/iam/onboarding) is down, the shell shows an error boundary. But the shell's workspace chooser (in `src/modules/shell/`) is not yet a remote — it runs inline in the host.

## Required Screens

| Screen | Required states |
| --- | --- |
| Login | loading, validation error, auth error, success |
| Workspace/store chooser | loading, empty workspace, pending store, failed store, ready store, permission denied |
| Provisioning status | queued, planning, provisioning, blocked, failed, ready |
| Backoffice home | loading, missing store, placement error, permission denied, success |

## Required Backend Capabilities

| Capability | Owner | Status |
| --- | --- | --- |
| Validate session/token | auth | Existing, verify |
| Resolve workspace membership | IAM/auth projection | Existing, verify |
| List/select stores with readiness state | onboarding/backoffice | Partial |
| Check placement allocation and route projection | onboarding | Partial |
| Resolve tenant DB route | pdtenantdb/onboarding projection | Existing, verify |
| Enforce protected API permission | backoffice -> IAM gRPC | Existing, verify |

## Minimum Contracts To Lock

1. Workspace/store chooser read contract.
2. Store request/provisioning status contract.
3. Placement readiness contract.
4. One protected Backoffice read contract.

## Exit Criteria

- A new user can sign in.
- The first workspace/root ownership is clear.
- Store list distinguishes pending, failed, and ready stores.
- Ready store has resolvable placement.
- Missing placement shows a provisioning/readiness error, not a generic crash.
- Backoffice never opens store-scoped mode without ready placement.
- One protected Backoffice API returns success for authorized user and
  permission detail for unauthorized user.

## First Agent-Ready Task Candidate

The readiness contract and endpoint already exist (see gap #1 above). The
next slice is integration and verification, not contract design.

Candidate:

```text
Wire the workspace/store chooser to GET /requests/:id/readiness and verify
the full backbone flow end to end in Docker dev.
```

Allowed docs:

- `docs/04-sprints/tasks/onboarding-readiness-api.md`
- `docs/03-architecture-detail-design/transport-contracts.md`
- `docs/01-srs/traceability-matrix.md`

Done:

- FE store chooser calls the readiness endpoint instead of inferring from
  store status alone;
- the 15 store-request statuses map to pending/failed/ready per the mapping
  in gap #4 above;
- flow verified end to end in Docker dev (sign in → workspace → ready store
  → Backoffice → one protected API call);
- traceability updated.
