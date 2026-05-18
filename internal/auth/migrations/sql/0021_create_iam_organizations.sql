-- +goose Up
-- +goose StatementBegin
ALTER TABLE tenants
ADD COLUMN IF NOT EXISTS org_id TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS iam_organizations (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS iam_org_service_control_policies (
    org_id TEXT NOT NULL REFERENCES iam_organizations(id) ON DELETE CASCADE,
    policy_id BIGINT NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, policy_id)
);

CREATE INDEX IF NOT EXISTS idx_tenants_org_id ON tenants(org_id);
CREATE INDEX IF NOT EXISTS idx_iam_org_scp_policy_id ON iam_org_service_control_policies(policy_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_iam_org_scp_policy_id;
DROP INDEX IF EXISTS idx_tenants_org_id;
DROP TABLE IF EXISTS iam_org_service_control_policies;
DROP TABLE IF EXISTS iam_organizations;
ALTER TABLE tenants DROP COLUMN IF EXISTS org_id;
-- +goose StatementEnd
