-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS iam_audit_logs (
  id TEXT PRIMARY KEY,
  actor_user_id BIGINT NOT NULL,
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL DEFAULT '',
  tenant_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_iam_audit_logs_actor_created_at
  ON iam_audit_logs(actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_iam_audit_logs_tenant_created_at
  ON iam_audit_logs(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_iam_audit_logs_resource_created_at
  ON iam_audit_logs(resource_type, resource_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_audit_logs;
-- +goose StatementEnd
