ALTER TABLE routed_orders
	ADD COLUMN IF NOT EXISTS activity_log_json TEXT NOT NULL DEFAULT '[]';
