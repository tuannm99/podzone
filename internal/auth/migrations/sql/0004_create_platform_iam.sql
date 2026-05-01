-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_platform_roles (
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id BIGINT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role_id)
);

INSERT INTO iam_roles (name, description, is_system)
VALUES
  ('platform_owner', 'Platform-wide administrative access', TRUE),
  ('platform_admin', 'Platform operational access', TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('tenant:create', 'tenant', 'create'),
  ('platform:manage_roles', 'platform', 'manage_roles')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:create',
  'platform:manage_roles'
)
WHERE r.name = 'platform_owner'
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN (
  'tenant:create'
)
WHERE r.name = 'platform_admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_platform_roles (user_id, role_id, status, created_at, updated_at)
SELECT u.id, r.id, 'active', now(), now()
FROM users u
JOIN iam_roles r ON r.name = 'platform_owner'
WHERE u.id = (SELECT id FROM users ORDER BY id ASC LIMIT 1)
ON CONFLICT (user_id, role_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM user_platform_roles
WHERE role_id IN (
  SELECT id FROM iam_roles WHERE name IN ('platform_owner', 'platform_admin')
);

DELETE FROM iam_role_permissions
WHERE role_id IN (
  SELECT id FROM iam_roles WHERE name IN ('platform_owner', 'platform_admin')
);

DELETE FROM iam_permissions
WHERE name IN ('tenant:create', 'platform:manage_roles');

DELETE FROM iam_roles
WHERE name IN ('platform_owner', 'platform_admin');

DROP TABLE IF EXISTS user_platform_roles;
-- +goose StatementEnd
