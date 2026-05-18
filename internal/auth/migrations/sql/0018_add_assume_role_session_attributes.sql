-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_session_name TEXT NOT NULL DEFAULT '';

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_source_identity TEXT NOT NULL DEFAULT '';

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_expires_at TIMESTAMPTZ NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_expires_at;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_source_identity;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_session_name;
-- +goose StatementEnd
