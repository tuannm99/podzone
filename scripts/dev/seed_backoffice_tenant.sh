#!/usr/bin/env sh
set -eu

CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
ONBOARDING_URL="${ONBOARDING_URL:-http://localhost:8800}"
TENANT_ID="${TENANT_ID:-tenant-dev}"
STORE_NAME="${STORE_NAME:-Demo Store}"
STORE_SUBDOMAIN="${STORE_SUBDOMAIN:-demo-store}"
CLUSTER_NAME="${CLUSTER_NAME:-pg-default}"
DB_NAME="${DB_NAME:-postgres}"
SCHEMA_NAME="${SCHEMA_NAME:-t_${TENANT_ID}}"
PG_HOST="${PG_HOST:-pgbouncer}"
PG_PORT="${PG_PORT:-6432}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_SSL_MODE="${PG_SSL_MODE:-disable}"
WAIT_SECONDS="${WAIT_SECONDS:-15}"
CREATE_STORE="${CREATE_STORE:-true}"

echo "Seeding postgres cluster config into Consul for ${CLUSTER_NAME}..."
curl -fsS -X PUT \
  "${CONSUL_URL}/v1/kv/podzone/postgres/clusters/${CLUSTER_NAME}" \
  --data "{\"host\":\"${PG_HOST}\",\"port\":${PG_PORT},\"user\":\"${PG_USER}\",\"password\":\"${PG_PASSWORD}\",\"ssl_mode\":\"${PG_SSL_MODE}\"}" >/dev/null

echo "Publishing tenant placement via onboarding..."
curl -fsS -X POST \
  "${ONBOARDING_URL}/onboarding/v1/infras/connections" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "X-User: dev-seed" \
  --data "{
    \"infra_type\": \"postgres\",
    \"name\": \"default\",
    \"endpoint\": \"postgres://${PG_USER}:***@${PG_HOST}:${PG_PORT}/${DB_NAME}\",
    \"status\": \"active\",
    \"meta\": {
      \"purpose\": \"backoffice\",
      \"cluster\": \"${CLUSTER_NAME}\"
    },
    \"config\": {
      \"driver\": \"postgres\",
      \"pool\": \"pgbouncer\"
    },
    \"cluster_name\": \"${CLUSTER_NAME}\",
    \"mode\": \"schema\",
    \"db_name\": \"${DB_NAME}\",
    \"schema_name\": \"${SCHEMA_NAME}\"
  }" >/dev/null

echo "Waiting for placement to appear in Consul..."
i=0
while [ "$i" -lt "$WAIT_SECONDS" ]; do
  if curl -fsS "${CONSUL_URL}/v1/kv/podzone/tenants/${TENANT_ID}/placement?raw" >/tmp/podzone-placement.json 2>/dev/null; then
    echo "Placement ready for ${TENANT_ID}:"
    cat /tmp/podzone-placement.json
    echo
    break
  fi
  i=$((i + 1))
  sleep 1
done

if [ "$i" -ge "$WAIT_SECONDS" ]; then
  echo "Timed out waiting for placement key podzone/tenants/${TENANT_ID}/placement" >&2
  exit 1
fi

if [ "${CREATE_STORE}" = "true" ]; then
  echo "Creating onboarding store record..."
  store_status="$(
    curl -sS -o /tmp/podzone-onboarding-store.json -w "%{http_code}" \
      -X POST \
      "${ONBOARDING_URL}/onboarding/v1/stores" \
      -H "Content-Type: application/json" \
      --data "{\"name\":\"${STORE_NAME}\",\"subdomain\":\"${STORE_SUBDOMAIN}\"}"
  )"
  cat /tmp/podzone-onboarding-store.json
  echo
  if [ "${store_status}" != "201" ] && [ "${store_status}" != "409" ]; then
    echo "Failed to create onboarding store, status=${store_status}" >&2
    exit 1
  fi
fi

echo "Backoffice tenant seed completed."
