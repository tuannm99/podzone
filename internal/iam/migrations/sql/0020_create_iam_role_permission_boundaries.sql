-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_role_permission_boundaries (
    role_id BIGINT PRIMARY KEY REFERENCES iam_roles(id) ON DELETE CASCADE,
    policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_iam_role_permission_boundaries_policy_id
ON iam_role_permission_boundaries(policy_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_role_permission_boundaries_policy_id;
DROP TABLE IF EXISTS iam_role_permission_boundaries;
-- +goose StatementEnd
