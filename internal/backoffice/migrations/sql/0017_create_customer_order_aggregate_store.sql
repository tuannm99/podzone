CREATE TABLE IF NOT EXISTS customer_orders (
	id TEXT PRIMARY KEY,
	aggregate_version BIGINT NOT NULL DEFAULT 0,
	store_id TEXT NOT NULL,
	candidate_id TEXT NOT NULL,
	product_title TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	total TEXT NOT NULL,
	customer_name TEXT NOT NULL,
	status TEXT NOT NULL,
	partner TEXT NOT NULL DEFAULT '',
	operator_assignee TEXT NOT NULL DEFAULT 'unassigned',
	shipment_sla_due_at TIMESTAMPTZ,
	issue_sla_due_at TIMESTAMPTZ,
	exception_status TEXT NOT NULL DEFAULT '',
	routing_block_code TEXT NOT NULL DEFAULT '',
	routing_block_reason TEXT NOT NULL DEFAULT '',
	settlement_status TEXT NOT NULL DEFAULT 'pending',
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

INSERT INTO customer_orders (
	id,
	store_id,
	candidate_id,
	product_title,
	quantity,
	total,
	customer_name,
	status,
	partner,
	operator_assignee,
	shipment_sla_due_at,
	issue_sla_due_at,
	exception_status,
	routing_block_code,
	routing_block_reason,
	settlement_status,
	created_at,
	updated_at
)
SELECT
	id,
	store_id,
	candidate_id,
	product_title,
	quantity,
	total,
	customer_name,
	status,
	partner,
	operator_assignee,
	shipment_sla_due_at,
	issue_sla_due_at,
	exception_status,
	routing_block_code,
	routing_block_reason,
	settlement_status,
	created_at,
	updated_at
FROM routed_orders
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_customer_orders_store_id
	ON customer_orders (store_id, created_at DESC);
