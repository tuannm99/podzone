-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_platform_user_inline_policies (
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, name)
);

CREATE TABLE IF NOT EXISTS iam_platform_user_inline_policy_statements (
  user_id BIGINT NOT NULL,
  policy_name TEXT NOT NULL,
  statement_index INT NOT NULL,
  effect TEXT NOT NULL,
  action_pattern TEXT NOT NULL,
  resource_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, policy_name, statement_index),
  FOREIGN KEY (user_id, policy_name)
    REFERENCES iam_platform_user_inline_policies(user_id, name)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS iam_tenant_user_inline_policies (
  tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id, name)
);

CREATE TABLE IF NOT EXISTS iam_tenant_user_inline_policy_statements (
  tenant_id TEXT NOT NULL,
  user_id BIGINT NOT NULL,
  policy_name TEXT NOT NULL,
  statement_index INT NOT NULL,
  effect TEXT NOT NULL,
  action_pattern TEXT NOT NULL,
  resource_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, user_id, policy_name, statement_index),
  FOREIGN KEY (tenant_id, user_id, policy_name)
    REFERENCES iam_tenant_user_inline_policies(tenant_id, user_id, name)
    ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_tenant_user_inline_policy_statements;
DROP TABLE IF EXISTS iam_tenant_user_inline_policies;
DROP TABLE IF EXISTS iam_platform_user_inline_policy_statements;
DROP TABLE IF EXISTS iam_platform_user_inline_policies;
-- +goose StatementEnd
