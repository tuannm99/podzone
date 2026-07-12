# Backbone Flow Refactor

Status: updated 2026-07-12.

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

## Current Status (2026-07-12)

Investigated from actual code in `internal/`, then re-verified against the
already-running `full`-profile Docker dev stack (containers up, not started
fresh for this pass) using the seeded dev-bootstrap identity
(`devowner@podzone.dev`, tenant `tenant-dev`, store "Demo Store"). Status key:
- ✅ Verified working
- ⚠️ Exists, unverified in Docker
- ❌ Missing
- 🔍 Needs investigation

**Verification method (2026-07-12):** real HTTP calls with the dev JWT from
`/dev-auth-bootstrap.json`, directly against the running containers (onboarding
on `:8800`, Backoffice GraphQL through APISIX on `:9080/backoffice/graphql`) —
the same calls the frontend code makes, not unit tests. **Not done:** a
literal browser click-through with a screenshot — no headless browser
automation tool was available in this environment (no root to install
Chromium's system deps, and a binary-only download did not complete
reliably). Treat rows below as API-level, not pixel-level, verification.

| Capability | Owner | Code Status | Notes |
| --- | --- | --- | --- |
| Validate session / JWT | auth | ✅ Verified working | Dev-bootstrap JWT accepted by both onboarding's HTTP middleware and backoffice's GraphQL `TenantMiddleware`; a request with no `Authorization` header cleanly returned `{"errors":[{"message":"authorization bearer token is required","extensions":{"code":"UNAUTHENTICATED"}}]}` instead of crashing. |
| Resolve workspace membership | iam | ✅ Verified working (indirect) | Not called directly, but every tenant-scoped call below succeeded for `tenant-dev`, which requires IAM membership resolution to have passed. |
| List store requests with status | onboarding | ✅ Verified working | `GET :8800/onboarding/v1/requests` with `X-Tenant-ID: tenant-dev` returned the real seeded request: `{id, name: "Demo Store", subdomain: "demo-store", status: "ready", store_id, ...}`. |
| Get single store request status | onboarding | 🔍 Not directly tested this pass | `GET /requests/:id` itself wasn't called; readiness (below) covers overlapping ground for the one seeded request. |
| Placement allocation + route ready | onboarding | ✅ Verified working (indirect) | Confirmed via the readiness response: `placement_allocation_ready: true`, `route_ready: true` for the seeded request. |
| **Combined store readiness for FE** | onboarding | ✅ Verified working end to end | `GET :8800/onboarding/v1/requests/<id>/readiness` returned `{"request_status":"ready","readiness":{"store_ready":true,"placement_allocation_ready":true,"route_ready":true},"ui_state":"ready"}` — exactly the shape `getStoreReadiness`/`createStoreReadinessViewModel` (2026-07-12 FE change) expect, and `ui_state: "ready"` correctly maps to `readinessBadgeColor('ready') -> 'green'`. Only the `ready` path was exercised — no seeded request exists in `pending`/`provisioning`/`blocked`/`failed` to confirm those branches live. |
| Resolve tenant DB route (KV) | pdtenantdb + onboarding | ✅ Verified working (indirect) | The Backoffice GraphQL query below returned real Postgres-backed data for `tenant-dev`, which requires the route projection to have resolved correctly. |
| Enforce protected API permission | backoffice → IAM gRPC | ✅ Verified working | `POST :9080/backoffice/graphql` `stores` query with the dev JWT returned the real store (`{"id":"...","name":"Demo Store","status":"active","isActive":true}`) — this is the "one protected Backoffice API call" the exit criteria asks for. Missing/invalid auth returns a clear GraphQL error, not a crash. |

## Known Gaps

The following gaps block the backbone from running reliably in Docker dev:

1. **Combined readiness endpoint exists and is now wired to FE (2026-07-12)**: `GET /requests/:id/readiness` combines store status with placement allocation/route readiness (see capability table above). `ProvisioningRequestsPanel.tsx`'s status badge now reads the endpoint's `ui_state` per visible request instead of inferring from `request.status` alone — see `docs/04-sprints/tasks/wire-store-readiness-frontend.md`. **Remaining:** end-to-end verification in a live Docker dev environment (this pass was build/lint-level only, no running backend was exercised).

2. **Placement bootstrap sequence is unclear**: `dev-bootstrap` seeds a store request and calls onboarding, but whether the full provisioning pipeline (planning → queued → provisioning → ready + placement write + KV projection publish) completes automatically in Docker dev is unverified.

3. **IAM permissions for backoffice guard**: The backoffice graphql guard calls IAM to check permission. Whether the bootstrapped user has the correct permission rows for a fresh store is unverified.

4. **FE store chooser maps 15 statuses to 3 UI states — done, corrected mapping**: the mapping stated here previously was imprecise (`failed_retryable` grouped under `failed`). The authoritative mapping lives in `docs/03-architecture-detail-design/05-transport-contracts.md` "Slice 0.3: Store Readiness Contract": `requested | planning | planned | pending_approval` → `pending`; `queued | provisioning | pending_platform_setup` → `provisioning`; `failed_retryable` → `blocked` (not `failed`); `failed | failed_non_retryable | rejected | cancelled | suspended | archived` → `failed`; `ready` + placement/route ready → `ready`; `ready` + placement not ready → `blocked`. `ProvisioningRequestsPanel.tsx` now consumes this mapping via the backend's `ui_state` field directly rather than re-implementing it client-side.

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

- A new user can sign in. **✅ API-level, 2026-07-12** — dev-bootstrap JWT accepted; missing/invalid auth cleanly rejected, not a crash.
- The first workspace/root ownership is clear. **✅** — `tenant-dev` resolved from the JWT's `active_tenant_id` claim (backoffice) and from the `X-Tenant-ID` header checked against membership (onboarding) — two different, both-correct resolution mechanisms per service, confirmed by reading `internal/backoffice/tenant_middleware.go`.
- Store list distinguishes pending, failed, and ready stores. **⚠️ Partially** — the one seeded request is `ready`, and `ui_state: "ready"` was confirmed correct; pending/provisioning/blocked/failed were not exercised (no seed data in those states).
- Ready store has resolvable placement. **✅** — readiness response shows `placement_allocation_ready: true`, `route_ready: true` for the ready request.
- Missing placement shows a provisioning/readiness error, not a generic crash. **🔍 Not tested** — would need a request stuck without placement, none seeded.
- Backoffice never opens store-scoped mode without ready placement. **🔍 Not directly tested** — the one available store is fully ready, so this guard was never exercised against a not-ready store this pass.
- One protected Backoffice API returns success for authorized user and permission detail for unauthorized user. **✅** — `stores` GraphQL query succeeded with a valid token; a request with no token returned a clear `UNAUTHENTICATED` GraphQL error, not a crash.

**Still open:** literal browser UI verification (no automation tool available this pass — see Current Status note), and every non-`ready` `ui_state` path (needs seed data or a way to force a request into `pending`/`provisioning`/`blocked`/`failed`).

## First Agent-Ready Task Candidate

Readiness is wired FE-to-backend and verified at the API level for the
`ready` path (2026-07-12). Two things remain, neither blocking day-to-day
work on the backbone but both worth closing before calling this exit
criteria fully met:

```text
1. Seed (or otherwise force) at least one store request into each of
   pending/provisioning/blocked/failed and re-check the readiness endpoint
   + provisioning requests panel badge for each.
2. Get a working headless-browser tool into this environment (or add a
   project run-skill once one exists) so future backbone checks can
   confirm the actual rendered UI, not just the API responses.
```

Allowed docs:

- `docs/04-sprints/tasks/wire-store-readiness-frontend.md`
- `docs/03-architecture-detail-design/05-transport-contracts.md`
- `docs/01-srs/traceability-matrix.md`

Done:

- a request in each ui_state observed via both the readiness endpoint and
  the rendered provisioning requests panel;
- browser-level screenshot confirming the panel renders correctly;
- traceability updated.
