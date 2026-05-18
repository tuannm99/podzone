-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_id BIGINT NOT NULL DEFAULT 0;

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_scope TEXT NOT NULL DEFAULT '';

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_name TEXT NOT NULL DEFAULT '';

ALTER TABLE auth_sessions
ADD COLUMN IF NOT EXISTS assumed_role_tenant_id TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_tenant_id;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_name;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_scope;

ALTER TABLE auth_sessions
DROP COLUMN IF EXISTS assumed_role_id;
-- +goose StatementEnd
