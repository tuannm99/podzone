-- +goose Up
ALTER TABLE partners
  ADD COLUMN IF NOT EXISTS partner_type TEXT NOT NULL DEFAULT 'print_on_demand';

UPDATE partners
SET partner_type = 'print_on_demand'
WHERE partner_type = '';

CREATE INDEX IF NOT EXISTS idx_partners_tenant_partner_type
  ON partners(tenant_id, partner_type);

-- +goose Down
DROP INDEX IF EXISTS idx_partners_tenant_partner_type;

ALTER TABLE partners
  DROP COLUMN IF EXISTS partner_type;
