CREATE TABLE IF NOT EXISTS product_setup_drafts (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	partner TEXT NOT NULL,
	base_cost TEXT NOT NULL,
	retail_price TEXT NOT NULL,
	status TEXT NOT NULL,
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS product_setup_candidates (
	id TEXT PRIMARY KEY,
	draft_id TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	sku TEXT NOT NULL,
	partner TEXT NOT NULL,
	base_cost TEXT NOT NULL,
	retail_price TEXT NOT NULL,
	estimated_margin TEXT NOT NULL,
	status TEXT NOT NULL,
	channel TEXT NOT NULL,
	variants_json TEXT NOT NULL,
	artwork_checklist_json TEXT NOT NULL,
	merchandising_notes TEXT NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ NOT NULL
);
