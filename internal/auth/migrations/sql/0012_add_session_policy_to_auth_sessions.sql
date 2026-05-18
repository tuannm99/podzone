-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS session_policy_json TEXT NOT NULL DEFAULT '[]';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS session_policy_json;
-- +goose StatementEnd
