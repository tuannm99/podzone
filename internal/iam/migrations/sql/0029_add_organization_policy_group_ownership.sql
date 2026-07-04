-- +goose Up
-- +goose StatementBegin
ALTER TABLE iam_policies
ADD COLUMN IF NOT EXISTS org_id TEXT REFERENCES iam_organizations(id) ON DELETE CASCADE;

ALTER TABLE iam_policies
DROP CONSTRAINT IF EXISTS iam_policies_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_policies_owner_name
ON iam_policies(scope, COALESCE(org_id, ''), name);

ALTER TABLE iam_policies
ADD CONSTRAINT chk_iam_policies_owner
CHECK (
  (scope = 'organization' AND org_id IS NOT NULL)
  OR (scope <> 'organization' AND org_id IS NULL)
);

ALTER TABLE iam_groups
ADD COLUMN IF NOT EXISTS org_id TEXT REFERENCES iam_organizations(id) ON DELETE CASCADE;

ALTER TABLE iam_groups
DROP CONSTRAINT IF EXISTS iam_groups_scope_tenant_id_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_groups_platform_name
ON iam_groups(name)
WHERE scope = 'platform';

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_groups_organization_name
ON iam_groups(org_id, name)
WHERE scope = 'organization';

CREATE UNIQUE INDEX IF NOT EXISTS idx_iam_groups_tenant_name
ON iam_groups(tenant_id, name)
WHERE scope = 'tenant';

ALTER TABLE iam_groups
ADD CONSTRAINT chk_iam_groups_owner
CHECK (
  (scope = 'platform' AND tenant_id IS NULL AND org_id IS NULL)
  OR (scope = 'organization' AND tenant_id IS NULL AND org_id IS NOT NULL)
  OR (scope = 'tenant' AND tenant_id IS NOT NULL AND org_id IS NULL)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM iam_groups WHERE scope = 'organization';
DELETE FROM iam_policies WHERE scope = 'organization';

ALTER TABLE iam_groups DROP CONSTRAINT IF EXISTS chk_iam_groups_owner;
DROP INDEX IF EXISTS idx_iam_groups_tenant_name;
DROP INDEX IF EXISTS idx_iam_groups_organization_name;
DROP INDEX IF EXISTS idx_iam_groups_platform_name;
ALTER TABLE iam_groups
ADD CONSTRAINT iam_groups_scope_tenant_id_name_key UNIQUE (scope, tenant_id, name);
ALTER TABLE iam_groups DROP COLUMN IF EXISTS org_id;

ALTER TABLE iam_policies DROP CONSTRAINT IF EXISTS chk_iam_policies_owner;
DROP INDEX IF EXISTS idx_iam_policies_owner_name;
ALTER TABLE iam_policies ADD CONSTRAINT iam_policies_name_key UNIQUE (name);
ALTER TABLE iam_policies DROP COLUMN IF EXISTS org_id;
-- +goose StatementEnd
