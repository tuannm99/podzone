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

INSERT INTO routed_order_activities (
	order_id,
	product_title,
	operator_assignee,
	activity_type,
	actor,
	message,
	details_json,
	created_at
)
SELECT
	ro.id,
	ro.product_title,
	ro.operator_assignee,
	COALESCE(activity.item->>'type', 'system'),
	COALESCE(NULLIF(activity.item->>'actor', ''), 'system'),
	COALESCE(activity.item->>'message', ''),
	COALESCE(activity.item->'details', '[]'::jsonb)::text,
	COALESCE(
		NULLIF(activity.item->>'createdAt', '')::timestamptz,
		ro.updated_at,
		ro.created_at,
		NOW()
	)
FROM routed_orders ro
CROSS JOIN LATERAL jsonb_array_elements(COALESCE(ro.activity_log_json, '[]')::jsonb) AS activity(item)
WHERE NOT EXISTS (
	SELECT 1
	FROM routed_order_activities existing
	WHERE existing.order_id = ro.id
);
