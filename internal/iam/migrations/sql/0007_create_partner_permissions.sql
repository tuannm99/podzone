-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_permissions (name, resource, action)
VALUES
  ('partner:read', 'partner', 'read'),
  ('partner:manage', 'partner', 'manage')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN ('partner:read', 'partner:manage')
WHERE r.name IN ('tenant_owner', 'tenant_admin', 'tenant_editor')
ON CONFLICT DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name IN ('partner:read')
WHERE r.name = 'tenant_viewer'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM iam_role_permissions
WHERE permission_id IN (
  SELECT id FROM iam_permissions WHERE name IN ('partner:read', 'partner:manage')
);

DELETE FROM iam_permissions
WHERE name IN ('partner:read', 'partner:manage');
-- +goose StatementEnd
