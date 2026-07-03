-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('platform:read_infrastructure', 'platform', 'read_infrastructure'),
  ('platform:manage_infrastructure', 'platform', 'manage_infrastructure')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT role.id, permission.id
FROM iam_roles role
JOIN iam_permissions permission
  ON permission.name IN (
    'platform:read_infrastructure',
    'platform:manage_infrastructure'
  )
WHERE role.name IN ('platform_owner', 'platform_admin')
ON CONFLICT DO NOTHING;

INSERT INTO iam_policy_statements (policy_id, effect, action_pattern, resource_pattern)
SELECT policy.id, 'allow', permission.name, '*'
FROM iam_policies policy
JOIN iam_permissions permission
  ON permission.name IN (
    'platform:read_infrastructure',
    'platform:manage_infrastructure'
  )
WHERE policy.name IN ('managed/platform_owner', 'managed/platform_admin')
  AND NOT EXISTS (
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
DELETE FROM iam_policy_statements statement
USING iam_policies policy
WHERE statement.policy_id = policy.id
  AND policy.name IN ('managed/platform_owner', 'managed/platform_admin')
  AND statement.action_pattern IN (
    'platform:read_infrastructure',
    'platform:manage_infrastructure'
  );

DELETE FROM iam_role_permissions
WHERE permission_id IN (
  SELECT id
  FROM iam_permissions
  WHERE name IN (
    'platform:read_infrastructure',
    'platform:manage_infrastructure'
  )
);

DELETE FROM iam_permissions
WHERE name IN (
  'platform:read_infrastructure',
  'platform:manage_infrastructure'
);
-- +goose StatementEnd
