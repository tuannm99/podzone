# PZEP-0015: Backoffice tenant DB — per-tenant Postgres roles via `SET LOCAL ROLE` (Phase 2 of MT1)

## Status
Done — implemented and tested against real Postgres 15 (testcontainers) in
this working tree, with `TenantRoleEnabled` defaulting to `false`. Not yet
rolled out to any real environment (docker dev or otherwise) — see Rollout
Log.

## Date
2026-07-25 (implemented 2026-07-26)

## Related Commit
(none yet — implemented in the working tree this session, not committed)

## Requirement Sources
- Business: recovery/hardening track following the 2026-07-23 multi-tenancy
  audit — see `docs/07-problems/multitenancy-2026-07-23.md` MT1, and
  [PZEP-0009](./PZEP-0009-backoffice-least-privilege-app-db-role.md) (Phase 1,
  Done), whose Non-Goals section explicitly deferred this work as Phase 2.
- Feature: none new — security-hardening for existing scope.
- Use Cases: a `search_path` bug (MT6), a schema-bootstrap race, or an
  application defect that lets a query address the wrong tenant's tables can
  no longer succeed, because the transaction's Postgres role has grants on
  only the one tenant schema it was meant to touch — not on every tenant
  schema, which is Phase 1's residual gap (see PZEP-0009 "Security
  Considerations": *"still one shared app role across all schema-mode
  tenants... full per-tenant isolation is Phase 2"*).
- Functional/Non-functional Requirements: `docs/00-governance/twelve-factor.md`
  (config via environment); no SRS behavior change.
- Acceptance Criteria: see below.
- UI Specs: none — backend/infra only.

## Summary

Phase 1 (PZEP-0009) split the single admin credential into an admin/bootstrap
role and a shared `backoffice_app` role used for all tenant query traffic.
That closed "every query runs as Postgres superuser" but left one gap: the
`backoffice_app` role's `ALTER DEFAULT PRIVILEGES` grants apply per-schema at
bootstrap time to **every** tenant schema, so the role itself can read/write
**any** tenant's tables. The only thing stopping a cross-tenant read today is
still the application setting `search_path` correctly — the same
application-level discipline the whole MT1 finding said should not be the
only backstop.

This PZEP proposes Phase 2: give every schema-mode tenant its own Postgres
role, scoped by `GRANT`/`ALTER DEFAULT PRIVILEGES` to **only its own schema**,
and have `WithTenantTx` switch into that role for the duration of each
transaction via `SET LOCAL ROLE` (a same-connection, same-pool role switch —
no new physical connection, no new pool). If a future `search_path` bug ever
resolves a query into the wrong schema, the transaction's *role* still has no
grant on that schema, so Postgres denies it — the database itself becomes the
enforcement boundary MT1 asked for, not just `search_path`.

## Problem

`internal/backoffice/migrations/sql/0016_create_backoffice_app_role.sql`
(Phase 1) grants `backoffice_app` `SELECT/INSERT/UPDATE/DELETE` scoped per
schema, but that migration runs once per schema bootstrap and grants the
*same* role every time — so after N tenants have bootstrapped, `backoffice_app`
has standing grants on all N tenant schemas simultaneously. A single Postgres
role cannot express "this connection may only touch tenant X's schema right
now" — the role's grant set is static, independent of which `search_path` a
particular transaction happens to have set.

Concretely: if `WithTenantTx` ever sets `search_path` to tenant B's schema
while operating on behalf of tenant A's request (a defect in the resolver,
not a hypothetical — MT9's placement-cache staleness or a copy-paste bug in a
future repository method are both plausible causes), the query succeeds. It
should instead fail with `permission denied for schema <B>`.

## Goals

- Every schema-mode tenant gets its own Postgres role, named identically to
  its own schema (schema names already carry a tenant prefix, e.g.
  `t_tenant_dev` — a role and a schema occupy different Postgres namespaces,
  so reusing the name is not a collision and gives the clearest possible
  `\du`/`pg_roles` mapping; a further `t_`-prefixed variant would just read
  as `t_t_tenant_dev`), created idempotently at the same schema-bootstrap
  step Phase 1's role migration already runs at (`OnSchemaBootstrap`, admin
  connection).
- The per-tenant role is granted `USAGE` on its own schema and
  `SELECT/INSERT/UPDATE/DELETE` (via `ALTER DEFAULT PRIVILEGES` scoped to
  that schema) on tables in that schema **only** — no grants on any other
  tenant's schema.
- `backoffice_app` (Phase 1's role) is granted membership in every per-tenant
  role. **Target Postgres is 15** (confirmed against the docker dev
  container), which has no per-`GRANT` `INHERIT` clause — that syntax is a
  PG16+ addition. `backoffice_app` was already created `NOINHERIT` at the
  role level by Phase 1's own migration (`0016_create_backoffice_app_role.sql`),
  which has the equivalent effect for every role it is later granted
  membership in: holding `backoffice_app` alone does not automatically use a
  granted role's privileges — an explicit `SET LOCAL ROLE` is required.
- `pkg/pdtenantdb.Manager.WithTenantTx` issues `SET LOCAL ROLE` (schema name,
  quoted) immediately after the existing `SET LOCAL search_path` statement,
  using the same `QuoteIdent` quoting already used for the schema name
  (`manager.go`) to prevent identifier injection from a malicious or
  malformed schema name.
- Feature-flagged and backward compatible, same rollout shape as Phase 1: a
  `TenantRoleEnabled` bool (config, default `false`) gates whether
  `WithTenantTx` issues `SET LOCAL ROLE` at all. Disabled = today's Phase 1
  behavior, byte-for-byte.
- Closes part of MT16's gap ("no test asserting mutual invisibility") with a
  real integration test — see Test Plan.

## Non-Goals

- **Row-level security (RLS) policies.** The MT1 finding mentions RLS as
  "defence in depth", but this codebase's tenant isolation primitive is
  schema-per-tenant (`ModeSchema`), not one shared table with a `tenant_id`
  column — every tenant's tables physically live in a separate schema
  (confirmed: no tenant table lives in `public`, `internal/backoffice/
  migrations/sql/*.sql` create tables only inside the bootstrapped tenant
  schema). RLS's marginal value over per-tenant roles is for a *shared*
  table read by all tenants through one role — this codebase has none today.
  If a future feature adds a genuinely shared/cross-tenant table (e.g. a
  reporting rollup), RLS should be evaluated then, scoped to that table, not
  spent here.
- `ModeDatabase` (dedicated database per big tenant) — those connections are
  already isolated by being a separate physical database and a distinct
  credential path; per-tenant roles matter for the shared `ModeSchema`
  database. Out of scope here; revisit only if `ModeDatabase` provisioning
  ships (currently blocked — see `docs/07-problems/platform-scaling-2026-07-23.md`
  PS4).
- A proactive backfill migration/job that creates roles for every
  already-provisioned tenant ahead of time. Role creation reuses the existing
  `OnSchemaBootstrap` hook, which already runs (idempotently) on first access
  per schema — see MT8 for why this hook runs on the request path today and
  why that is a separate, already-tracked problem this PZEP does not change.
- Applying this to any real staging/production cluster — code and migration
  only; rollout against live infrastructure is a separate operator action,
  same pattern as PZEP-0009's Rollout section.
- MT5, MT7 (IAM/application-level authorization model) — unrelated boundary,
  tracked separately in the item-6 backlog PZEP.

## Proposed Solution

1. `internal/backoffice/migrations/sql/0018_create_backoffice_tenant_role.sql`:
   a migration that runs inside `OnSchemaBootstrap` (same hook Phase 1's
   `0016_` migration uses). Every migration file in this directory already
   runs once per tenant schema with `search_path` pinned to that schema
   before it executes (`migrations.ApplyToSchema`), and `0016_`'s own
   pattern already reads the target schema back out via `current_schema()`
   at runtime rather than needing an externally-substituted template — this
   migration reuses that same pattern instead of inventing new templating
   infrastructure. Idempotently, using `target_schema := current_schema()`:
   - `CREATE ROLE %I NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE` (role name
     = schema name; guarded by `IF NOT EXISTS` via the same `DO $$ ... $$`
     pattern PZEP-0009 used for `backoffice_app`)
   - `GRANT USAGE ON SCHEMA %I TO %I`
   - `GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO %I`
     (covers tables that already existed before this migration ran, for
     tenants bootstrapped under Phase 1 only)
   - `ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT, INSERT, UPDATE,
     DELETE ON TABLES TO %I` (covers tables created by later migrations)
   - `GRANT %I TO backoffice_app` (plain form — no `WITH INHERIT`, per the
     PG15 note in Goals; guarded by a `pg_auth_members` existence check so a
     re-run is a no-op)

2. `pkg/pdtenantdb/manager.go`: `WithTenantTx`, right after the existing
   `SET LOCAL search_path TO %s`, add (implemented):
   ```go
   if m.cfg.TenantRoleEnabled {
       _, err = tx.ExecContext(ctx, fmt.Sprintf(
           `SET LOCAL ROLE %s`, QuoteIdent(tenantRoleName(pl.SchemaName))))
       ...
   }
   ```
   `tenantRoleName(schemaName string) string { return schemaName }` is kept
   as a named function (not inlined) purely so the migration file and this
   call site cannot silently drift onto different naming conventions later.

3. `pkg/pdtenantdb/types.go`: add `TenantRoleEnabled bool` to `Config`
   (top-level manager config, not per-cluster — this is a rollout switch, not
   a credential).

4. `internal/backoffice/config.TenantDBConfig`: add `tenant_role_enabled`
   (default `false`), threaded through `internal/backoffice/tenant_db_bootstrap.go`.

**Known residual gap, not closed by this PZEP:** Phase 1's migration
(`0016_`) already granted `backoffice_app` *direct* `SELECT/INSERT/UPDATE/
DELETE` on every tenant schema (not via role membership — a direct grant,
which `NOINHERIT` does not gate). This PZEP does not revoke those grants,
because doing so would break the documented safe-rollback path (`Security
Considerations` below): if `0018_`'s migration also revoked `0016_`'s direct
grants, setting `tenant_role_enabled: false` to roll back would leave
`backoffice_app` with **no** access at all, not Phase 1 behavior. Consequence:
a caller that bypasses `WithTenantTx` entirely — the exact hazard
[MT4](../07-problems/multitenancy-2026-07-23.md) already tracks
(`DBForTenant` handing out a raw `*sqlx.DB` with no `search_path`/role
applied) — still runs as bare `backoffice_app` and still has its Phase 1
blanket grants, unaffected by whether Phase 2 is enabled. Closing that
specific gap is MT4's own fix (remove `DBForTenant` from the public
`Manager` interface), not this PZEP's — see
[PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md) Wave 4. What
this PZEP does close is narrower but real: **within** a `WithTenantTx`
transaction, once `SET LOCAL ROLE` has run, Postgres checks the tenant
role's own (single-schema) privileges for the rest of that transaction —
`backoffice_app`'s broader grants do not apply while `current_user` is the
tenant role — so a query that resolves against the wrong schema *inside that
transaction* (MT6's scenario) is denied even though the same connection,
outside that transaction, could have reached it via Phase 1's grants.

## Affected Components
- Backend: `pkg/pdtenantdb`, `internal/backoffice/migrations`,
  `internal/backoffice/config`
- Gateway/Frontend/Worker/Database schema/External Integration: none beyond
  the new per-schema role (not a table/column change)

## API / DB / Event Contract Changes
- API: none.
- DB: one new role per tenant schema, created at bootstrap time. No
  table/column changes.
- Event: none.

## Permission Changes
- New per-tenant Postgres roles as described. No application-level
  (IAM/JWT) permission change — this is a database-layer boundary only.

## Security Considerations
This narrows Phase 1's documented residual risk (PZEP-0009's "Security
Considerations": *"still one shared app role across all schema-mode
tenants"*) for the code path that matters most: `WithTenantTx`, which is how
every repository in this codebase talks to tenant data. **Inside** a
`WithTenantTx` transaction with `TenantRoleEnabled=true`, Postgres-level
privileges are scoped to exactly one tenant schema for the duration of that
transaction, regardless of what `search_path` says — a database-enforced
backstop independent of application correctness for that call path, which is
what MT1's core scenario (a `search_path` mistake, MT6) asked for.

Two residual risks, both deliberate, both already tracked:
- `backoffice_app`'s own Phase 1 grants (direct, not via role membership)
  are **not** revoked by this PZEP — see Proposed Solution's "Known residual
  gap" note. A caller that bypasses `WithTenantTx` (MT4's `DBForTenant`
  hazard) is unaffected by Phase 2 either way. Revoking those grants would
  close this but would also make `tenant_role_enabled: false` an unsafe
  rollback (see below) — not worth trading a safe rollback for a gap that
  has its own dedicated fix (MT4) already queued.
- Role count grows linearly with tenant count (`pg_roles` scales fine for
  this — cheap, no per-role connection or memory cost, since
  `SET LOCAL ROLE` reuses the existing pooled connection).

Rollback is cheap and safe *because* backoffice_app's Phase 1 grants are left
intact: set `tenant_role_enabled: false`, redeploy — `WithTenantTx` stops
issuing `SET LOCAL ROLE` and behavior reverts to Phase 1 exactly (the
per-tenant roles remain in the database, unused, harmless).

## Test Plan
- Unit: `pkg/pdtenantdb` — `WithTenantTx`'s SQL-statement composition tested
  for both `TenantRoleEnabled=false` (must not regress Phase 1 byte-for-byte)
  and `=true` (issues `SET LOCAL ROLE` with the correctly quoted identifier,
  including a case with a schema name containing characters that need
  quoting, to prove `QuoteIdent` is actually applied and not string-formatted
  directly).
- Integration: new test using testcontainers Postgres — provisions two tenant
  schemas + roles via the real migration template, then, using the tenant-A
  role's `SET LOCAL ROLE` inside a transaction, attempts to `SELECT` from a
  tenant-B table and asserts `permission denied`. This is the concrete
  "mutual invisibility" test MT16 flagged as missing.
- Existing `internal/backoffice/infrastructure/repository/*` tests continue
  to pass with `TenantRoleEnabled=false` (default) — proves no regression to
  any current caller.
- Manual QA: same staged rollout process as PZEP-0009 (deploy dark, enable in
  one non-production environment, exercise full query surface, only then
  consider production) — not part of this PZEP's automated verification.

## Agent Implementation Plan
- TASK-0001: `Config.TenantRoleEnabled` + config wiring (backward compatible,
  default off). **Done** — `pkg/pdtenantdb/types.go`, `internal/backoffice/
  config/config.go`, `internal/backoffice/tenant_db_bootstrap.go`.
- TASK-0002: `0018_` per-schema role-provisioning migration template +
  `tenantRoleName` helper. **Done** — `internal/backoffice/migrations/sql/
  0018_create_backoffice_tenant_role.sql`, `pkg/pdtenantdb/manager.go`
  (`tenantRoleName`). Diverged from the original Draft's `t_<schema>` naming
  and `WITH INHERIT FALSE` syntax — see Goals/Proposed Solution, corrected
  during implementation once PG15 (not PG16) was confirmed as the actual
  target.
- TASK-0003: `WithTenantTx` `SET LOCAL ROLE` wiring, gated by
  `TenantRoleEnabled`. **Done** — `pkg/pdtenantdb/manager.go`.
- TASK-0004: unit + integration test coverage per Test Plan, including the
  MT16 mutual-invisibility test. **Done** —
  `pkg/pdtenantdb/manager_test.go` (`TestManager_WithTenantTx_SetsRoleWhenTenantRoleEnabled`),
  `internal/backoffice/migrations/migrations_test.go`
  (`TestApplyTxCreatesPerTenantRoleWithMutualIsolation`). Both run against
  real Postgres 15 via testcontainers, both pass.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: with `TenantRoleEnabled=false`, behavior is byte-for-byte identical to Phase 1 | TASK-0001, TASK-0003 | `TestManager_WithTenantTx_SetsRoleWhenTenantRoleEnabled` + full existing suite green |
| AC-2: with `TenantRoleEnabled=true`, each transaction runs under its tenant's own role | TASK-0003 | `TestManager_WithTenantTx_SetsRoleWhenTenantRoleEnabled` |
| AC-3: tenant A's role cannot read or write tenant B's schema | TASK-0002 | `TestApplyTxCreatesPerTenantRoleWithMutualIsolation` |
| AC-4: the per-tenant role can perform every CRUD operation the existing repository layer needs | TASK-0002 | `TestApplyTxCreatesPerTenantRoleWithMutualIsolation` |
| AC-5: `backoffice_app` is a member of each per-tenant role it needs to `SET LOCAL ROLE` into | TASK-0002 | `TestApplyTxCreatesPerTenantRoleWithMutualIsolation` |

## Rollout Log

**2026-07-26, implementation (not yet a real-environment rollout):**
1. Confirmed target Postgres is 15 (`docker exec podzone-postgres psql -c
   'SELECT version()'`), which changed the migration's `GRANT` syntax from
   the Draft's `WITH INHERIT FALSE` (PG16+ only) to a plain `GRANT`, relying
   on `backoffice_app`'s existing role-level `NOINHERIT` (set by Phase 1's
   own migration) for the equivalent effect — see Goals.
2. Chose role name = schema name (not a further `t_`-prefixed variant) once
   real schema names were checked directly in docker dev and confirmed
   already tenant-prefixed (`t_tenant_dev`, etc.) — a `t_<schema>` role name
   would have doubled the prefix.
3. Implemented TASK-0001 through TASK-0004, all passing against real
   Postgres 15 via testcontainers (`go test ./pkg/pdtenantdb/...
   ./internal/backoffice/...` — full package suites, not just the new
   tests). `go build ./...`, `go vet`, and
   `golangci-lint run ./pkg/pdtenantdb/... ./internal/backoffice/...` all
   clean for every file touched (golangci-lint's 34 pre-existing findings
   are all in files this PZEP did not touch).
4. During implementation, caught that the Draft's AC-5 ("`backoffice_app`
   alone has no table access") was not actually achievable without revoking
   Phase 1's grants, which would break the documented safe-rollback path —
   corrected AC-5 and added the "Known residual gap" note in Proposed
   Solution rather than silently shipping a PZEP whose acceptance criteria
   didn't match what was actually built. See that note for the MT4 tie-in.

Not done: `tenant_role_enabled` has not been set to `true` anywhere, no
migration has run against docker dev's real tenant schemas, and no manual
QA (Rollout section's original 4-step plan) has happened. This PZEP's Status
is "Done" for the code; the operator rollout is still fully open.

## Open Questions
- Whether per-tenant role creation should move off the request-path
  `OnSchemaBootstrap` hook once MT8 is fixed (migrations off the request
  path) — likely yes, same mechanism, deferred to whichever PZEP implements
  MT8's fix to avoid editing the same call sites twice (same reasoning MT8's
  note already applies to RQ4 step 1).
- Naming convention: role name = schema name assumes schema names are
  already safe/unique Postgres identifiers (true today — see
  `pdsql/schema.go` identifier validation) — no new validation needed, but
  worth a one-line confirmation in code review when implemented.
- Whether `ModeDatabase` tenants should eventually get an equivalent
  "role scoped to just this database" treatment — likely lower priority
  since a dedicated database is already a stronger boundary than a schema;
  revisit only alongside PS4 (dedicated-database provisioning currently
  doesn't exist end-to-end).
- Whether MT4's fix (remove `DBForTenant` from the public `Manager`
  interface) should land before or after this PZEP ships — landing it first
  would close the "known residual gap" noted in Proposed Solution
  completely rather than leaving it as a named, accepted gap; not blocking
  on it here since MT4 is already independently scoped in PZEP-0016 Wave 4.
