-- +goose Up
-- +goose StatementBegin
ALTER TABLE iam_roles
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'tenant';

UPDATE iam_roles
SET scope = 'platform'
WHERE name IN ('platform_owner', 'platform_admin');

UPDATE iam_roles
SET scope = 'tenant'
WHERE name IN ('tenant_owner', 'tenant_admin', 'tenant_editor', 'tenant_viewer');

CREATE TABLE IF NOT EXISTS iam_role_trust_statements (
  id BIGSERIAL PRIMARY KEY,
  role_id BIGINT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  effect TEXT NOT NULL,
  principal_type TEXT NOT NULL,
  principal_pattern TEXT NOT NULL,
  tenant_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_iam_role_trust_statements_role_id
ON iam_role_trust_statements(role_id);

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'platform_role', 'platform_owner', '*'
FROM iam_roles r
WHERE r.name = 'platform_owner'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'platform_role'
      AND t.principal_pattern = 'platform_owner'
      AND t.tenant_pattern = '*'
  );

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'platform_role', role_name, '*'
FROM iam_roles r
JOIN (VALUES ('platform_owner'), ('platform_admin')) AS src(role_name) ON TRUE
WHERE r.name = 'platform_admin'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'platform_role'
      AND t.principal_pattern = src.role_name
      AND t.tenant_pattern = '*'
  );

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'tenant_role', role_name, '*'
FROM iam_roles r
JOIN (VALUES ('tenant_owner')) AS src(role_name) ON TRUE
WHERE r.name = 'tenant_owner'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'tenant_role'
      AND t.principal_pattern = src.role_name
      AND t.tenant_pattern = '*'
  );

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'tenant_role', role_name, '*'
FROM iam_roles r
JOIN (VALUES ('tenant_owner'), ('tenant_admin')) AS src(role_name) ON TRUE
WHERE r.name = 'tenant_admin'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'tenant_role'
      AND t.principal_pattern = src.role_name
      AND t.tenant_pattern = '*'
  );

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'tenant_role', role_name, '*'
FROM iam_roles r
JOIN (VALUES ('tenant_owner'), ('tenant_admin'), ('tenant_editor')) AS src(role_name) ON TRUE
WHERE r.name = 'tenant_editor'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'tenant_role'
      AND t.principal_pattern = src.role_name
      AND t.tenant_pattern = '*'
  );

INSERT INTO iam_role_trust_statements (role_id, effect, principal_type, principal_pattern, tenant_pattern)
SELECT r.id, 'allow', 'tenant_role', role_name, '*'
FROM iam_roles r
JOIN (VALUES ('tenant_owner'), ('tenant_admin'), ('tenant_editor'), ('tenant_viewer')) AS src(role_name) ON TRUE
WHERE r.name = 'tenant_viewer'
  AND NOT EXISTS (
    SELECT 1 FROM iam_role_trust_statements t
    WHERE t.role_id = r.id
      AND t.principal_type = 'tenant_role'
      AND t.principal_pattern = src.role_name
      AND t.tenant_pattern = '*'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_role_trust_statements_role_id;
DROP TABLE IF EXISTS iam_role_trust_statements;
ALTER TABLE iam_roles
DROP COLUMN IF EXISTS scope;
-- +goose StatementEnd
