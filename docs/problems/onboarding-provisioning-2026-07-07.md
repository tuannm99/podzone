# Onboarding Provisioning — Known Problems

Identified from code review of commits b3deb73..43530c2 (2026-07-07).
Scope: `internal/onboarding/`, `scripts/dev/reconcile_legacy_tenant.go`.

---

## P1 — LIKE `_` wildcard over-counts tenant schemas

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

**File:** `scripts/dev/reconcile_legacy_tenant.go:178`
**Severity:** Low — dev/seed script only, but causes confusing timeout failures

`findRequest` always fetches page 1 with size 100. In a workspace with >100 existing store
requests, the target subdomain falls on page 2+ and is never found. The poll loop prints
stderr errors and hits the 120s timeout, halting the seed pipeline with a misleading error.

**Fix:** Either paginate until found, or (preferred) add a `?subdomain=` query param to the
list API and filter server-side.
