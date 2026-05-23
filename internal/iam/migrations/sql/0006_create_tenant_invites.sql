-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tenant_invites (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  role_id BIGINT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'pending',
  invited_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  accepted_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
  token_hash TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ NULL,
  revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant_created_at
  ON tenant_invites(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_email_status
  ON tenant_invites(email, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tenant_invites;
-- +goose StatementEnd
