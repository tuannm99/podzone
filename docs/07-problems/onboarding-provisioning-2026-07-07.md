# Onboarding Provisioning — Known Problems

Identified from code review of commits b3deb73..43530c2 (2026-07-07).
Scope: `internal/onboarding/`, `scripts/dev/reconcile_legacy_tenant.go`.

Status as of 2026-07-11: P1 confirmed resolved (see note inline).
Status as of 2026-07-26: **P2, P3, P4, P5, P6, P7 all confirmed resolved
against current code** (see notes inline). P2/P3/P4/P6/P7 were fixed by
commits `73be9f5` ("fix(onboarding): P1/P2/P3/P5 from known problems doc")
and `27b26f3` ("fix(onboarding): P4/P6/P7 — audit phantom, TOCTOU
double-fetch, double-provision race"), both 2026-07-09 — this file's status
line simply hadn't been updated since, the same class of staleness found in
`pkg-2026-07-07.md`. **P8 resolved** (Wave 1 fix, PZEP-0016 — `findRequest`
now filters server-side by subdomain instead of paginating a fixed 100-row
page). Every item in this note is now resolved.

**One new, real finding surfaced while re-verifying P7**: the fix applied
for P7 (removing `Planning` from `ClaimNextQueued`'s filter) closed the
double-provisioning race but left no path back to `Queued` for a request
whose worker crashed mid-`Planning` — `isValidStatusTransition` does not
allow `Planning → Queued`, so neither the automatic claim path nor the
manual `RetryStoreRequest` action could ever recover it; it would stay
`Planning` forever once its lease expired. **Fixed 2026-07-26** (PZEP-0016
Wave 3): added `StoreRepository.ReclaimStalePlanning` (atomic
`FindOneAndUpdate`, same safety property as `ClaimNextQueued`) and
`Usecase.ReclaimNextStalePlanningRequest`, wired into
`StoreProvisioningWorker.tick()` as a sweep that runs ahead of claiming new
work — this is the "separate explicit re-queue path" the original P7 fix
note anticipated but never added, implemented as an automatic sweep rather
than an operator action since the worker already ticks on an interval.

---

## P1 — LIKE `_` wildcard over-counts tenant schemas

**Resolved as of 2026-07-11.** `health.go` now uses
`starts_with(schema_name, $1)`, exactly the fix suggested below.

**File:** `internal/onboarding/infrastructure/provisioning/provider/health.go:71`
**Severity:** High — corrupts capacity metrics used for placement planning

`countTenantSchemas` runs:
```sql
SELECT count(*) FROM information_schema.schemata WHERE schema_name LIKE $1
```
with argument `"t_%"`. In PostgreSQL, `_` is a single-character wildcard, so schemas like
`test_data`, `tmp_log`, `ta_cache` all match. `CurrentSchemas` is inflated; the placement
planner may think a cluster is full when it is not.

**Fix:** escape the underscore:
```go
schemaPrefix+`\_%`, // then add ESCAPE '\' to the SQL, or use starts_with()
```
Or use `starts_with(schema_name, $1)` (PostgreSQL 11+) which has no wildcard semantics.

---

## P2 — `authorizer==nil` + `ownerID!=requestedBy` → ErrAccessDenied blocks dev seeding

**Resolved (commit `73be9f5`), confirmed against current code 2026-07-26.**
`CreateStoreRequest`'s owner-override check is now gated by
`s.authorizer != nil` (`interactor.go:83`) — when the authorizer is not
wired (dev mode), the override is permitted with no check at all, exactly
the suggested fix's "permissive dev mode" option.

**File:** `internal/onboarding/domain/store/interactor.go:84`
**Severity:** High — breaks `make dev-pod-up`, `make dev-pod-sample`, `make dev-onboarding-reconcile-tenant`

`CreateStoreRequest` returns `ErrAccessDenied` immediately when `ownerID != requestedBy`
and `authorizer == nil`. `reconcile_legacy_tenant.go` always sets `owner_id` explicitly
(from `OWNER_ID` env var), which differs from the service-token identity used as
`requestedBy`. In dev environments where the authorizer is not wired, the create call fails.

**Fix:** When `authorizer == nil`, treat the ownerID override as permitted (permissive dev
mode). Or gate the check on a config flag (`provisioner.RequireApprovalForOwnerOverride`).

---

## P3 — Finalization batch loop breaks on route-not-ready, starving other Provisioning items

**Resolved (commit `73be9f5`), confirmed against current code 2026-07-26.**
`tick()` now checks `errors.Is(err, storedomain.ErrRouteNotReady)` and
`continue`s instead of breaking, exactly the suggested fix.

**File:** `internal/onboarding/infrastructure/messaging/worker/store_provisioning_worker.go:94`
**Severity:** High — provisioning starvation in steady state

`FinalizeNextStoreRequest` returns `(nil, nil)` in two distinct cases:
1. Queue is empty (intended break)
2. A Provisioning item was claimed but its KV route is not yet ready — lease is released, `nil` returned

`tick()` treats both as "no more work" and breaks. If the oldest Provisioning item (sorted
by `updated_at` asc) has no route yet, it is claimed first, the batch loop breaks, and all
other stores with ready routes wait a full tick interval. A slow route can repeatedly block
the queue.

**Fix:** Return a distinct sentinel (`ErrRouteNotReady` or a `(nil, false, nil)` triplet)
so `tick()` can `continue` instead of `break` when route is not ready.

---

## P4 — Phantom Queued→Planning audit entry when re-claiming expired Planning record

**Resolved (commit `27b26f3`), confirmed against current code 2026-07-26.**
`claimNext` now uses `SetReturnDocument(options.Before)`, and
`ProcessNextStoreRequest` passes the real pre-update `request.Status` (not a
hardcoded `RequestStatusQueued`) as `recordTransition`'s `from` argument —
exactly the suggested fix. (In current code this is now moot in the way the
original finding described it, since P7's fix means `ClaimNextQueued` only
ever matches `Queued` records — but the general "use the real pre-update
status" fix is still what's in place, and it's what makes the *new*
`ReclaimNextStalePlanningRequest` sweep's own audit trail correct too — see
this file's top-level note.)

**File:** `internal/onboarding/domain/store/interactor.go:384`
**Severity:** Medium — corrupts provisioning audit trail

`ClaimNextQueued` now filters on `[Queued, Planning]` to recover from crashed workers.
When a lease-expired `Planning` record is re-claimed, `FindOneAndUpdate` returns
`status=Planning`, but `recordTransition` still hardcodes `from=RequestStatusQueued`.
The transition log shows a `Queued→Planning` entry that never occurred.

**Fix:** Capture the pre-update status returned by the repo (use `SetReturnDocument(Before)`
or add a separate field) and pass the actual `from` status to `recordTransition`.

---

## P5 — `pg_stat_activity` self-count inflates `CurrentConnections` by 1

**Resolved, confirmed against current code 2026-07-26.**
`countDatabaseConnections` already excludes the health check's own backend
(`AND pid <> pg_backend_pid()`) — the exact fix suggested below.

**File:** `internal/onboarding/infrastructure/provisioning/provider/health.go:53`
**Severity:** Medium — systematically wrong connection count persisted to inventory

`countDatabaseConnections` runs while the health-check's own `sql.DB` connection is still
open. The query counts all sessions for the database, including itself:
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = $1
```
Every health check persists a `CurrentConnections` value that is at least 1 higher than
the actual tenant workload.

**Fix:**
```sql
SELECT count(*) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()
```

---

## P6 — TOCTOU in `ReconcileTenantPlacement` — returned Status mismatches published route

**Resolved (commit `27b26f3`), confirmed against current code 2026-07-26.**
`ReconcileTenantPlacement` now calls a single `fetchPlacementState` helper
once and threads the result through both the status computation and the
route publish — no second fetch, no TOCTOU window.

**File:** `internal/onboarding/domain/infrasmanager/placement_reconcile.go:63`
**Severity:** Medium — repair response reports stale state

`ReconcileTenantPlacement` calls `GetTenantPlacementStatus` (which fetches allocation A),
then calls `GetTenantPlacementAllocation` a second time. If a concurrent provisioning call
updates the allocation between the two reads, the published route reflects allocation B but
the returned `PlacementReconcileResponse.Status` is built from allocation A. Callers
receive `Repaired=true` with mismatched status data.

**Secondary issue:** The double fetch is also a wasted MongoDB round-trip on every repair call.

**Fix:** Return the allocation from `GetTenantPlacementStatus` (or a combined result struct)
and thread it through to avoid the second fetch entirely.

---

## P7 — `ClaimNextQueued` includes `Planning` in filter — double-provisioning window

**Resolved (commit `27b26f3`), confirmed against current code 2026-07-26.**
`ClaimNextQueued` now filters on `Queued` only — the double-provisioning
race is closed. **Residual gap found and fixed 2026-07-26** (PZEP-0016
Wave 3, see this file's top-level note): removing `Planning` from the
filter left no recovery path for a crashed-worker's stuck `Planning`
record, since `Planning → Queued` was never a valid manual transition
either. Added `ReclaimStalePlanning`/`ReclaimNextStalePlanningRequest`,
exactly the "separate explicit re-queue path" this fix's own original note
(below) anticipated but didn't implement at the time.

**File:** `internal/onboarding/infrastructure/repository/store/repository.go:193`
**Severity:** Medium — concurrent double-provisioning on lease expiry

`ClaimNextQueued` passes `[RequestStatusQueued, RequestStatusPlanning]` as the `current`
status filter. If Worker A holds a Planning record and its lease expires before it finishes
`ProvisionStorePlacement`, Worker B will match the same record (lease_until ≤ now) and
also start provisioning. The partial unique index on `(tenant_id, status=ready)` only fires
at the ready-transition — two concurrent in-progress placements can exist simultaneously.

**Fix:** Remove `RequestStatusPlanning` from `ClaimNextQueued`'s filter. Create a separate
explicit re-queue path (`RequeueStalePlanning`) for operator-triggered recovery, or extend
the lease TTL to be safely longer than the worst-case provisioning duration.

---

## P8 — `findRequest` hardcodes `pageSize=100`, misses target if >100 store requests exist

**Resolved 2026-07-26 (PZEP-0016 Wave 1).** `findRequest` now sends
`collection.filters[0].field=subdomain&operator=eq&values=<subdomain>` with
`collection.pageSize=1`, filtering server-side via the same `collection.Query`
mechanism `ListStoreRequests` already parses
(`internal/onboarding/controller/httphandler/store/controller.go:106`,
`internal/onboarding/infrastructure/repository/store/collection_query.go:32`
confirms `subdomain` is a filterable field) instead of paginating a fixed
100-row client-side scan. The in-memory `WorkspaceID`/`Subdomain` match in
the loop below is kept as a defense-in-depth check, not the primary filter.

**File:** `scripts/dev/reconcile_legacy_tenant.go:178`
**Severity:** Low — dev/seed script only, but causes confusing timeout failures

`findRequest` always fetches page 1 with size 100. In a workspace with >100 existing store
requests, the target subdomain falls on page 2+ and is never found. The poll loop prints
stderr errors and hits the 120s timeout, halting the seed pipeline with a misleading error.

**Fix:** Either paginate until found, or (preferred) add a `?subdomain=` query param to the
list API and filter server-side.
