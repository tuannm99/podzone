-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_policies (
  id BIGSERIAL PRIMARY KEY,
  scope TEXT NOT NULL,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  is_system BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_policy_statements (
  id BIGSERIAL PRIMARY KEY,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  effect TEXT NOT NULL,
  action_pattern TEXT NOT NULL,
  resource_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_role_policy_attachments (
  role_id BIGINT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (role_id, policy_id)
);

CREATE TABLE IF NOT EXISTS iam_user_policy_attachments (
  user_id BIGINT NOT NULL,
  scope TEXT NOT NULL,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, scope, policy_id)
);

CREATE TABLE IF NOT EXISTS iam_tenant_user_policy_attachments (
  tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id, policy_id)
);

INSERT INTO iam_policies (scope, name, description, is_system)
VALUES
  ('tenant', 'managed/tenant_owner', 'Managed policy for tenant owner role', TRUE),
  ('tenant', 'managed/tenant_admin', 'Managed policy for tenant admin role', TRUE),
  ('tenant', 'managed/tenant_editor', 'Managed policy for tenant editor role', TRUE),
  ('tenant', 'managed/tenant_viewer', 'Managed policy for tenant viewer role', TRUE),
  ('platform', 'managed/platform_owner', 'Managed policy for platform owner role', TRUE),
  ('platform', 'managed/platform_admin', 'Managed policy for platform admin role', TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern)
SELECT DISTINCT p.id, 'allow', perm.name, '*'
FROM iam_policies p
JOIN iam_roles r ON (
  (p.name = 'managed/tenant_owner' AND r.name = 'tenant_owner') OR
  (p.name = 'managed/tenant_admin' AND r.name = 'tenant_admin') OR
  (p.name = 'managed/tenant_editor' AND r.name = 'tenant_editor') OR
  (p.name = 'managed/tenant_viewer' AND r.name = 'tenant_viewer') OR
  (p.name = 'managed/platform_owner' AND r.name = 'platform_owner') OR
  (p.name = 'managed/platform_admin' AND r.name = 'platform_admin')
)
JOIN iam_role_permissions rp ON rp.role_id = r.id
JOIN iam_permissions perm ON perm.id = rp.permission_id
WHERE NOT EXISTS (
  SELECT 1
  FROM iam_policy_statements existing
  WHERE existing.policy_id = p.id
    AND existing.effect = 'allow'
    AND existing.action_pattern = perm.name
    AND existing.resource_pattern = '*'
);

INSERT INTO iam_role_policy_attachments (role_id, policy_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_policies p ON (
  (r.name = 'tenant_owner' AND p.name = 'managed/tenant_owner') OR
  (r.name = 'tenant_admin' AND p.name = 'managed/tenant_admin') OR
  (r.name = 'tenant_editor' AND p.name = 'managed/tenant_editor') OR
  (r.name = 'tenant_viewer' AND p.name = 'managed/tenant_viewer') OR
  (r.name = 'platform_owner' AND p.name = 'managed/platform_owner') OR
  (r.name = 'platform_admin' AND p.name = 'managed/platform_admin')
)
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_tenant_user_policy_attachments;
DROP TABLE IF EXISTS iam_user_policy_attachments;
DROP TABLE IF EXISTS iam_role_policy_attachments;
DROP TABLE IF EXISTS iam_policy_statements;
DROP TABLE IF EXISTS iam_policies;
-- +goose StatementEnd
