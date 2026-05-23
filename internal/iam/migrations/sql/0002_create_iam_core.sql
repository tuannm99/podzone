-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_roles (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  is_system BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_permissions (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  resource TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_role_permissions (
  role_id BIGINT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  permission_id BIGINT NOT NULL REFERENCES iam_permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS tenants (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tenant_memberships (
  tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id BIGINT NOT NULL REFERENCES iam_roles(id),
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id)
);

INSERT INTO iam_roles (name, description, is_system)
VALUES
  ('tenant_owner', 'Full access within tenant', TRUE),
  ('tenant_admin', 'Administrative access within tenant', TRUE),
  ('tenant_editor', 'Edit business resources within tenant', TRUE),
  ('tenant_viewer', 'Read-only access within tenant', TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('tenant:read', 'tenant', 'read'),
  ('tenant:manage_members', 'tenant', 'manage_members'),
  ('store:read', 'store', 'read'),
  ('store:create', 'store', 'create'),
  ('store:update', 'store', 'update'),
  ('store:activate', 'store', 'activate'),
  ('store:deactivate', 'store', 'deactivate'),
  ('store_config:read', 'store_config', 'read'),
  ('store_config:update', 'store_config', 'update')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:read',
  'tenant:manage_members',
  'store:read',
  'store:create',
  'store:update',
  'store:activate',
  'store:deactivate',
  'store_config:read',
  'store_config:update'
)
WHERE r.name = 'tenant_owner'
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:read',
  'store:read',
  'store:create',
  'store:update',
  'store:activate',
  'store:deactivate',
  'store_config:read',
  'store_config:update'
)
WHERE r.name = 'tenant_admin'
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:read',
  'store:read',
  'store:create',
  'store:update',
  'store_config:read',
  'store_config:update'
)
WHERE r.name = 'tenant_editor'
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:read',
  'store:read',
  'store_config:read'
)
WHERE r.name = 'tenant_viewer'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tenant_memberships;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS iam_role_permissions;
DROP TABLE IF EXISTS iam_permissions;
DROP TABLE IF EXISTS iam_roles;
-- +goose StatementEnd
