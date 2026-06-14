-- +goose Up
-- +goose StatementBegin
INSERT INTO iam_permissions (name, resource, action)
VALUES ('store:approve', 'store', 'approve')
ON CONFLICT (name) DO NOTHING;

INSERT INTO iam_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM iam_roles r
JOIN iam_permissions p ON p.name = 'store:approve'
WHERE r.name IN ('platform_owner', 'platform_admin')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM iam_role_permissions
WHERE permission_id = (
  SELECT id FROM iam_permissions WHERE name = 'store:approve'
);

DELETE FROM iam_permissions
WHERE name = 'store:approve';
-- +goose StatementEnd
