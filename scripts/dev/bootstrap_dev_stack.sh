#!/usr/bin/env sh
set -eu

TENANT_ID="${1:-${TENANT_ID:-tenant-dev}}"
STORE_NAME="${2:-${STORE_NAME:-Demo Store}}"
STORE_SUBDOMAIN="${3:-${STORE_SUBDOMAIN:-demo-store}}"
DEV_USERNAME="${4:-${DEV_USERNAME:-devowner}}"
DEV_EMAIL="${5:-${DEV_EMAIL:-${DEV_USERNAME}@podzone.dev}}"
DEV_PASSWORD="${6:-${DEV_PASSWORD:-DevPass123!}}"
TENANT_NAME="${TENANT_NAME:-Demo POD Tenant}"
TENANT_SLUG="${TENANT_SLUG:-$TENANT_ID}"
DEV_FULL_NAME="${DEV_FULL_NAME:-Dev Owner}"
CLUSTER_NAME="${CLUSTER_NAME:-pg-default}"
DB_NAME="${DB_NAME:-podzone_tenants}"
SCHEMA_NAME="${SCHEMA_NAME:-$(printf '%s' "t_${TENANT_ID}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/_/g; s/^_*//; s/_*$//')}"
PG_HOST="${PG_HOST:-pgbouncer}"
PG_PORT="${PG_PORT:-6432}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_SSL_MODE="${PG_SSL_MODE:-disable}"
ONBOARDING_URL="${ONBOARDING_URL:-http://onboarding-service:8800}"
JWT_SECRET="${JWT_SECRET:-dev-secret}"
JWT_KEY="${JWT_KEY:-}"
AUTH_BOOTSTRAP_OUTPUT="${AUTH_BOOTSTRAP_OUTPUT:-/workspace/frontend/public/dev-auth-bootstrap.json}"
UI_AUTH_BOOTSTRAP_TARGET="${UI_AUTH_BOOTSTRAP_TARGET:-/workspace/frontend/public/dev-auth-bootstrap.json}"
DEV_OWNER_ID_OUTPUT="${DEV_OWNER_ID_OUTPUT:-/tmp/podzone-dev-owner-id}"
create_store="true"

TENANT_ID="${TENANT_ID}" \
TENANT_NAME="${TENANT_NAME}" \
TENANT_SLUG="${TENANT_SLUG}" \
DEV_USERNAME="${DEV_USERNAME}" \
DEV_EMAIL="${DEV_EMAIL}" \
DEV_PASSWORD="${DEV_PASSWORD}" \
DEV_FULL_NAME="${DEV_FULL_NAME}" \
PG_HOST="postgres" \
PG_PORT="5432" \
PG_USER="${PG_USER}" \
PG_PASSWORD="${PG_PASSWORD}" \
PG_SSL_MODE="${PG_SSL_MODE}" \
JWT_SECRET="${JWT_SECRET}" \
JWT_KEY="${JWT_KEY}" \
AUTH_BOOTSTRAP_OUTPUT="${AUTH_BOOTSTRAP_OUTPUT}" \
DEV_OWNER_ID_OUTPUT="${DEV_OWNER_ID_OUTPUT}" \
go run /workspace/scripts/dev/seed_auth_bootstrap.go

store_owner_id="$(cat "${DEV_OWNER_ID_OUTPUT}")"

TENANT_ID="${TENANT_ID}" \
DB_NAME="${DB_NAME}" \
SCHEMA_NAME="${SCHEMA_NAME}" \
PG_HOST="${PG_HOST}" \
PG_PORT="${PG_PORT}" \
PG_USER="${PG_USER}" \
PG_PASSWORD="${PG_PASSWORD}" \
PG_SSL_MODE="${PG_SSL_MODE}" \
CREATE_STORE="${create_store}" \
STORE_OWNER_ID="${store_owner_id}" \
sh /workspace/scripts/dev/seed_backoffice_tenant.sh \
  "${TENANT_ID}" \
  "${STORE_NAME}" \
  "${STORE_SUBDOMAIN}" \
  "${ONBOARDING_URL}"

TENANT_ID="${TENANT_ID}" \
STORE_NAME="${STORE_NAME}" \
STORE_SUBDOMAIN="${STORE_SUBDOMAIN}" \
DB_NAME="${DB_NAME}" \
SCHEMA_NAME="${SCHEMA_NAME}" \
PG_HOST="${PG_HOST}" \
PG_PORT="${PG_PORT}" \
PG_USER="${PG_USER}" \
PG_PASSWORD="${PG_PASSWORD}" \
PG_SSL_MODE="${PG_SSL_MODE}" \
ONBOARDING_URL="${ONBOARDING_URL}" \
ONBOARDING_SERVICE_TOKEN="${ONBOARDING_SERVICE_TOKEN:-dev-bootstrap-token}" \
go run /workspace/scripts/dev/seed_backoffice_sample.go

AUTH_BOOTSTRAP_OUTPUT="${AUTH_BOOTSTRAP_OUTPUT}" \
UI_AUTH_BOOTSTRAP_TARGET="${UI_AUTH_BOOTSTRAP_TARGET}" \
sh /workspace/scripts/dev/sync_ui_auth_bootstrap.sh

echo "Dev stack bootstrap completed."
