CREATE TABLE IF NOT EXISTS routed_order_activities (
	id BIGSERIAL PRIMARY KEY,
	order_id TEXT NOT NULL,
	product_title TEXT NOT NULL,
	operator_assignee TEXT NOT NULL DEFAULT 'unassigned',
	activity_type TEXT NOT NULL,
	actor TEXT NOT NULL DEFAULT 'system',
	message TEXT NOT NULL,
	details_json TEXT NOT NULL DEFAULT '[]',
	created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_created_at
	ON routed_order_activities (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_order_id
	ON routed_order_activities (order_id);
