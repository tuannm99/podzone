-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_permissions (name, resource, action)
VALUES ('store:read_all', 'store', 'read_all')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name = 'store:read_all'
WHERE r.name IN ('platform_owner', 'platform_admin')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM iam_role_permissions
WHERE permission_id = (
  SELECT id FROM iam_permissions WHERE name = 'store:read_all'
);

DELETE FROM iam_permissions
WHERE name = 'store:read_all';
-- +goose StatementEnd
