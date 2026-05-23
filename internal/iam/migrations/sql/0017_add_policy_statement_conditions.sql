-- +goose Up
-- +goose StatementBegin
ALTER TABLE iam_policy_statements
ADD COLUMN IF NOT EXISTS conditions_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE iam_policy_version_statements
ADD COLUMN IF NOT EXISTS conditions_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE iam_group_inline_policy_statements
ADD COLUMN IF NOT EXISTS conditions_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE iam_platform_user_inline_policy_statements
ADD COLUMN IF NOT EXISTS conditions_json TEXT NOT NULL DEFAULT '[]';

ALTER TABLE iam_tenant_user_inline_policy_statements
ADD COLUMN IF NOT EXISTS conditions_json TEXT NOT NULL DEFAULT '[]';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE iam_tenant_user_inline_policy_statements
DROP COLUMN IF EXISTS conditions_json;

ALTER TABLE iam_platform_user_inline_policy_statements
DROP COLUMN IF EXISTS conditions_json;

ALTER TABLE iam_group_inline_policy_statements
DROP COLUMN IF EXISTS conditions_json;

ALTER TABLE iam_policy_version_statements
DROP COLUMN IF EXISTS conditions_json;

ALTER TABLE iam_policy_statements
DROP COLUMN IF EXISTS conditions_json;
-- +goose StatementEnd
