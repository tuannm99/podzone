-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS session_tags_json TEXT NOT NULL DEFAULT '{}';

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_service_principal TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_service_principal;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS session_tags_json;
-- +goose StatementEnd
