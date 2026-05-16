ALTER TABLE routed_order_activities
	ADD COLUMN IF NOT EXISTS partner TEXT NOT NULL DEFAULT '';

UPDATE routed_order_activities activities
SET partner = routed_orders.partner
FROM routed_orders
WHERE activities.order_id = routed_orders.id
  AND activities.partner = '';

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_partner_created_at
	ON routed_order_activities (partner, created_at DESC, id DESC);
