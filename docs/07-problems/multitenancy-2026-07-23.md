# Multi-Tenancy (`pkg/pdtenantdb` + `internal/backoffice`) — Known Problems

Identified from code review on 2026-07-23.
Scope: `pkg/pdtenantdb/`, `internal/backoffice/` (tenant middleware, tenancy
runtime, repositories, migrations), plus the `internal/onboarding` code path
that publishes tenant placement and cluster routes into the KV store.

Status as of 2026-07-25: **MT2 and MT8 resolved.** MT1 has a **Phase 1 fix
implemented but not yet rolled out to any real environment** — see
[PZEP-0009](../09-pzep/PZEP-0009-backoffice-least-privilege-app-db-role.md).
All other items below remain **open and unverified against a running
environment** — they were read out of the code, not reproduced. Re-check the
referenced file/line before treating any item as current.

Validation run at review time: `go test ./pkg/pdtenantdb/...
./internal/backoffice/...` — all packages pass. None of the problems below are
caught by the existing test suite; that is itself part of the finding (see
MT16).

Companion note: [Platform scaling and runtime topology, 2026-07-23](./platform-scaling-2026-07-23.md)
covers DB-cluster placement/capacity admission, PgBouncer, and the Kubernetes
topology. MT2 (pool sizing) and MT12 (PgBouncer bypass) each have a
deployment-side counterpart there.

Cross-reference: [Package issues, 2026-07-07](./pkg-2026-07-07.md) K2 covered
`CloseIdleDedicated` not being wired — that one is now **resolved**
(`pkg/pdtenantdb/fxmodule.go:17-44` runs a `DedicatedIdleTTL/2` ticker and the
method is on the `Manager` interface). MT11 below is a *different* defect in
the same function.

---

## Current Flow

```text
JWT(active_tenant_id) -> TenantMiddleware -> auth.GetSession + iam.GetTenantMembership
                              |
                   X-Store-ID +-> tenancy.ResolveRequestScope
                              |     |- readiness.EnsureReady   (runs migrations)
                              |     |- placements.Resolve      (KV podzone/tenants/<t>/placement)
                              |     '- storeAccess.ResolveStore
                              '-> scope.TenantContext / scope.StoreContext

repo -> mgr.WithTenantTx(tenantID) -> resolver(KV) -> registry(KV podzone/postgres/clusters/<c>)
     -> pool -> SET LOCAL search_path
```

What is correct today and should not be regressed:

- `tenant_id` comes **only** from the JWT `active_tenant_id` claim and is
  cross-checked against the auth session and IAM membership
  (`internal/backoffice/tenant_middleware.go:69-80`,
  `internal/backoffice/authz.go:100-138`). There is no tenant header override.
- `store_id` reaching the resolvers comes from the request scope, never from
  GraphQL arguments (`internal/backoffice/controller/graphql/resolver/routing_mapper.go:182-188`);
  the GraphQL schema exposes no `storeId` argument.
- Postgres identifiers are quoted before interpolation
  (`pkg/pdtenantdb/manager.go:293-296`, `pkg/pdsql/schema.go:31`).
- IAM performs real resource-pattern matching on
  `podzone:tenant/<t>/store/<s>` (`internal/iam/domain/interactor/authz_test.go:77-130`).

---

## MT1 — Backoffice connects to Postgres with the admin DSN; there is no database-enforced tenant boundary

**Phase 1 fix implemented, not yet rolled out** — see
[PZEP-0009](../09-pzep/PZEP-0009-backoffice-least-privilege-app-db-role.md).
`pkg/pdtenantdb.ClusterConfig` now supports a separate least-privilege
`AppUser`/`AppPassword` used for all query connections, admin credential
reserved for schema bootstrap/migrations only — opt-in, backward compatible
(falls back to today's single-credential behavior when unset). The
migration that provisions the actual least-privilege role, and switching
any real environment's config to use it, is an operator-executed rollout
step per PZEP-0009 — **not done as part of this fix**. Per-tenant roles and
RLS (this finding's full closure) remain Phase 2, deliberately deferred.

**File:** `internal/onboarding/infrastructure/provisioning/router/placement_route_reader.go:108-155`,
`pkg/pdtenantdb/manager.go:278-291`, `cmd/onboarding/config.yml:44`
**Severity:** Critical — a single application bug exposes every tenant

Onboarding parses its `store_provisioning.admin_dsn` and publishes the raw
host / port / **user / password** into the KV key
`podzone/postgres/clusters/<clusterName>`:

```go
return map[string]interface{}{
    "host":     parsed.Hostname(),
    "port":     port,
    "user":     parsed.User.Username(),
    "password": password,
    "ssl_mode": sslMode,
}, nil
```

`pdtenantdb` builds its runtime DSN verbatim from that payload
(`buildDSN`), so **backoffice runs every tenant query as the Postgres
administrator**. In schema-per-tenant mode the only thing separating tenant A
from tenant B is the application-level `SET LOCAL search_path` in
`WithTenantTx`. There is no per-tenant role, no `REVOKE`/`GRANT` restriction,
and no row-level security as defence in depth. Consequences:

- One missing `WHERE`, one direct `db.QueryContext` (see MT4), or one
  injection point yields cross-tenant read/write.
- The admin credential is stored in plaintext in the KV store, which is
  MongoDB in this deployment (`cmd/backoffice/config.yml:7-10`,
  `kv_store.provider: mongo`, collection `runtime_kv`).

**Fix:** separate the provisioning credential from the runtime credential.
Onboarding should publish a *runtime* role (non-superuser, no `CREATE`
on other schemas) into the cluster route, and keep `admin_dsn` for
provisioning only. Longer term, add `SET LOCAL ROLE` per tenant plus RLS so
the boundary is enforced by Postgres rather than by application discipline.

---

## MT2 — Every schema-mode tenant shares one 10-connection pool

**File:** `internal/backoffice/module.go:31`, `pkg/pdtenantdb/manager.go:47-75`
**Severity:** Critical — hard scaling ceiling, not configurable

`module.go` provides an empty config:

```go
func() *pdtenantdb.Config { return &pdtenantdb.Config{} },
```

so `NewManager` falls back to `MaxOpenConns: 10`, `MaxIdleConns: 2`. Pools are
keyed by `ConnKey{ClusterName, DBName}` (`manager.go:30-33`), and in
schema-per-tenant mode every tenant resolves to the same `DBName`
(`podzone_tenants`). Net effect: **10 Postgres connections shared by all
tenants on the cluster**, with each request holding one for the duration of a
full transaction that also includes the migration check (see MT8).

There is no `pdtenantdb` section in `cmd/backoffice/config.yml`, so none of
`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `MaxDedicatedPools`,
`DedicatedIdleTTL` can be tuned per environment.

**Fix:** bind `pdtenantdb.Config` to koanf like `boconfig.NewConfigFromKoanf`
does, add a `pdtenantdb:` block to `cmd/backoffice/config.yml` and
`config.prod.yml`, and size the shared pool for the expected tenant count.

---

## MT3 — `Config.SharedDB` default disagrees with the database onboarding actually provisions

**File:** `pkg/pdtenantdb/manager.go:48-50`, `cmd/onboarding/config.yml:42`
**Severity:** High — latent; breaks pool accounting the moment `mode: database` is used

`NewManager` defaults `SharedDB` to `"backoffice"`. Onboarding provisions
`db_name: podzone_tenants`. Because backoffice passes an empty `Config`
(MT2), the manager believes the shared database is called `backoffice` while
every placement points at `podzone_tenants`.

Three places branch on this value:

- `manager.go:90` — `dedicated := pl.Mode == ModeDatabase && key.DBName != m.cfg.SharedDB`
- `manager.go:217-233` — `reserveDedicatedLocked` counts pools where `k.DBName != m.cfg.SharedDB`
- `manager.go:248-264` — `CloseIdleDedicated` skips pools where `k.DBName == m.cfg.SharedDB`

Today nothing misbehaves because all placements are `mode: schema`, so
`dedicated` is always `false` and `lastUsed` is never populated. As soon as
one tenant is placed with `mode: database`, the shared `podzone_tenants` pool
starts being counted against `MaxDedicatedPools` and becomes eligible for the
idle janitor.

**Fix:** set `SharedDB` from config (same change as MT2) and make it match
`store_provisioning.db_name`. Consider failing startup when the two disagree.

---

## MT4 — `Manager.DBForTenant` hands out a pool with no `search_path` applied

**File:** `pkg/pdtenantdb/manager.go:19-26,77-112,145-153`
**Severity:** High — silent wrong-tenant reads if ever used

`DBForTenant` is part of the public `Manager` interface and returns a raw
`*sqlx.DB`. The tenant `search_path` is set only inside `WithTenantTx`:

```go
if pl.Mode == ModeSchema && pl.SchemaName != "" {
    _, err = tx.ExecContext(ctx, fmt.Sprintf(
        `SET LOCAL search_path TO %s, public`, pgQuoteIdent(pl.SchemaName)))
```

Any caller that takes the `*sqlx.DB` and queries it directly runs against
`public`, not the tenant schema — no error, just wrong (or missing) data.
No production caller does this today (all repositories go through
`WithTenantTx`), which is exactly why the hazard is easy to trip over later.

**Fix:** remove `DBForTenant` from the `Manager` interface (keep it
unexported for `WithTenantTx`'s use), or return a wrapper type that pins the
schema on every checkout.

---

## MT5 — `RequiredStoreScope` lets a caller-supplied store ID override the authorized scope

**File:** `internal/backoffice/domain/routing/support.go:12-21`,
`internal/backoffice/tenant_middleware.go:232-238`
**Severity:** High — latent store-level privilege escalation

```go
func RequiredStoreScope(ctx context.Context, explicitStoreID string) (string, error) {
	storeID := strings.TrimSpace(explicitStoreID)
	if storeID == "" {
		storeID = strings.TrimSpace(scope.CurrentStoreID(ctx))
	}
	...
}
```

The explicit value wins and is **never compared** against the scope that the
middleware authorized. Meanwhile the IAM resource string is derived from the
request scope, not from the store actually operated on:

```go
func resourceForPermission(ctx context.Context, tenantID string) string {
	storeID := strings.TrimSpace(scope.CurrentStoreID(ctx))
	...
	return "podzone:tenant/" + strings.TrimSpace(tenantID) + "/store/" + storeID
}
```

That is the shape of "check permission on store A, act on store B" within a
tenant. It is **not exploitable today** — no GraphQL input carries a
`storeId`, so every `explicitStoreID` is already the scope value — but the
domain contract permits it and the next DTO that adds a store field opens it.

**Fix:** make the mismatch an error rather than a silent override — if
`explicitStoreID != ""` and differs from `scope.CurrentStoreID(ctx)`, return
`ErrStoreScopeRequired` (or a dedicated mismatch error).

---

## MT6 — Tenant `search_path` includes `public`

**File:** `pkg/pdtenantdb/manager.go:145-153`
**Severity:** High — silent fallthrough to shared tables

`SET LOCAL search_path TO "<schema>", public` means a table missing from the
tenant schema resolves to a same-named table in `public` instead of raising
`relation does not exist`. Given that migrations are applied per-schema at
runtime (MT8), a partially-migrated tenant can read or write the wrong table
without any error surfacing. Keeping `public` on the search path of an admin
connection is also a well-known Postgres privilege-escalation surface
(function/operator shadowing).

**Fix:** `SET LOCAL search_path TO "<schema>"` only, or `"<schema>",
pg_catalog` if a catalog fallback is needed.

---

## MT7 — Store visibility is gated on `owner_id` instead of IAM, so non-owner tenant members cannot use a store

**File:** `internal/backoffice/infrastructure/repository/store/repository.go:47-55,100-106,175-186`,
`internal/backoffice/runtime/storeaccess/access.go:25-38`
**Severity:** High — multi-user tenants do not work

Every store read and write filters by the calling user:

```go
predicates = append([]sq.Sqlizer{sq.Eq{"owner_id": ownerID}}, predicates...)
```

`storeaccess.ResolveStore` goes through `repo.FindByID`, so a tenant member
who is *not* the store owner fails scope resolution in the middleware and gets
an error for every store-scoped request — even when IAM has granted them
`store:read` on exactly that store. Backoffice therefore runs two conflicting
authorization models: IAM policy statements, and an `owner_id` column that
silently overrides them.

**Fix:** pick IAM as the single authority. Drop the `owner_id` predicate from
`FindByID`/`UpdateStatus`, keep `owner_id` as informational metadata, and let
`RequirePermission` on `podzone:tenant/<t>/store/<s>` decide access.
`FindPage` needs a listing policy decision (all stores the caller has
`store:read` on) — that requires an IAM listing capability, so flag it before
implementing.

---

## MT8 — Migrations run on the request path, inside the business transaction

**File:** `internal/backoffice/runtime/tenancy/runtime.go:55-60`,
`internal/backoffice/migrations/migrations.go:26-48`,
`internal/backoffice/infrastructure/repository/store/repository.go:59`,
`.../catalog/repository.go:87`, `.../routing/repository.go:157`
**Severity:** High — DDL locks taken by end-user requests

Two separate costs on every request:

1. `ResolveRequestScope` calls `readiness.EnsureReady`, which opens its own
   transaction and runs `migrations.ApplyTx` — once per GraphQL operation.
2. Each repository method calls `ensureStoreTables` / `ensureProductSetupTables`
   / `ensureRoutedOrderTables`, all of which are `migrations.ApplyTx` **inside
   the business transaction**.

Even on the warm path (`appliedScopes` hit) each call still costs a
`SELECT current_database(), current_schema()` round trip. On the cold path, or
the first request after a deploy that adds a migration, real DDL executes
inside a user request and holds `pg_advisory_xact_lock` for the whole business
transaction:

```go
if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLockKey(scopeKey)); err != nil {
```

Combined with MT2's 10-connection pool, a slow migration serializes the whole
service.

**Fix:** move schema migration into tenant provisioning (onboarding, at
placement time) and/or an explicit migrate job, and reduce the request path to
a readiness *check*. Remove `ensure*Tables` from repository methods — this is
also RQ4 step 1 in
[Repository and query layer, 2026-07-24](./repository-query-layer-2026-07-24.md);
do the two together rather than editing the same 20 call sites twice.

---

## MT9 — Placement cache cannot be invalidated

**File:** `pkg/pdtenantdb/resolver_kv.go:40-51,53-119`
**Severity:** Medium — stale routing window after a tenant move

`NewKVPlacementResolver` hardcodes a 2-minute TTL and caches in-process with
no invalidation channel. After a tenant is migrated to a different
schema/database/cluster, up to two minutes of writes still land on the old
placement — and every replica has its own independent expiry, so the window is
per-instance. There is also no negative caching: a tenant with no placement
key hits the KV store on every single request.

**Fix:** make the TTL configurable, add a KV-watch or event-driven
invalidation on `podzone/tenants/<t>/placement`, and short-negative-cache
`ErrPlacementNotFound`.

---

## MT10 — `bootstrapped` and `appliedScopes` are process-global maps that are never invalidated

**File:** `pkg/pdtenantdb/manager.go:42-44,114-125`,
`internal/backoffice/migrations/migrations.go:23,105-107`
**Severity:** Medium — recovery requires a process restart

`ensureSchema` records `cluster|db|schema` in a `sync.Map` and skips
`CREATE SCHEMA IF NOT EXISTS` forever after. `migrations.appliedScopes` does
the same for `database|schema`. If a schema is dropped or restored out of band,
the process keeps believing it is present and every query fails until restart.
`appliedScopes` being a package-level global also leaks state between tests in
the same binary.

**Fix:** give both caches a TTL, or re-validate on the first error rather than
caching a permanent "done" bit.

---

## MT11 — Idle-pool janitor closes pools with no reference counting

**File:** `pkg/pdtenantdb/manager.go:241-264`
**Severity:** Medium — rare `sql: database is closed` on in-flight requests

`CloseIdleDedicated` calls `db.Close()` while holding only `m.mu`. A caller
that has already returned from `DBForTenant` (which released the lock) but has
not yet called `BeginTxx` will fail. The window is narrow because `markUsed`
runs immediately before the return, but there is no refcount or lease to make
it impossible.

Note this is a different defect from K2 in
[Package issues, 2026-07-07](./pkg-2026-07-07.md), which was about the janitor
never running at all and is now resolved.

**Fix:** refcount checkouts, or have `WithTenantTx` re-acquire and pin the
pool under the lock for the duration of `BeginTxx`.

---

## MT12 — PgBouncer is documented in the placement strategy but bypassed at runtime

**File:** `pkg/pdtenantdb/manager.go:144`,
`internal/onboarding/infrastructure/provisioning/provider/provider.go:200-215`,
`internal/onboarding/infrastructure/provisioning/router/placement_route_reader.go:129-155`
**Severity:** Medium — documented pooling strategy is not the real one

`provider.go` records `"pgbouncer_service": "pgbouncer"` in the placement
metadata and `manager.go` carries the comment *"IMPORTANT: use SET LOCAL for
PgBouncer transaction pooling"*, but the cluster route published to KV is
derived from `admin_dsn`, i.e. the Postgres host directly. Nothing in the
connection path goes through PgBouncer. The `SET LOCAL` discipline is correct
and worth keeping regardless, but capacity planning based on the documented
architecture will be wrong.

**Fix:** either publish the PgBouncer endpoint as the cluster route host, or
correct the metadata and comments to state that connections are direct.

---

## MT13 — Internal store-bootstrap endpoint accepts an arbitrary workspace ID behind a single shared token

**File:** `internal/backoffice/controller/httphandler/store_bootstrap.go:34-54`
**Severity:** Medium — blast radius of one leaked token is every tenant

```go
requestCtx := toolkit.WithTenantID(ctx.Request.Context(), strings.TrimSpace(request.WorkspaceID))
requestCtx = toolkit.WithUserID(requestCtx, strings.TrimSpace(request.OwnerID))
```

The tenant is taken straight from the request body, authenticated only by a
static `X-Backoffice-Service-Token` shared across all callers. Anyone holding
it can write into any tenant's database. Secondary issue: every failure —
including validation failures from the usecase — is returned as
`503 Service Unavailable`, which misleads callers and retry logic.

**Fix:** scope the credential (per-caller token or mTLS/service identity),
and map errors to accurate status codes.

---

## MT14 — IAM permission check issues a gRPC call per resolved field

**File:** `internal/backoffice/tenant_middleware.go:110-145,76-80`
**Severity:** Medium — N+1 network calls per request

`InterceptField` calls `authorizer.RequirePermission` for every resolver field
with a permission mapping, with no per-request memoization. A query touching
several mapped fields makes one `CheckPermission` gRPC round trip each, on top
of the two calls (`GetSession`, `GetTenantMembership`) that
`InterceptOperation` already makes.

**Fix:** memoize `(userID, tenantID, permission, resource)` decisions in the
request context for the lifetime of the operation.

---

## MT15 — Duplicate migration number `0015`

**File:** `internal/backoffice/migrations/sql/0015_add_store_scope_to_product_setup.sql`,
`internal/backoffice/migrations/sql/0015_create_customer_order_aggregate_store.sql`
**Severity:** Low — currently deterministic, but fragile

`migrations.go:71-73` sorts by the full filename and the recorded `version` is
also the full filename, so there is no primary-key collision and the order is
stable (`..._add_store_scope...` before `..._create_customer_order...`). It
still breaks the numbering invariant every other tool and reader assumes.
Already noted in
[services/backoffice/db-design.md](../03-architecture-detail-design/services/backoffice/db-design.md).

**Fix:** renumber the later file to `0016_`. Note that renaming changes the
recorded `version` string, so environments that already applied it need a
manual `backoffice_schema_migrations` update — do not rename without a
migration plan.

---

## MT16 — No tenancy observability, and no test covers the isolation boundary

**File:** `pkg/pdtenantdb/` (whole package), `internal/backoffice/runtime/`
**Severity:** Medium — failures are undiagnosable in production

Nothing emits metrics or structured logs for: open pool count per cluster,
placement resolution failures, placement cache hit rate, dedicated-pool
capacity rejections (`ErrDedicatedPoolCapacity`), or per-tenant query latency.
When a tenant reports "my data is missing", there is no signal to distinguish a
stale placement from a wrong `search_path` from a failed migration.

Equally, the passing test suite covers none of the boundaries above: there is
no test asserting that tenant A's transaction cannot see tenant B's rows, and
none asserting that a mismatched explicit store ID (MT5) is rejected.

**Fix:** add pool/resolution metrics to `pdtenantdb`, and add an integration
test that provisions two tenants against a real Postgres and asserts mutual
invisibility.

---

## Suggested Order Of Work

1. **MT1** — split runtime credentials from the admin DSN. This closes the
   whole class that MT4 and MT6 belong to.
2. **MT2 / MT3** — bind `pdtenantdb.Config` to configuration; set
   `shared_db: podzone_tenants`; size the pool.
3. **MT8** — get migrations off the request path.
4. **MT5 / MT7** — settle on IAM as the single store-authorization model and
   tighten `RequiredStoreScope`.

MT1 and MT7 change contracts and behaviour, so each needs its own sprint slice
or PZEP before implementation — do not code them straight from this note.
