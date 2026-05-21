-- +goose Up
-- +goose StatementBegin
ALTER TABLE partners
  ADD COLUMN IF NOT EXISTS base_fulfillment_cost TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS shipping_cost_rules_json JSONB NOT NULL DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE partners
  DROP COLUMN IF EXISTS shipping_cost_rules_json,
  DROP COLUMN IF EXISTS base_fulfillment_cost;
-- +goose StatementEnd
