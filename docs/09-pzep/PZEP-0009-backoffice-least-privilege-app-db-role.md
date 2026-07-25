# PZEP-0009: Backoffice tenant DB — separate least-privilege app role from the admin/bootstrap role (Phase 1 of MT1)

## Status
Implementing — code + migration + tests done (TASK-0001/0002/0004); not
rolled out to any real environment (TASK-0003's docker dev config comment
and the Rollout section below are still to do).

## Date
2026-07-25

## Related Commit
(none yet — Phase 1 below is implemented and tested against a real
Postgres via testcontainers, not applied to any persistent/staging cluster)

## Requirement Sources
- Business: recovery/hardening track following the 2026-07-23 multi-tenancy
  audit — see `docs/07-problems/multitenancy-2026-07-23.md` MT1.
- Feature: none new — this is a security-hardening PZEP for existing scope,
  not a product feature.
- Use Cases: a backoffice query-path connection can no longer, by itself,
  read/write another tenant's schema or database even if a `search_path`
  bug (MT6), an application bug, or a future SQL-injection defect lets a
  query escape its intended schema.
- Functional/Non-functional Requirements: `docs/00-governance/twelve-factor.md`
  (config via environment); no SRS behavior change.
- Acceptance Criteria: see below.
- UI Specs: none — backend/infra only.

## Summary

MT1 (critical, `docs/07-problems/multitenancy-2026-07-23.md`) found that
backoffice connects to Postgres with the same admin/bootstrap credential for
**everything**: `CREATE SCHEMA`, running migrations, *and* every ordinary
tenant-scoped query. Tenant isolation currently rests entirely on the
application always setting `search_path` correctly (MT4/MT6) — there is no
database-level backstop.

Full per-tenant roles + row-level security (RLS) is a large, cross-service
change (touches onboarding's provisioning flow, every backoffice table, and
needs load-tested rollout) and is **not** what this PZEP proposes. This
PZEP proposes **Phase 1 only**: split the single admin credential into two
roles used for two different purposes —

- **admin/bootstrap role** (today's credential, unchanged): `CREATE SCHEMA`,
  runs migrations. Used only by `pdtenantdb.Config.OnSchemaBootstrap` (see
  MT8's fix, already landed — `internal/backoffice/migrations.ApplyToSchema`).
- **app role** (new, least-privilege): used for every `WithTenantTx` query
  connection. Can `SELECT`/`INSERT`/`UPDATE`/`DELETE` on tables inside tenant
  schemas, cannot `CREATE`/`DROP` a schema, cannot create roles, is not
  superuser.

This closes the specific MT1 finding ("every ordinary query runs with admin
rights") without needing per-tenant roles or RLS. Per-tenant roles/RLS
remain a real Phase 2, deliberately deferred (see Non-Goals) — MT1 itself is
not fully closed by this PZEP, only its highest-severity sub-finding is.

## Problem

`pkg/pdtenantdb.ClusterConfig` (via `ClusterRegistry`) holds one credential
per cluster, used by `buildDSN` for every connection `Manager` opens —
bootstrap and query connections are indistinguishable. If a query-path bug
ever executes an unintended DDL statement, or a `search_path` mistake (MT6)
lets a query reach an unexpected schema, the connection has full rights to
do real damage there, because it is the same role that can `CREATE SCHEMA`
anywhere in the cluster.

## Goals (Phase 1 — this PZEP)

- `pkg/pdtenantdb.ClusterConfig` gains an optional second credential
  (`AppUser`/`AppPassword`). `Manager` uses it for all `WithTenantTx` query
  connections when set; falls back to the existing admin credential when
  not set, so this is opt-in and backward compatible with every environment
  that hasn't provisioned the app role yet.
- A new backoffice migration file provisions the least-privilege role,
  idempotently (`DO $$ ... IF NOT EXISTS ... $$`), granting it only:
  `CONNECT` on the relevant database(s); `USAGE` on tenant schemas;
  `SELECT, INSERT, UPDATE, DELETE` on tables in tenant schemas (via
  `ALTER DEFAULT PRIVILEGES` scoped per schema, applied at schema-bootstrap
  time alongside the existing migration run — same `OnSchemaBootstrap` hook).
  Explicitly **not** granted: `CREATEDB`, `CREATEROLE`, `SUPERUSER`, DDL on
  any schema.
- `AppUser`/`AppPassword` live on `pdtenantdb.ClusterConfig`, sourced from
  the KV-stored per-cluster record (`podzone/postgres/clusters/<c>`, via
  `KVClusterRegistry` — not a local YAML file, unlike the pool-sizing
  settings in `TenantDBConfig`/MT2's fix). Documented via the doc comment
  on `ClusterConfig.AppUser` in `pkg/pdtenantdb/cluster_registry.go`;
  whichever process seeds that KV record for a given environment is where
  `app_user`/`app_password` get added when Phase 1 is actually rolled out
  there (not found in this repo's `deployments/` tree — out of scope to
  chase down further here).

## Non-Goals

- Per-tenant Postgres roles (one role per tenant) — Phase 2, needs
  onboarding-side provisioning changes and its own PZEP.
- Row-level security policies — Phase 2, same reason.
- Changing `ModeDatabase` (database-per-tenant) credentials — those
  connections are already isolated by being a separate physical database;
  the admin/app split matters most for `ModeSchema`'s shared DB.
- Rotating or changing the *admin* credential itself.
- Applying this migration to any real staging/production cluster as part of
  this PZEP. The migration file and code are delivered; **running it against
  a real cluster, verifying the app role's grants are sufficient for every
  existing query path, and flipping the config to use it is a separate,
  operator-executed rollout step** — not something to do unverified against
  live infrastructure. See Rollout below.

## Proposed Solution

1. `pkg/pdtenantdb/types.go`: add `AppUser`, `AppPassword` to `ClusterConfig`.
2. `pkg/pdtenantdb/manager.go`: `buildDSN` takes a `role` parameter
   (`admin`/`app`); `getOrCreateDB` opens **two** pools per cluster+db when
   `AppUser` is set (admin pool used only inside `OnSchemaBootstrap`, app
   pool used by `WithTenantTx`) — falls back to one pool (today's behavior)
   when `AppUser` is empty.
3. New migration `internal/backoffice/migrations/sql/00XX_backoffice_app_role.sql`:
   creates the role and default-privilege grants, guarded so it is a no-op
   if the role already exists (safe to run repeatedly, safe to run before
   the role is actually used by any connection).
4. Config: `internal/backoffice/config.TenantDBConfig` gains
   `app_user`/`app_password` (empty by default — opt-in).

## Rollout (operator-executed, outside this PZEP's automated scope)

1. Deploy the code + migration with `app_user` still unset — no behavior
   change, admin role keeps doing everything, exactly like today.
2. Set `app_user`/`app_password` in **one non-production environment**
   first (docker dev or a staging cluster). On next deploy, the migration
   creates the role; `Manager` starts opening a second, app-role pool for
   query connections.
3. Exercise the full backoffice query surface (store/catalog/routing CRUD,
   collection queries, GraphQL resolvers) against that environment and
   confirm nothing fails with a permission-denied error — if something
   does, the migration's grants are missing something and need a follow-up
   migration, not a rollback of this PZEP's code (the fallback-to-admin
   path means unsetting `app_user` immediately un-breaks it either way).
4. Only after step 3 passes clean, roll `app_user` out to production
   config.

## Affected Components
- Backend: `pkg/pdtenantdb`, `internal/backoffice/config`,
  `internal/backoffice/migrations`
- Gateway/Frontend/Worker/Database schema/External Integration: none

## API / DB / Event Contract Changes
- API: none.
- DB: adds one role + default-privilege grants. No table/column changes.
- Event: none.

## Permission Changes
- New Postgres role, scoped as described above. No application-level
  (IAM/JWT) permission change.

## Security Considerations
This *is* the security fix (MT1's highest-severity sub-finding). Residual
risk after this PZEP: still one shared app role across all schema-mode
tenants (schema isolation, not role isolation, remains the tenant boundary)
— full per-tenant isolation is Phase 2. Rollback is cheap and safe (unset
`app_user`, redeploy, back to today's single-role behavior) because Phase 1
deliberately keeps the admin path as the always-available fallback.

## Test Plan
- Unit: `pkg/pdtenantdb` — `buildDSN`/pool-selection logic covered for both
  "AppUser unset" (existing behavior, must not regress) and "AppUser set"
  (opens/uses the app pool) cases.
- Integration: the existing `testkit`-backed repository tests
  (`internal/backoffice/infrastructure/repository/routing`) continue to
  pass with `AppUser` unset (default) — proves no regression.
  `migrations.TestApplyTxCreatesLeastPrivilegeAppRole` (new) runs the real
  migration set against a testcontainers Postgres and asserts:
  `backoffice_app` exists, can log in, is not superuser/createdb/createrole,
  has no password set (per the Rollout section's "operator sets it
  out-of-band" requirement), and — using `SET ROLE` from the admin
  connection to avoid needing a password in the test — genuinely cannot
  `CREATE SCHEMA`.
- Manual QA: the Rollout section above, against a real non-production
  environment — not part of this PZEP's automated verification.

## Agent Implementation Plan
- TASK-0001: `ClusterConfig` + `buildDSN` role split, backward compatible. Done.
- TASK-0002: role-provisioning migration
  (`internal/backoffice/migrations/sql/0016_create_backoffice_app_role.sql`).
  Done, verified against real Postgres (see Test Plan).
- TASK-0003: document where `app_user`/`app_password` get set for a real
  environment (see Proposed Solution) — the KV-seeding process itself is
  outside this repo. Not done (nothing to point to yet).
- TASK-0004: unit + integration test coverage per Test Plan. Done.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: with `AppUser` unset, behavior is byte-for-byte identical to today | TASK-0001 | Existing test suite green, unchanged |
| AC-2: with `AppUser` set, query connections use the app role; schema-bootstrap/migrations still use the admin role | TASK-0001 | Unit test |
| AC-3: the app role cannot `CREATE SCHEMA` or create roles | TASK-0002 | Integration test |
| AC-4: the app role can perform every CRUD operation the existing repository layer needs | TASK-0002 | Integration test against real backoffice tables |

## Open Questions
- Whether the app role should be **shared** across all schema-mode tenants
  (this PZEP's assumption) or the first step toward per-tenant roles should
  happen here instead — deferred to Phase 2's own PZEP once Phase 1 is
  verified in a real environment.
- Exact `ALTER DEFAULT PRIVILEGES` scoping needs verification against a
  real Postgres cluster before the Rollout section's step 2 — the migration
  in TASK-0002 is a starting point, not guaranteed complete on the first try.
