#!/usr/bin/env sh
set -eu

TENANT_ID="${1:-tenant-dev}"
STORE_NAME="${2:-Demo Store}"
STORE_SUBDOMAIN="${3:-demo-store}"
DEV_USERNAME="${4:-devowner}"
DEV_EMAIL="${5:-${DEV_USERNAME}@podzone.dev}"
DEV_PASSWORD="${6:-DevPass123!}"

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"

echo "Starting local POD dev stack..."
echo "  tenant: ${TENANT_ID}"
echo "  store: ${STORE_NAME} (${STORE_SUBDOMAIN})"
echo "  user: ${DEV_USERNAME} / ${DEV_EMAIL}"

cd "${ROOT_DIR}"

TENANT_ID="${TENANT_ID}" \
STORE_NAME="${STORE_NAME}" \
STORE_SUBDOMAIN="${STORE_SUBDOMAIN}" \
DEV_USERNAME="${DEV_USERNAME}" \
DEV_EMAIL="${DEV_EMAIL}" \
DEV_PASSWORD="${DEV_PASSWORD}" \
docker compose -f deployments/docker/infras.yml -f deployments/docker/services.yml up -d --build

echo ""
echo "Stack is up. Bootstrap runs in podzone-dev-bootstrap."
echo "Follow bootstrap logs with:"
echo "  docker compose -f deployments/docker/infras.yml -f deployments/docker/services.yml logs -f dev-bootstrap"
echo ""
echo "When bootstrap completes, open:"
echo "  http://localhost:3000/auth/dev/bootstrap"
