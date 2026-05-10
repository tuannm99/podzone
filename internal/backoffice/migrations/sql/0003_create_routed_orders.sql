CREATE TABLE IF NOT EXISTS routed_orders (
	id TEXT PRIMARY KEY,
	candidate_id TEXT NOT NULL,
	product_title TEXT NOT NULL,
	partner TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	total TEXT NOT NULL,
	customer_name TEXT NOT NULL,
	status TEXT NOT NULL,
	timeline_json TEXT NOT NULL,
	exception_type TEXT NOT NULL DEFAULT '',
	exception_status TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);
