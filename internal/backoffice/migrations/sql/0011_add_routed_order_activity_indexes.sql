CREATE INDEX IF NOT EXISTS idx_routed_order_activities_type_created_at
	ON routed_order_activities (activity_type, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_order_created_at
	ON routed_order_activities (order_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_actor_created_at
	ON routed_order_activities (actor, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_routed_order_activities_non_system_created_at
	ON routed_order_activities (created_at DESC, id DESC)
	WHERE activity_type <> 'system';
