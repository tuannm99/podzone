-- +goose Up
-- +goose StatementBegin
ALTER TABLE partners
ADD COLUMN IF NOT EXISTS supported_product_types TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
ADD COLUMN IF NOT EXISTS supported_regions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
ADD COLUMN IF NOT EXISTS sla_days INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS routing_priority INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE partners
DROP COLUMN IF EXISTS routing_priority,
DROP COLUMN IF EXISTS sla_days,
DROP COLUMN IF EXISTS supported_regions,
DROP COLUMN IF EXISTS supported_product_types;
-- +goose StatementEnd
