-- Per-tenant least-privilege role (Phase 2 of MT1) - scoped to exactly one
-- tenant schema, unlike backoffice_app (Phase 1), whose grants apply to
-- every tenant schema. See docs/09-pzep/
-- PZEP-0015-backoffice-per-tenant-db-role-phase-2.md and
-- docs/07-problems/multitenancy-2026-07-23.md MT1.
--
-- Role name is the tenant schema name itself (schema names already carry a
-- tenant prefix, e.g. t_tenant_dev - a role and a schema live in different
-- Postgres namespaces, so an identical name is not a collision, and it
-- gives the clearest possible mapping in `\du` / `pg_roles`).
--
-- NOLOGIN: this role is never connected to directly. backoffice_app (the
-- Phase 1 login role, migration 0016) was created NOINHERIT at the role
-- level, so granting it membership here does NOT automatically confer this
-- role's table access - backoffice_app must explicitly `SET LOCAL ROLE`
-- into this role for the duration of a transaction (see
-- pkg/pdtenantdb/manager.go, gated by Config.TenantRoleEnabled). This
-- targets PostgreSQL 15, which has no per-GRANT INHERIT option (that is a
-- PG16+ addition) - the role-level NOINHERIT on backoffice_app already
-- gives the equivalent effect for every role it is a member of.
DO $$
DECLARE
    target_schema text := current_schema();
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = target_schema) THEN
        EXECUTE format('CREATE ROLE %I NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE', target_schema);
    END IF;
END
$$;

DO $$
DECLARE
    target_schema text := current_schema();
BEGIN
    EXECUTE format('GRANT USAGE ON SCHEMA %I TO %I', target_schema, target_schema);
    EXECUTE format(
        'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO %I',
        target_schema, target_schema
    );
    EXECUTE format(
        'ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %I',
        target_schema, target_schema
    );
END
$$;

DO $$
DECLARE
    target_schema text := current_schema();
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_auth_members m
        JOIN pg_roles r ON r.oid = m.roleid
        JOIN pg_roles g ON g.oid = m.member
        WHERE r.rolname = target_schema AND g.rolname = 'backoffice_app'
    ) THEN
        EXECUTE format('GRANT %I TO backoffice_app', target_schema);
    END IF;
END
$$;
