-- +goose Up
-- +goose StatementBegin
ALTER TABLE iam_role_trust_statements
ADD COLUMN IF NOT EXISTS external_id_pattern TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE iam_role_trust_statements
DROP COLUMN IF EXISTS external_id_pattern;
-- +goose StatementEnd
