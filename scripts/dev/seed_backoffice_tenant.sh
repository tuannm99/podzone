#!/usr/bin/env sh
set -eu

TENANT_ID="${1:-${TENANT_ID:-tenant-dev}}"
STORE_NAME="${2:-${STORE_NAME:-Demo Store}}"
STORE_SUBDOMAIN="${3:-${STORE_SUBDOMAIN:-demo-store}}"
ONBOARDING_URL="${4:-${ONBOARDING_URL:-http://localhost:8800}}"
WAIT_SECONDS="${WAIT_SECONDS:-90}"
CREATE_STORE="${CREATE_STORE:-true}"
ONBOARDING_SERVICE_TOKEN="${ONBOARDING_SERVICE_TOKEN:-dev-bootstrap-token}"

echo "Waiting for onboarding..."
until curl -fsS "${ONBOARDING_URL}" >/dev/null 2>&1; do
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
        \"subdomain\":\"${STORE_SUBDOMAIN}\"
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
  curl -fsS \
    "${ONBOARDING_URL}/onboarding/v1/requests?collection.page=1&collection.pageSize=100" \
    -H "X-Onboarding-Service-Token: ${ONBOARDING_SERVICE_TOKEN}" \
    -H "X-Tenant-ID: ${TENANT_ID}" \
    -H "X-User-ID: dev-bootstrap" \
    >/tmp/podzone-onboarding-requests.json

  if grep -q '"status":"ready"' /tmp/podzone-onboarding-requests.json; then
    echo "Placement route ready for ${TENANT_ID}."
    exit 0
  fi
  if grep -q '"status":"failed"' /tmp/podzone-onboarding-requests.json; then
    cat /tmp/podzone-onboarding-requests.json
    echo "Onboarding provisioning failed for ${TENANT_ID}." >&2
    exit 1
  fi
  i=$((i + 1))
  sleep 1
done

cat /tmp/podzone-onboarding-requests.json
echo "Timed out waiting for onboarding provisioning for ${TENANT_ID}." >&2
exit 1
