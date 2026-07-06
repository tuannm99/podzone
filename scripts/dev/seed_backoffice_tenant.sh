#!/usr/bin/env sh
set -eu

TENANT_ID="${1:-${TENANT_ID:-tenant-dev}}"
STORE_NAME="${2:-${STORE_NAME:-Demo Store}}"
STORE_SUBDOMAIN="${3:-${STORE_SUBDOMAIN:-demo-store}}"
ONBOARDING_URL="${4:-${ONBOARDING_URL:-http://localhost:8800}}"
WAIT_SECONDS="${WAIT_SECONDS:-90}"
CREATE_STORE="${CREATE_STORE:-true}"
ONBOARDING_SERVICE_TOKEN="${ONBOARDING_SERVICE_TOKEN:-dev-bootstrap-token}"
STORE_OWNER_ID="${STORE_OWNER_ID:-}"

if [ "${CREATE_STORE}" = "true" ] && [ -z "${STORE_OWNER_ID}" ]; then
  echo "STORE_OWNER_ID is required when CREATE_STORE=true." >&2
  exit 1
fi

echo "Waiting for onboarding..."
i=0
until curl -fsS "${ONBOARDING_URL}/healthz" >/dev/null 2>&1; do
  i=$((i + 1))
  if [ "$i" -ge "$WAIT_SECONDS" ]; then
    echo "Timed out waiting for onboarding at ${ONBOARDING_URL}/healthz." >&2
    exit 1
  fi
  sleep 1
done

if [ "${CREATE_STORE}" = "true" ]; then
  echo "Creating onboarding store request..."
  store_status="$(
    curl -sS -o /tmp/podzone-onboarding-store.json -w "%{http_code}" \
      -X POST \
      "${ONBOARDING_URL}/onboarding/v1/stores" \
      -H "Content-Type: application/json" \
      -H "X-Onboarding-Service-Token: ${ONBOARDING_SERVICE_TOKEN}" \
      -H "X-Tenant-ID: ${TENANT_ID}" \
      -H "X-User-ID: dev-bootstrap" \
      --data "{
        \"name\":\"${STORE_NAME}\",
        \"subdomain\":\"${STORE_SUBDOMAIN}\",
        \"owner_id\":\"${STORE_OWNER_ID}\"
      }"
  )"
  cat /tmp/podzone-onboarding-store.json
  echo
  if [ "${store_status}" != "201" ] && [ "${store_status}" != "409" ]; then
    echo "Failed to create onboarding store request, status=${store_status}" >&2
    exit 1
  fi
fi

echo "Waiting for onboarding provisioning and Mongo KV publication..."
i=0
while [ "$i" -lt "$WAIT_SECONDS" ]; do
  if curl -fsS \
    "${ONBOARDING_URL}/onboarding/v1/requests?collection.page=1&collection.pageSize=100" \
    -H "X-Onboarding-Service-Token: ${ONBOARDING_SERVICE_TOKEN}" \
    -H "X-Tenant-ID: ${TENANT_ID}" \
    -H "X-User-ID: dev-bootstrap" \
    >/tmp/podzone-onboarding-requests.json; then
    if grep -q '"status":"ready"' /tmp/podzone-onboarding-requests.json; then
      echo "Placement route ready for ${TENANT_ID}."
      exit 0
    fi
    if grep -q '"status":"failed"' /tmp/podzone-onboarding-requests.json; then
      cat /tmp/podzone-onboarding-requests.json
      echo "Onboarding provisioning failed for ${TENANT_ID}." >&2
      exit 1
    fi
  fi
  i=$((i + 1))
  sleep 1
done

if [ -f /tmp/podzone-onboarding-requests.json ]; then
  cat /tmp/podzone-onboarding-requests.json
fi
echo "Timed out waiting for onboarding provisioning for ${TENANT_ID}." >&2
exit 1
