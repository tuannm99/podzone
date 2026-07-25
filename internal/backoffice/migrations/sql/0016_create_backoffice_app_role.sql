-- Least-privilege role for ordinary backoffice tenant query connections,
-- kept separate from the admin/bootstrap credential that applies this
-- migration set. See docs/09-pzep/PZEP-0009-backoffice-least-privilege-app-db-role.md
-- and docs/07-problems/multitenancy-2026-07-23.md MT1.
--
-- No password is set here on purpose - do not hardcode a credential in a
-- checked-in migration file. The role exists but cannot authenticate until
-- an operator runs `ALTER ROLE backoffice_app WITH PASSWORD '<real secret>'`
-- out of band, per PZEP-0009's Rollout section. Only after that, and after
-- setting ClusterConfig.AppUser/AppPassword in config, does Manager start
-- using this role for query connections.
--
-- Role creation is cluster-wide, but (like every file in this directory)
-- this migration runs once per tenant schema bootstrap - the existence
-- check below makes every run after the first a safe no-op. The grants
-- below are scoped to current_schema(), which is always the schema this
-- particular bootstrap call is for (the tenant schema in schema mode, or
-- "public" in database mode, where each dedicated DB's own tables live).
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'backoffice_app') THEN
        CREATE ROLE backoffice_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
    END IF;
END
$$;

DO $$
DECLARE
    db_name text := current_database();
BEGIN
    EXECUTE format('GRANT CONNECT ON DATABASE %I TO backoffice_app', db_name);
END
$$;

DO $$
DECLARE
    target_schema text := current_schema();
BEGIN
    EXECUTE format('GRANT USAGE ON SCHEMA %I TO backoffice_app', target_schema);
    EXECUTE format(
        'GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO backoffice_app',
        target_schema
    );
    EXECUTE format(
        'ALTER DEFAULT PRIVILEGES IN SCHEMA %I GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO backoffice_app',
        target_schema
    );
END
$$;
