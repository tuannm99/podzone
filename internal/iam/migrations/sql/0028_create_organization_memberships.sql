-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_roles (scope, name, description, is_system)
VALUES
  ('organization', 'organization_root', 'Immutable root authority within an organization', TRUE),
  ('organization', 'organization_admin', 'Administrative authority within an organization', TRUE),
  ('organization', 'organization_viewer', 'Read-only organization access', TRUE)
ON CONFLICT (name) DO UPDATE SET scope = EXCLUDED.scope;

INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('organization:read', 'organization', 'read'),
  ('organization:manage_members', 'organization', 'manage_members'),
  ('organization:manage_iam', 'organization', 'manage_iam')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM iam_roles role
JOIN iam_permissions permission ON permission.name IN (
  'organization:read',
  'organization:manage_members',
  'organization:manage_iam'
)
WHERE role.name IN ('organization_root', 'organization_admin')
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM iam_roles role
JOIN iam_permissions permission ON permission.name = 'organization:read'
WHERE role.name = 'organization_viewer'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS iam_organization_memberships (
  org_id TEXT NOT NULL REFERENCES iam_organizations(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL REFERENCES iam_roles(id),
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, user_id)
);

INSERT INTO iam_organization_memberships (org_id, user_id, role_id, status)
SELECT organization.id, organization.root_user_id, role.id, 'active'
FROM iam_organizations organization
JOIN iam_roles role ON role.name = 'organization_root'
WHERE organization.root_user_id > 0
ON CONFLICT (org_id, user_id) DO UPDATE
SET role_id = EXCLUDED.role_id, status = 'active', updated_at = now();

CREATE INDEX IF NOT EXISTS idx_iam_organization_memberships_user_id
ON iam_organization_memberships(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_organization_memberships_user_id;
DROP TABLE IF EXISTS iam_organization_memberships;
DELETE FROM iam_role_permissions
WHERE role_id IN (
  SELECT id FROM iam_roles
  WHERE name IN ('organization_root', 'organization_admin', 'organization_viewer')
);
DELETE FROM iam_roles
WHERE name IN ('organization_root', 'organization_admin', 'organization_viewer');
DELETE FROM iam_permissions
WHERE name IN ('organization:read', 'organization:manage_members', 'organization:manage_iam');
-- +goose StatementEnd
