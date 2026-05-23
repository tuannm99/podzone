-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_groups (
  id BIGSERIAL PRIMARY KEY,
  scope TEXT NOT NULL,
  tenant_id TEXT REFERENCES tenants(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  is_system BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (scope, tenant_id, name)
);

CREATE TABLE IF NOT EXISTS iam_group_members (
  group_id BIGINT NOT NULL REFERENCES iam_groups(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS iam_group_policy_attachments (
  group_id BIGINT NOT NULL REFERENCES iam_groups(id) ON DELETE CASCADE,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, policy_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_group_policy_attachments;
DROP TABLE IF EXISTS iam_group_members;
DROP TABLE IF EXISTS iam_groups;
-- +goose StatementEnd
