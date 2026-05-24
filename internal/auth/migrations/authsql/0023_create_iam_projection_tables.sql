-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_tenants_projection (
  tenant_id TEXT PRIMARY KEY,
  slug TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_tenant_memberships_projection (
  tenant_id TEXT NOT NULL,
  user_id BIGINT NOT NULL,
  role_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_iam_tenant_memberships_projection_user_id
  ON iam_tenant_memberships_projection(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_tenant_memberships_projection_user_id;
DROP TABLE IF EXISTS iam_tenant_memberships_projection;
DROP TABLE IF EXISTS iam_tenants_projection;
-- +goose StatementEnd
