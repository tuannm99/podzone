-- +goose Up
ALTER TABLE IF EXISTS suppliers RENAME TO partners;

ALTER INDEX IF EXISTS uq_suppliers_tenant_code
  RENAME TO uq_partners_tenant_code;

ALTER INDEX IF EXISTS idx_suppliers_tenant_created_at
  RENAME TO idx_partners_tenant_created_at;

ALTER INDEX IF EXISTS idx_suppliers_tenant_status
  RENAME TO idx_partners_tenant_status;

ALTER INDEX IF EXISTS idx_suppliers_tenant_partner_type
  RENAME TO idx_partners_tenant_partner_type;

-- +goose Down
ALTER TABLE IF EXISTS partners RENAME TO suppliers;

ALTER INDEX IF EXISTS uq_partners_tenant_code
  RENAME TO uq_suppliers_tenant_code;

ALTER INDEX IF EXISTS idx_partners_tenant_created_at
  RENAME TO idx_suppliers_tenant_created_at;

ALTER INDEX IF EXISTS idx_partners_tenant_status
  RENAME TO idx_suppliers_tenant_status;

ALTER INDEX IF EXISTS idx_partners_tenant_partner_type
  RENAME TO idx_suppliers_tenant_partner_type;
