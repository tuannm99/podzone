# PZEP-0017: `pkg/pdtenantdb` correctness & observability batch (MT4, MT9, MT10, MT11, MT16)

## Status
Done — all five items (MT4, MT9, MT10, MT11, MT16) implemented and tested
against real Postgres 15 (testcontainers). See Rollout Log for design
deviations from the Draft found during implementation.

## Date
2026-07-26

## Related Commit
(none yet — implemented in the working tree this session, not committed)

## Requirement Sources
- Business: recovery track, continuing
  [PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md) Wave 4 (the
  items not already closed by MT6's direct fix or PZEP-0015).
- Feature: none new — correctness/observability hardening for existing
  `pkg/pdtenantdb` behavior.
- Use Cases: an operator restoring a dropped-then-recreated schema doesn't
  need a process restart to recover (MT10); a tenant using a dedicated
  database pool doesn't hit a rare `sql: database is closed` from the idle
  janitor racing a fresh checkout (MT11); nobody can accidentally reintroduce
  MT4's raw-connection hazard because the type system no longer offers it;
  a stale placement cache entry expires within a bounded, configurable
  window instead of a hardcoded 2 minutes, and a tenant with no placement at
  all doesn't hit the KV store on every single request (MT9); when a tenant
  reports "my data is missing", there is a log line to check instead of
  nothing (MT16, observability half only — see Non-Goals for the metrics
  gap).
- Functional/Non-functional Requirements:
  `docs/00-governance/twelve-factor.md` (config via environment).
- Acceptance Criteria: see below.
- UI Specs: none — backend only.

## Summary

Four independent, already-diagnosed defects in `pkg/pdtenantdb`
(`docs/07-problems/multitenancy-2026-07-23.md` MT4, MT9, MT10, MT11), plus
the observability half of a fifth (MT16 — the test half of MT16 is already
satisfied by [PZEP-0015](./PZEP-0015-backoffice-per-tenant-db-role-phase-2.md)'s
mutual-invisibility integration test). None change the IAM/permission
contract (that's [PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md),
Wave 5). Grouped into one PZEP because they all live in the same package and
several touch the same `Manager` struct's locking/caching internals, so
reviewing them together avoids three small diffs each re-deriving the same
context about `m.mu`'s scope.

## Problem

- **MT4**: `Manager.DBForTenant` is on the public `Manager` interface and
  returns a raw `*sqlx.DB` with no `search_path`/role applied. No caller
  outside `pkg/pdtenantdb` currently uses it directly (confirmed: `grep -rn
  "\.DBForTenant(" internal/ cmd/` outside `_test.go` returns nothing), but
  the hazard is exactly that it's easy to reach for later without noticing
  it skips every isolation mechanism `WithTenantTx` provides — including
  [PZEP-0015](./PZEP-0015-backoffice-per-tenant-db-role-phase-2.md)'s
  per-tenant role. This is also the specific gap that PZEP-0015's "Known
  residual gap" note names as still open after Phase 2 ships.
- **MT9**: `pkg/pdtenantdb/fxmodule.go` wires `NewKVPlacementResolver` (the
  zero-arg constructor, hardcoded 2-minute TTL) — confirmed the
  configurable `NewKVPlacementResolverWithTTL` escape hatch exists in code
  but nothing calls it anywhere in the repo. So the real deployed behavior
  really does have an un-configurable 2-minute placement-cache TTL and no
  negative caching: a tenant with no placement key (e.g. mid-onboarding)
  hits the KV store on every single request.
- **MT10**: `pkg/pdtenantdb/manager.go`'s `bootstrapped sync.Map` and
  `internal/backoffice/migrations/migrations.go`'s package-level
  `appliedScopes sync.Map` both record a permanent "done" bit per scope key,
  with no expiry and no re-validation. If a schema is dropped or restored
  out of band, every subsequent call believes it still exists and fails
  until the process restarts.
- **MT11**: `CloseIdleDedicated` (`manager.go`) closes a dedicated pool
  under `m.mu` based only on `lastUsed`, with no refcount. The exact race:
  `DBForTenant` returns a pool reference and releases `m.mu` (after calling
  `markUsed`); before the caller's subsequent `db.BeginTxx` actually checks
  out a physical connection, the janitor's tick can observe the just-updated
  `lastUsed` is still `>= DedicatedIdleTTL` in the past (a stale read from
  before `markUsed` ran, or a caller that took unusually long between
  `DBForTenant` returning and `BeginTxx` starting) and close the pool.
- **MT16** (observability half only): no log line anywhere in
  `pkg/pdtenantdb` for pool open/close, placement resolution failure,
  placement cache hit/miss, or `ErrDedicatedPoolCapacity` rejections. When a
  tenant reports missing data, there is no signal to distinguish "stale
  placement" from "wrong search_path" from "failed migration" without
  attaching a debugger.

## Goals

- **MT4**: remove `DBForTenant` from the `Manager` interface; keep the
  underlying logic as an unexported method on `managerImpl` (`WithTenantTx`
  is the only caller). `pkg/pdtenantdb/mocks/mock_Manager.go` regenerates
  without it; any test that currently calls `Manager.DBForTenant` through
  the interface switches to testing through `WithTenantTx` instead (proof
  that the public surface genuinely no longer offers the raw-handout path).
- **MT9**: `pdtenantdb.Config` gains `PlacementCacheTTL time.Duration`
  (wired through `internal/backoffice/config.TenantDBConfig`, same pattern
  MT2 already established for pool sizing) and
  `pdtenantdb.Module`/`fxmodule.go` passes it to `NewKVPlacementResolverWithTTL`
  instead of the hardcoded-TTL constructor. `KVPlacementResolver` also
  negative-caches `ErrPlacementNotFound` with its own (shorter, fixed at
  `min(ttl, 15s)`) TTL, so a tenant with no placement yet stops hitting the
  KV store on every request without needing a config knob for this specific
  sub-case.
- **MT10**: `ensureSchema`'s `bootstrapped` map and `migrations.appliedScopes`
  both evict their entry for a scope key when a subsequent query against
  that scope fails with a Postgres error indicating the schema is actually
  gone (`SQLSTATE 3F000`, `invalid_schema_name` — covers both "schema does
  not exist" and, conservatively, "search_path invalid" class errors), so
  the *next* call re-runs `CREATE SCHEMA IF NOT EXISTS` + migrations instead
  of failing forever. This is detection at the call site that uses the
  pool, not inside `ensureSchema` itself (which has no way to know a schema
  disappeared after it last checked) — see Proposed Solution for where this
  hooks in.
- **MT11**: add a `checkouts map[ConnKey]int` refcount to `managerImpl`,
  protected by `m.mu`. `getOrCreateDB` increments the resolved key's count
  before returning the pool; `WithTenantTx` decrements it in a `defer`
  immediately after `db.BeginTxx` returns (success or error) — this brackets
  exactly the unsafe window the bug report describes (pool reference handed
  out, but a physical connection not yet checked out for the caller's use).
  `CloseIdleDedicated` skips any key with a nonzero refcount, in addition to
  its existing `lastUsed` check.
- **MT16** (observability half): structured log lines (via the existing
  `pdlog.Logger`, no new dependency) for: pool opened/closed (cluster, db,
  dedicated y/n), placement resolution failure (tenant, error), placement
  cache hit vs. miss (debug level, sampled or rate-limited to avoid
  per-request log spam — see Open Questions), and `ErrDedicatedPoolCapacity`
  rejections (cluster, current count, limit). Not a metrics/counter system
  — see Non-Goals.

## Non-Goals

- **MT16's metrics half.** This codebase has no metrics/tracing library
  wired anywhere (`grep -rln "prometheus\|otel\|metrics\." pkg/
  internal/backoffice/` — outside test/mock files — returns nothing). Adding
  one is an observability-stack decision affecting every service, not just
  `pdtenantdb`, and belongs in its own ADR, not this PZEP. The log lines
  this PZEP adds are a floor, not a replacement for real metrics.
- **MT9's KV-watch / event-driven invalidation.** A configurable TTL and
  negative caching close the two most impactful sub-findings (stale window
  bound is now operator-controlled; the "hits KV every request" case for a
  never-placed tenant is fixed). A push-based invalidation channel needs a
  KV-watch capability or an event bus subscription — bigger scope, revisit
  only if the configurable-TTL window (still capped low, e.g. seconds, not
  minutes) proves insufficient in practice.
- **MT10's fully general "any Postgres error evicts the cache" behavior.**
  Scoped specifically to the `invalid_schema_name` SQLSTATE class to avoid
  evicting (and re-running `CREATE SCHEMA IF NOT EXISTS` + a full migration
  pass) on unrelated transient errors (e.g. a deadlock, a connection drop) —
  that would turn every unrelated failure into extra DDL load.
- **MT11's alternative "hold `m.mu` through `BeginTxx`" option** the
  original finding also mentions — rejected because `BeginTxx` can do
  network I/O (opening a fresh physical connection) and holding the
  manager-wide lock for that duration would serialize every tenant's pool
  operations behind one connection's dial time. Refcounting has no such
  cost.
- Any change to `WithTenantTx`'s public signature or `Manager`'s other
  methods (`CloseAll`, `CloseIdleDedicated`'s own signature).
- MT5, MT7 — permission-model changes, tracked separately in
  [PZEP-0018](./PZEP-0018-store-authorization-iam-single-authority.md).

## Proposed Solution

1. **MT4** — `pkg/pdtenantdb/manager.go`: rename the `Manager` interface's
   `DBForTenant` method away from the interface (keep the concrete
   `managerImpl.dbForTenant` — lowercase — as the implementation
   `WithTenantTx` calls). `pkg/pdtenantdb/manager_test.go`'s
   `TestManager_DBForTenant_ReusesPoolByClusterAndDB` and
   `TestManager_DedicatedPoolCapacity` currently call `m.DBForTenant(...)`
   directly through the `Manager` interface type — both need to switch to
   asserting pool reuse / capacity errors through `WithTenantTx` instead
   (e.g. run two `WithTenantTx` calls for the same tenant and assert no
   error / same underlying behavior, rather than comparing `*sqlx.DB`
   pointers directly — may need a small test-only accessor or an
   observable side effect instead of pointer equality; resolve exact
   mechanism during implementation, not in this PZEP). Regenerate
   `pkg/pdtenantdb/mocks/mock_Manager.go` via `go tool mockery`.

   **Implemented as:** the concrete method was renamed `dbForTenant`
   (unexported), not deleted — `WithTenantTx` is its only caller.
   `TestManager_DBForTenant_ReusesPoolByClusterAndDB` was renamed
   `TestManager_WithTenantTx_ReusesPoolByClusterAndDB` and now wraps the
   package-level `SQLXOpen` var (already overridable, used elsewhere in
   this file) to count real pool opens — two `WithTenantTx` calls for the
   same tenant must open exactly one pool. `TestManager_DedicatedPoolCapacity`
   simply calls `WithTenantTx` with a no-op `fn` instead of `DBForTenant`
   directly; the capacity error still propagates identically since
   `WithTenantTx` calls `dbForTenant` first thing and returns its error
   immediately.

2. **MT9** — `pkg/pdtenantdb/types.go`: add `PlacementCacheTTL
   time.Duration` to `Config` (defaults to today's 2 minutes when zero, so
   this is backward compatible). `pkg/pdtenantdb/fxmodule.go`: wire
   `NewKVPlacementResolverWithTTL(kv, "podzone/tenants",
   cfg.PlacementCacheTTL)` instead of the zero-arg constructor.
   `pkg/pdtenantdb/resolver_kv.go`: `KVPlacementResolver.Resolve` caches a
   negative result (`ErrPlacementNotFound` specifically, not
   `ErrPlacementBackend`) in a small second map with its own short TTL, so
   repeated lookups for a tenant with no placement yet don't each hit KV.
   `internal/backoffice/config.TenantDBConfig` gains
   `placement_cache_ttl_seconds` (same pattern as MT2's
   `conn_max_lifetime_seconds`), threaded through
   `internal/backoffice/tenant_db_bootstrap.go`.

3. **MT10** — `pkg/pdtenantdb/manager.go`: `WithTenantTx`'s error path (and
   any other call site that runs a query against a schema-scoped
   connection) checks the returned error for `SQLSTATE 3F000`
   (`invalid_schema_name`) via `github.com/lib/pq.Error.Code` and, if
   matched, evicts the scope key from `m.bootstrapped` before returning the
   error — so the *next* call re-bootstraps. Same shape for
   `internal/backoffice/migrations/migrations.go`'s `appliedScopes`: add an
   exported `InvalidateScope(database, schema string)` (or similar) that
   `WithTenantTx`'s error path calls when it detects the schema is gone,
   since `appliedScopes` is a separate cache keyed differently
   (`database|schema` vs. `bootstrapped`'s `cluster|db|schema`) and
   `pdtenantdb` cannot reach into `migrations`'s package-level map
   directly (layering — `internal/backoffice/migrations` imports
   `pkg/pdtenantdb`, not the reverse). Exact wiring (a callback registered
   via `Config`, similar to `OnSchemaBootstrap`) to be finalized during
   implementation.

   **Implemented as:** the actual SQLSTATE is **`42P01` (`undefined_table`),
   not `3F000`** — verified empirically against a real Postgres 15 container
   before writing any code (`SET LOCAL search_path` to a dropped schema
   succeeds silently; the failure only surfaces on the next unqualified
   table reference, as `42P01`). Since `42P01` alone is too broad a trigger
   (any missing table, including an unrelated app bug, raises it), added a
   confirmation step: `schemaLikelyMissing` only evicts after also checking
   `SELECT to_regnamespace($1) IS NOT NULL` (also verified empirically)
   returns false for the schema name — i.e. the schema itself is
   confirmed gone, not just one table in it. `OnScopeInvalidated
   func(database, schemaName string)` was added to `Config` exactly as
   sketched, wired to `migrations.InvalidateScope` in
   `internal/backoffice/tenant_db_bootstrap.go`.

4. **MT11** — `pkg/pdtenantdb/manager.go`: add `checkouts map[ConnKey]int`
   to `managerImpl`, initialized in `NewManager`. `getOrCreateDB` increments
   `checkouts[key]` under `m.mu` right before returning the resolved pool
   (both the cache-hit fast path and the newly-created-pool path).
   `WithTenantTx` decrements it (also under `m.mu`, via a small helper) in a
   `defer` scheduled immediately after `db.BeginTxx` returns.
   `CloseIdleDedicated` adds `checkouts[k] > 0` to its existing skip
   condition.

   **Implemented as:** the checkout/release pair lives in `dbForTenant`
   (checkout, right after `getOrCreateDB` succeeds) and `WithTenantTx`
   (release, via `defer` immediately after `dbForTenant` returns
   successfully) — meaning the refcount is held for the pool's *entire*
   `WithTenantTx` lifetime, not narrowed to just the pre-`BeginTxx` window
   originally sketched. This is simpler and strictly safer (a pool with an
   open transaction is definitionally not idle, so holding the checkout
   through `fn`/`Commit` doesn't change `CloseIdleDedicated`'s intended
   behavior) and needed no extra bookkeeping beyond one `checkout`/`release`
   call pair.

5. **MT16** (observability half) — add `pdlog.Logger` as a field on
   `managerImpl` (currently absent; `Config` or `NewManager`'s signature
   gains it, optional/nil-safe so existing callers that don't pass one
   don't break — check whether making it required is acceptable instead,
   given `internal/backoffice/tenant_db_bootstrap.go` already has a logger
   available via fx). Log points: `getOrCreateDB` (pool opened/closed,
   dedicated y/n), `KVPlacementResolver.Resolve` (resolution failure only,
   not every success — avoid per-request spam), `reserveDedicatedLocked`
   (capacity rejection).

## Affected Components
- Backend: `pkg/pdtenantdb`, `internal/backoffice/migrations`,
  `internal/backoffice/config`, `internal/backoffice/tenant_db_bootstrap.go`
- Gateway/Frontend/Worker/Database schema/External Integration: none

## API / DB / Event Contract Changes
- API: none (all changes are internal Go interfaces/config, not
  HTTP/GraphQL contracts).
- DB: none.
- Event: none.
- Internal interface: `Manager.DBForTenant` removed from the public
  interface (MT4) — a breaking change for any out-of-tree caller, but
  confirmed there are none in this repo today.

## Permission Changes
None — no IAM/JWT permission model change. (MT5/MT7, which do change this,
are explicitly out of scope — see Non-Goals and PZEP-0018.)

## Security Considerations
MT4's fix closes a real, if currently unused, escape hatch around
PZEP-0015's per-tenant-role isolation — worth doing regardless of whether
Phase 2 (`TenantRoleEnabled`) is ever turned on, since it also bypasses
Phase 1's `search_path` discipline. The other four items (MT9/MT10/MT11/MT16)
are reliability/observability, not security boundaries.

## Test Plan
- MT4: existing `manager_test.go` tests adapted to go through `WithTenantTx`
  instead of calling `DBForTenant` directly (see Proposed Solution); a
  compile-time check that `Manager` no longer exposes `DBForTenant` is
  implicit (the old test code simply won't compile against the new
  interface until updated).
- MT9: unit test with a fake `kvstores.KVStore` — assert
  `NewKVPlacementResolverWithTTL`'s configured TTL is actually honored (not
  hardcoded 2 minutes) via `fxmodule.go`'s wiring — likely needs a small
  integration test at the `internal/backoffice` config-wiring level, not
  just `pkg/pdtenantdb` in isolation. Unit test for negative caching:
  `ErrPlacementNotFound` on the first call, confirm the second call within
  the negative-cache TTL doesn't hit the fake KV store's `Get` again
  (`mock.Once()` style assertion).
- MT10: integration test (testcontainers Postgres) — bootstrap a schema,
  manually `DROP SCHEMA` out of band (simulating operator action), assert
  the next `WithTenantTx` call for that tenant fails once with the
  `invalid_schema_name` error, then the call *after that* succeeds (proving
  eviction + re-bootstrap, not just a single retry).
- MT11: unit test using a fake/slow `SQLXOpen` (already an overridable
  package var, see `manager.go`'s `var SQLXOpen = sqlx.Open`) to widen the
  window between pool-returned and `BeginTxx`-called, racing
  `CloseIdleDedicated` against it — assert the pool survives (refcount
  holds it) — mirrors how MT17's fix was verified with a concurrency test.
- MT16: unit test asserting the expected log calls happen (using the
  existing `pdlog/mocks.MockLogger`, same pattern
  `store_provisioning_worker_test.go` already uses) for at least the
  dedicated-pool-capacity-rejection and placement-resolution-failure cases.
- Existing `pkg/pdtenantdb` and `internal/backoffice` suites continue to
  pass with all new config fields at their zero-value defaults (backward
  compatible, matching every prior PZEP in this multi-tenancy track).

## Agent Implementation Plan
- TASK-0001: MT4 — remove `DBForTenant` from `Manager`, adapt tests,
  regenerate mocks. **Done.**
- TASK-0002: MT9 — configurable `PlacementCacheTTL` + negative caching. **Done.**
- TASK-0003: MT10 — scope-cache eviction on `undefined_table` (`42P01`,
  confirmed empirically — see Proposed Solution). **Done.**
- TASK-0004: MT11 — checkout refcounting. **Done.**
- TASK-0005: MT16 (observability half) — structured log points. **Done.**
- TASK-0006: test coverage per Test Plan. **Done** — one dedicated test per
  concern, see Rollout Log.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: `Manager` interface no longer exposes a way to get a raw, unscoped `*sqlx.DB` | TASK-0001 | `TestManager_WithTenantTx_ReusesPoolByClusterAndDB`, `TestManager_DedicatedPoolCapacity` |
| AC-2: placement cache TTL is configurable and actually wired end to end (not just available as an unused constructor) | TASK-0002 | `fxmodule.go`'s `newKVPlacementResolverFromConfig` (code review — see Rollout Log for why no separate integration test) |
| AC-3: a tenant with no placement does not hit the KV store on every request | TASK-0002 | `TestKVPlacementResolver_NotFoundIsNegativelyCached` |
| AC-4: a schema dropped out of band is usable again after one failed call, without a process restart | TASK-0003 | `TestManager_WithTenantTx_RecoversFromSchemaDroppedOutOfBand`, `TestInvalidateScope_EvictsMatchingEntry` |
| AC-5: the idle-pool janitor cannot close a pool a concurrent caller has an active checkout on | TASK-0004 | `TestManager_CloseIdleDedicated_DoesNotCloseCheckedOutPool` |
| AC-6: placement failures and dedicated-pool capacity rejections are logged | TASK-0005 | `TestManager_LogsDedicatedPoolCapacityRejection`, `TestManager_LogsPlacementResolutionFailure` |

## Rollout Log

**2026-07-26, all five items implemented in one pass** (grouped as this PZEP
itself proposed, since they share `managerImpl`'s locking/caching internals):

1. **MT4**: renamed `DBForTenant` → unexported `dbForTenant`; removed from
   `Manager`. Regenerated `mock_Manager.go` (`go tool mockery`) — confirmed
   via `git status` only that one mock file plus the two already-pending
   from Wave 3 changed. Adapted the two tests that called it directly
   through the interface to instead prove behavior through `WithTenantTx`
   (see Proposed Solution's "Implemented as" note).
2. **MT9**: added `Config.PlacementCacheTTL`, wired via a new
   `newKVPlacementResolverFromConfig(kv, cfg)` fx provider (replacing the
   zero-arg `NewKVPlacementResolver` in `fxmodule.go`'s `fx.Provide` list —
   fx resolves `*Config` automatically since it's already in the graph via
   `internal/backoffice/tenant_db_bootstrap.go`). Added negative caching
   (`notFound map[string]time.Time`, capped at 15s) to `KVPlacementResolver`.
   No separate "config-wiring integration test" was written (AC-2) beyond
   code review, since exercising `internal/backoffice`'s full fx graph in a
   test is disproportionate for a config plumbing check — the existing
   `pkg/pdtenantdb` unit tests cover the resolver's own TTL/negative-cache
   behavior directly.
3. **MT10**: **the actual SQLSTATE is `42P01` (`undefined_table`), not
   `3F000`** — this was an incorrect guess in the Draft, caught by testing
   directly against the docker dev Postgres container *before* writing any
   eviction code (`docker exec podzone-postgres psql ...` — `SET LOCAL
   search_path` to a dropped schema succeeds silently; only the next
   unqualified table reference fails, as `42P01`). Added the
   `to_regnamespace` confirmation step (also verified empirically) so an
   unrelated missing-table bug in an otherwise-intact schema doesn't
   trigger a full re-bootstrap. `migrations.InvalidateScope` added and
   wired via `Config.OnScopeInvalidated`.
4. **MT11**: checkout/release implemented with the simplified
   whole-`WithTenantTx`-lifetime scope described in Proposed Solution's
   "Implemented as" note.
5. **MT16**: four log points added (pool opened, pool closed, placement
   resolution failure, dedicated-pool capacity rejection) using
   `Config.Logger` (defaults to `pdlog.NopLogger` when unset — decided
   "optional, nil-safe" over "required", since several existing
   `NewManager` call sites in tests don't pass one). Wired `pdlog.Logger`
   into `internal/backoffice/tenant_db_bootstrap.go`'s `newTenantDBConfig`
   (already had `pdlog.Module` in the fx graph via `cmd/backoffice/main.go`,
   so no new wiring was needed beyond adding the parameter).

Verification: `go build ./...`, `go vet ./...`, and the **entire repo's**
`go test ./...` all pass (not just the touched packages). `pkg/pdtenantdb`
specifically also run with `-race`, clean. `golangci-lint run` on
`pkg/pdtenantdb`, `internal/backoffice`, `internal/onboarding` — 0 new
issues (34 pre-existing findings, all outside files this PZEP touched).

Not done: no commit yet; no rollout to any real environment (this PZEP has
no feature flag of its own — MT4/MT10/MT11/MT16 are always-on correctness
fixes, MT9's `PlacementCacheTTL` is a config value that defaults to
today's behavior at zero).

## Open Questions
- ~~MT16's cache-hit/miss logging~~ — resolved during implementation:
  dropped entirely, only misses/failures are logged (no per-request hit
  spam risk).
- ~~MT4: exact mechanism for the two adapted tests~~ — resolved: wrapping
  `SQLXOpen` to count real pool opens, see Rollout Log.
- ~~MT10: `3F000` vs `42P01`~~ — resolved empirically: `42P01`, plus a
  `to_regnamespace` confirmation step. See Rollout Log.
- ~~Whether `managerImpl` should require a non-nil `pdlog.Logger`~~ —
  resolved: kept optional, defaults to `pdlog.NopLogger`.
