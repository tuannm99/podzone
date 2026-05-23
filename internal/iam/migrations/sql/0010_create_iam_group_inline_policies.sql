-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_group_inline_policies (
  group_id BIGINT NOT NULL REFERENCES iam_groups(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, name)
);

CREATE TABLE IF NOT EXISTS iam_group_inline_policy_statements (
  group_id BIGINT NOT NULL,
  policy_name TEXT NOT NULL,
  statement_index INT NOT NULL,
  effect TEXT NOT NULL,
  action_pattern TEXT NOT NULL,
  resource_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, policy_name, statement_index),
  FOREIGN KEY (group_id, policy_name)
    REFERENCES iam_group_inline_policies(group_id, name)
    ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_group_inline_policy_statements;
DROP TABLE IF EXISTS iam_group_inline_policies;
-- +goose StatementEnd
