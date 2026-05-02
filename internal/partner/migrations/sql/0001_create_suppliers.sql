-- +goose Up
CREATE TABLE IF NOT EXISTS partners (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  contact_name TEXT NOT NULL DEFAULT '',
  contact_email TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  partner_type TEXT NOT NULL DEFAULT 'print_on_demand',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT uq_partners_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_partners_tenant_created_at
  ON partners(tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_partners_tenant_status
  ON partners(tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_partners_tenant_partner_type
  ON partners(tenant_id, partner_type);

-- +goose Down
DROP TABLE IF EXISTS partners;
