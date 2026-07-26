# PZEP-0018: Backoffice store authorization — IAM as the single authority (MT5, MT7)

## Status
Done — MT5 and MT7 both implemented and tested (real Postgres 15 for MT7's
repository test, unit tests for MT5). `FindPage`'s Non-Goal is unchanged, as
scoped.

## Date
2026-07-26

## Related Commit
(none yet — implemented in the working tree this session, not committed)

## Requirement Sources
- Business: recovery track, continuing
  [PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md) Wave 5.
- Feature: none new — this closes a conflict between two existing
  authorization mechanisms (IAM policy statements vs. an `owner_id` column),
  it does not add a permission model.
- Use Cases: a tenant member who is not the store's `owner_id` but has been
  granted `store:read`/`store:update`/etc. in IAM can actually use that
  store — today they cannot (MT7), which means multi-user tenants do not
  work for anything but the literal creator of each store. A future DTO
  that adds a `storeId` argument anywhere in the routing domain cannot
  silently let a caller act on a different store than the one IAM checked
  permission against (MT5).
- Functional/Non-functional Requirements:
  `docs/03-architecture-detail-design/11-iam-platform.md` (IAM is the
  authority for tenant/store-scoped permission decisions).
- Acceptance Criteria: see below.
- UI Specs: none — backend only.

## Summary

Backoffice runs two authorization models that disagree with each other.
IAM policy statements (`internal/iam`) are checked once per GraphQL
operation/field via `TenantMiddleware.InterceptField` →
`RequirePermission(ctx, userID, tenantID, permission, resource)` — this is
the intended single authority per
`docs/03-architecture-detail-design/11-iam-platform.md`. But the store
repository (`internal/backoffice/infrastructure/repository/store/
repository.go`) separately filters every read/write by `owner_id ==
current user`, which silently overrides whatever IAM just decided: a tenant
member with a real `store:read`/`store:update` grant on a store they don't
personally own gets "not found" instead of access, because the SQL predicate
excludes the row before IAM's decision matters at all.

Separately, `RequiredStoreScope` (`internal/backoffice/domain/routing/
support.go`) lets an explicit caller-supplied store ID silently win over the
store ID IAM's `resourceForPermission` actually checked permission against
(`scope.CurrentStoreID(ctx)`), with no comparison between the two. Not
exploitable today (no GraphQL input currently carries a `storeId` argument —
confirmed against `internal/backoffice/controller/graphql/schema/*.graphqls`),
but the function's contract permits store-A-permission-checked /
store-B-actually-touched the moment such a field is added, and nothing
would catch it in review since the code already "supports" it.

This PZEP makes IAM's decision authoritative in both places, and closes the
gap in `RequiredStoreScope` before it becomes exploitable rather than after.

## Problem

**MT7** — `internal/backoffice/infrastructure/repository/store/
repository.go`:
```go
predicates = append([]sq.Sqlizer{sq.Eq{"owner_id": ownerID}}, predicates...) // FindPage
...
Where(sq.Eq{"id": id, "owner_id": ownerID}).ToSql()                          // FindByID
...
Where(sq.Eq{"id": id, "owner_id": ownerID}).ToSql()                          // UpdateStatus
```
`storeaccess.Service.ResolveStore` (used by `TenantMiddleware.ResolveRequestScope`
to populate `scope.CurrentStoreID` for every store-scoped request) goes
through `FindByID`, so a tenant member who is *not* the store owner fails
scope resolution — and gets an error for every store-scoped request — even
when IAM has already granted them `store:read` on exactly that store via a
policy statement (confirmed IAM's policy model supports this: both wildcard
`ResourcePattern: "*"` and per-resource `entity.ResourceStore(tenantID,
storeID)` statements exist and are tested,
`internal/iam/domain/interactor/authz_test.go`).

**MT5** — `internal/backoffice/domain/routing/support.go`:
```go
func RequiredStoreScope(ctx context.Context, explicitStoreID string) (string, error) {
	storeID := strings.TrimSpace(explicitStoreID)
	if storeID == "" {
		storeID = strings.TrimSpace(scope.CurrentStoreID(ctx))
	}
	...
}
```
The explicit value wins outright and is never compared against
`scope.CurrentStoreID(ctx)` — the store ID `resourceForPermission` actually
built the IAM resource string from
(`internal/backoffice/tenant_middleware.go:232-238`). Today every
`explicitStoreID` passed into this function already equals
`scope.CurrentStoreID(ctx)` (no GraphQL input has its own `storeId`), so the
mismatch can't currently occur — but the function's contract is "trust the
caller's value first," which is backwards from "IAM checked permission on
X, so only X may be acted on."

## Goals

- **MT7**: `FindByID` and `UpdateStatus` drop the `owner_id` predicate —
  access is decided entirely by `TenantMiddleware.InterceptField`'s
  `RequirePermission` call before the resolver reaches the repository at
  all. `owner_id` stays a real column (creator attribution / informational
  metadata), just no longer an authorization filter.
- **MT5**: `RequiredStoreScope` becomes an authority-comparison, not a
  fallback chain: if `explicitStoreID` is non-empty *and* differs from
  `scope.CurrentStoreID(ctx)`, return a new `ErrStoreScopeMismatch` instead
  of silently using the explicit value. Matches the fix note in
  `docs/07-problems/multitenancy-2026-07-23.md` MT5 exactly.

## Non-Goals

- **`FindPage` (the `stores` list query).** Its GraphQL field-level
  permission check uses `resourceForPermission(ctx, tenantID)`, which
  resolves to `"*"` for a list query (no single store is selected yet at
  list time) — so `RequirePermission` only proves "this caller has
  `store:read` on *some* resource pattern that happens to match `*`," not
  "here is the exact set of stores they may see." Correctly listing "every
  store this caller has a `store:read` grant on" requires an IAM capability
  this codebase does not have today (confirmed: no
  `ListAuthorizedResources`-shaped method anywhere in `internal/iam` —
  `ListPermissions` lists permission *definitions*, not per-resource
  grants). This is exactly the blocker MT7's own fix note flagged
  ("`FindPage` needs a listing policy decision... that requires an IAM
  listing capability, so flag it before implementing"). **`FindPage` keeps
  its current `owner_id` filter as an interim measure** — changing it
  without the IAM capability would either under-expose (if some members
  have wildcard grants, they'd still only see their own stores) or
  over-expose (if `owner_id`-filtering were simply dropped with no
  replacement, every tenant member with the tenant-wide `store:read`
  wildcard would see every store, including ones a more specific per-store
  policy meant to exclude them from — this codebase cannot currently tell
  those two cases apart without the missing IAM capability). Revisit once
  IAM adds resource listing.
- Adding the IAM listing capability itself — that is IAM's own feature,
  with its own PZEP, not something to bolt onto a backoffice authorization
  fix.
- Any change to `TenantMiddleware`, `permissionForField`, or the
  `RequirePermission` call itself — those already work correctly and are
  the authority this PZEP defers to, not something being replaced.
- MT4, MT9, MT10, MT11, MT16 — unrelated, tracked in
  [PZEP-0017](./PZEP-0017-multitenancy-correctness-observability.md).

## Proposed Solution

1. `internal/backoffice/infrastructure/repository/store/repository.go`:
   - `FindByID`: drop the `owner_id` predicate from the `Where` clause —
     `Where(sq.Eq{"id": id})` only. Drop the now-unused
     `toolkit.GetUserID(ctx)` call in this method if nothing else in it
     needs the caller's user ID.
   - `UpdateStatus`: same — `Where(sq.Eq{"id": id})` only.
   - `FindPage`: unchanged (see Non-Goals).

2. `internal/backoffice/domain/routing/support.go`:
   ```go
   var ErrStoreScopeMismatch = errors.New("store scope mismatch")

   func RequiredStoreScope(ctx context.Context, explicitStoreID string) (string, error) {
       explicit := strings.TrimSpace(explicitStoreID)
       scoped := strings.TrimSpace(scope.CurrentStoreID(ctx))
       switch {
       case explicit == "":
           explicit = scoped
       case scoped != "" && explicit != scoped:
           return "", ErrStoreScopeMismatch
       }
       if explicit == "" {
           return "", ErrStoreScopeRequired
       }
       return explicit, nil
   }
   ```
   Map `ErrStoreScopeMismatch` to the same GraphQL error class
   `ErrStoreScopeRequired` already maps to (403/permission-style, not a
   generic 500) — check `internal/backoffice/controller/graphql/resolver`'s
   error-mapping layer for the existing pattern before adding a new one.

## Affected Components
- Backend: `internal/backoffice/infrastructure/repository/store`,
  `internal/backoffice/domain/routing`
- Gateway/Frontend/Worker/Database schema/External Integration: none

## API / DB / Event Contract Changes
- API: no field/schema change. Behavior change: a tenant member with an
  IAM `store:read`/`store:update`/etc. grant on a store they don't own can
  now actually use it (previously silently blocked) — this is a bugfix
  matching IAM's already-documented intent, not a new capability being
  granted.
- DB: none — `owner_id` column unchanged, stops being read as an auth
  filter in two methods.
- Event: none.

## Permission Changes
This *is* the permission-model fix: after this PZEP, IAM policy statements
are the sole authority for store-scoped read/write access in the two
methods listed. `owner_id` is no longer a de facto second permission model.

## Security Considerations
- **Authorization**: strictly narrows to "trust IAM's decision, nothing
  else" — this closes a bug where `owner_id` was *more* restrictive than
  IAM in the single-user case (today's only real-world case, since
  multi-user tenants don't work yet) and will become *less* restrictive
  than before for the specific case IAM was already supposed to allow (a
  granted, non-owner member) — i.e. this fixes an under-permissioning bug,
  it does not introduce an over-permissioning one, provided
  `RequirePermission` is correctly checked before every call site that
  reaches these two repository methods (verify this holds for all current
  and future callers as part of Test Plan/code review, not assumed here).
- **Tenant/store isolation**: unaffected — `withTenantTx`'s tenant-scoping
  (`toolkit.GetTenantID` → `mgr.WithTenantTx`) is untouched; a store row
  physically lives in its tenant's schema regardless of this PZEP.
- MT5's fix is defense-in-depth against a currently-nonexistent exploit
  path (see Problem) — low urgency, but cheap and worth closing before a
  future DTO opens it.

## Test Plan
- MT7: `internal/backoffice/infrastructure/repository/store` — repository
  test asserting `FindByID`/`UpdateStatus` succeed for a store whose
  `owner_id` differs from the calling context's user ID (proving the
  predicate is genuinely gone, not just unused). Existing tests that relied
  on the `owner_id` filter as an implicit access check need review — none
  should have been asserting *security* through it (that's
  `RequirePermission`'s job, tested in `internal/iam`), only "my own store
  is findable," which continues to hold.
- MT5: `internal/backoffice/domain/routing` — unit tests: (a) empty
  `explicitStoreID` + populated `scope.CurrentStoreID` → scoped value
  returned (existing behavior, must not regress); (b) `explicitStoreID`
  equal to `scope.CurrentStoreID` → that value returned, no error; (c)
  `explicitStoreID` non-empty and differing from a populated
  `scope.CurrentStoreID` → `ErrStoreScopeMismatch`; (d) both empty →
  `ErrStoreScopeRequired` (existing behavior, must not regress).
- Full `internal/backoffice/...` suite continues to pass.
- Manual QA: exercise a store-scoped query as a tenant member who is not
  the store's `owner_id` but has an explicit IAM `store:read` grant on it,
  against a real (docker dev) environment with IAM policies actually
  configured — this is the concrete multi-user scenario MT7 names as
  currently broken, and unit tests alone won't prove the IAM policy side is
  wired correctly end to end.

## Agent Implementation Plan
- TASK-0001: MT7 — drop `owner_id` predicate from `FindByID`/`UpdateStatus`. **Done.**
- TASK-0002: MT5 — `RequiredStoreScope` mismatch detection +
  `ErrStoreScopeMismatch` + GraphQL error mapping. **Done.**
- TASK-0003: test coverage per Test Plan (unit); flag the manual-QA step
  for whoever implements this, since it needs a real IAM policy set up in
  a running environment, not just automated tests. **Done** (automated
  portion) — manual QA against a real IAM policy set is still open, see
  Rollout Log.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: a tenant member with an IAM `store:read` grant on a store they don't own can `FindByID` it | TASK-0001 | `TestStoreRepository_FindByIDAndUpdateStatus_IgnoreOwnerID` |
| AC-2: a tenant member with an IAM `store:update`-class grant on a store they don't own can `UpdateStatus` it | TASK-0001 | `TestStoreRepository_FindByIDAndUpdateStatus_IgnoreOwnerID` |
| AC-3: `RequiredStoreScope` returns `ErrStoreScopeMismatch` when the explicit and scoped store IDs disagree | TASK-0002 | `TestRequiredStoreScope` |
| AC-4: `RequiredStoreScope`'s existing single-value behavior (scoped-only, explicit-matching-scoped, both-empty) is unchanged | TASK-0002 | `TestRequiredStoreScope` |

## Rollout Log

**2026-07-26:**
1. **MT7**: dropped the `owner_id` predicate from `FindByID` and
   `UpdateStatus` (`internal/backoffice/infrastructure/repository/store/
   repository.go`); `FindPage` unchanged per Non-Goals. Added
   `internal/backoffice/infrastructure/repository/store/repository_test.go`
   (new file — this package had no repository-level test before), a real
   Postgres 15 (testcontainers) integration test proving a non-owner tenant
   member can `FindByID` and `UpdateStatus` a store they don't own.
2. **MT5**: added `ErrStoreScopeMismatch` to
   `internal/backoffice/domain/routing/errors.go` using the existing
   `ddd.NewDomainError` pattern — confirmed this already auto-maps to a
   GraphQL error via `ddd.IsDomainError`'s generic presenter, no per-error
   mapping code needed (resolves this PZEP's own "exact GraphQL error-code
   mapping" Open Question). Rewrote `RequiredStoreScope` to compare
   explicit vs. scoped values instead of letting explicit win
   unconditionally. Added `internal/backoffice/domain/routing/
   support_test.go` (new file), five table-driven cases covering both the
   new mismatch behavior and all pre-existing behavior (regression).
3. Kept `FindByID`'s `toolkit.GetUserID(ctx)`... **removed** - decided
   during implementation rather than left open: `FindByID` no longer uses
   the caller's user ID for anything, so keeping an unused call would just
   be dead code with no audit-logging consumer today; can be re-added
   if/when audit logging is actually implemented.

Verification: `go build ./...`, `go vet ./internal/backoffice/...`,
`golangci-lint run` on the two touched packages (0 issues), and the full
`internal/backoffice/...` suite — all pass.

Not done: manual QA against a real IAM policy set in a running environment
(this PZEP's Test Plan flags this as needed beyond automated tests — a
tenant member with a real, specifically-scoped `store:read`/`store:update`
IAM grant, exercised end to end through the GraphQL API). No commit yet.

## Open Questions
- ~~Exact GraphQL error-code mapping for `ErrStoreScopeMismatch`~~ —
  resolved: automatic via the existing `ddd.DomainError` presenter, no new
  mapping code needed.
- ~~Whether `FindByID`'s now-possibly-unused `toolkit.GetUserID(ctx)` call
  should be removed~~ — resolved: removed, see Rollout Log.
- `FindPage`'s Non-Goal is a real, user-visible limitation (non-owner
  members still can't list stores they have access to, only fetch them
  directly by ID once they know it) — worth surfacing to the user/product
  owner as a known gap rather than treating it as fully closed once this
  PZEP ships.
