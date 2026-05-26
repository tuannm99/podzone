-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_platform_user_permission_boundaries (
  user_id BIGINT PRIMARY KEY,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam_tenant_user_permission_boundaries (
  tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_iam_platform_user_permission_boundaries_policy_id
ON iam_platform_user_permission_boundaries(policy_id);

CREATE INDEX IF NOT EXISTS idx_iam_tenant_user_permission_boundaries_policy_id
ON iam_tenant_user_permission_boundaries(policy_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_tenant_user_permission_boundaries_policy_id;
DROP INDEX IF EXISTS idx_iam_platform_user_permission_boundaries_policy_id;
DROP TABLE IF EXISTS iam_tenant_user_permission_boundaries;
DROP TABLE IF EXISTS iam_platform_user_permission_boundaries;
-- +goose StatementEnd
