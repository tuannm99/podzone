-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_policy_versions (
  id BIGSERIAL PRIMARY KEY,
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  version TEXT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (policy_id, version)
);

CREATE TABLE IF NOT EXISTS iam_policy_version_statements (
  policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
  version TEXT NOT NULL,
  statement_index INT NOT NULL,
  effect TEXT NOT NULL,
  action_pattern TEXT NOT NULL,
  resource_pattern TEXT NOT NULL DEFAULT '*',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (policy_id, version, statement_index)
);

ALTER TABLE iam_policies
ADD COLUMN IF NOT EXISTS default_version TEXT NOT NULL DEFAULT 'v1';

INSERT INTO iam_policy_versions (policy_id, version, is_default, created_at)
SELECT p.id, 'v1', TRUE, p.created_at
FROM iam_policies p
WHERE NOT EXISTS (
  SELECT 1 FROM iam_policy_versions pv WHERE pv.policy_id = p.id AND pv.version = 'v1'
);

INSERT INTO iam_policy_version_statements (policy_id, version, statement_index, effect, action_pattern, resource_pattern, created_at)
SELECT ps.policy_id, 'v1', ROW_NUMBER() OVER (PARTITION BY ps.policy_id ORDER BY ps.id) - 1, ps.effect, ps.action_pattern, ps.resource_pattern, ps.created_at
FROM iam_policy_statements ps
WHERE NOT EXISTS (
  SELECT 1
  FROM iam_policy_version_statements pvs
  WHERE pvs.policy_id = ps.policy_id
    AND pvs.version = 'v1'
);

UPDATE iam_policies
SET default_version = 'v1'
WHERE default_version = '' OR default_version IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE iam_policies
DROP COLUMN IF EXISTS default_version;
DROP TABLE IF EXISTS iam_policy_version_statements;
DROP TABLE IF EXISTS iam_policy_versions;
-- +goose StatementEnd
