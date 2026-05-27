ALTER TABLE routed_orders
	ADD COLUMN IF NOT EXISTS store_id TEXT NOT NULL DEFAULT '';

ALTER TABLE routed_order_activities
	ADD COLUMN IF NOT EXISTS store_id TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = current_schema()
			AND table_name = 'stores'
	) THEN
		UPDATE routed_orders
		SET store_id = (
			SELECT stores.id
			FROM stores
			ORDER BY stores.created_at ASC
			LIMIT 1
		)
		WHERE store_id = ''
			AND EXISTS (SELECT 1 FROM stores);
	END IF;
END $$;

UPDATE routed_order_activities activities
SET store_id = orders.store_id
FROM routed_orders orders
WHERE activities.order_id = orders.id
  AND activities.store_id = '';

CREATE INDEX IF NOT EXISTS idx_routed_orders_store_id
	ON routed_orders (store_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_store_id
	ON routed_order_activities (store_id, created_at DESC, id DESC);
