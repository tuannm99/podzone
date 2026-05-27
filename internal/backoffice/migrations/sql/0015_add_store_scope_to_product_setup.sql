DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = current_schema()
			AND table_name = 'product_setup_drafts'
	) THEN
		ALTER TABLE product_setup_drafts
			ADD COLUMN IF NOT EXISTS store_id TEXT NOT NULL DEFAULT '';

		IF EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = current_schema()
				AND table_name = 'stores'
		) THEN
			UPDATE product_setup_drafts
			SET store_id = (
				SELECT stores.id
				FROM stores
				ORDER BY stores.created_at ASC
				LIMIT 1
			)
			WHERE store_id = ''
				AND EXISTS (SELECT 1 FROM stores);
		END IF;

		CREATE INDEX IF NOT EXISTS idx_product_setup_drafts_store_id
			ON product_setup_drafts (store_id, created_at DESC);
	END IF;
END $$;

DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = current_schema()
			AND table_name = 'product_setup_candidates'
	) THEN
		ALTER TABLE product_setup_candidates
			ADD COLUMN IF NOT EXISTS store_id TEXT NOT NULL DEFAULT '';

		UPDATE product_setup_candidates candidates
		SET store_id = drafts.store_id
		FROM product_setup_drafts drafts
		WHERE candidates.draft_id = drafts.id
			AND candidates.store_id = ''
			AND drafts.store_id <> '';

		IF EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = current_schema()
				AND table_name = 'stores'
		) THEN
			UPDATE product_setup_candidates
			SET store_id = (
				SELECT stores.id
				FROM stores
				ORDER BY stores.created_at ASC
				LIMIT 1
			)
			WHERE store_id = ''
				AND EXISTS (SELECT 1 FROM stores);
		END IF;

		CREATE INDEX IF NOT EXISTS idx_product_setup_candidates_store_id
			ON product_setup_candidates (store_id, updated_at DESC);
	END IF;
END $$;
