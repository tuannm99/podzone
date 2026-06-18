-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('store:read', 'store', 'read'),
  ('store:create', 'store', 'create')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN ('store:read', 'store:create')
WHERE r.name IN ('tenant_owner', 'tenant_admin', 'tenant_editor')
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name = 'store:read'
WHERE r.name = 'tenant_viewer'
ON CONFLICT DO NOTHING;

INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern)
SELECT policy.id, 'allow', permission.name, '*'
FROM iam_policies policy
JOIN (
  VALUES
    ('managed/tenant_owner', 'store:read'),
    ('managed/tenant_owner', 'store:create'),
    ('managed/tenant_admin', 'store:read'),
    ('managed/tenant_admin', 'store:create'),
    ('managed/tenant_editor', 'store:read'),
    ('managed/tenant_editor', 'store:create'),
    ('managed/tenant_viewer', 'store:read')
) AS access(policy_name, permission_name)
  ON access.policy_name = policy.name
JOIN iam_permissions permission
  ON permission.name = access.permission_name
WHERE NOT EXISTS (
  SELECT 1
  FROM iam_policy_statements statement
  WHERE statement.policy_id = policy.id
    AND statement.effect = 'allow'
    AND statement.action_pattern = permission.name
    AND statement.resource_pattern = '*'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
