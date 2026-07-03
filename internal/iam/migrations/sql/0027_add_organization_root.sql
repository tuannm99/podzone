-- +goose Up
-- +goose StatementBegin
ALTER TABLE iam_organizations
ADD COLUMN IF NOT EXISTS root_user_id BIGINT NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_organizations_root_user_id
ON iam_organizations(root_user_id)
WHERE root_user_id > 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_organizations_root_user_id;
ALTER TABLE iam_organizations DROP COLUMN IF EXISTS root_user_id;
-- +goose StatementEnd
