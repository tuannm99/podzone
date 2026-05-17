#!/bin/sh

set -eu

SOURCE_FILE="${AUTH_BOOTSTRAP_OUTPUT:-/tmp/podzone-dev-auth.json}"
TARGET_FILE="${UI_AUTH_BOOTSTRAP_TARGET:-internal/ui-podzone/public/dev-auth-bootstrap.json}"

if [ ! -f "$SOURCE_FILE" ]; then
  echo "Auth bootstrap bundle not found: $SOURCE_FILE" >&2
  echo "Run make dev-auth-bootstrap or make dev-pod-sample first." >&2
  exit 1
fi

mkdir -p "$(dirname "$TARGET_FILE")"
cp "$SOURCE_FILE" "$TARGET_FILE"
chmod 600 "$TARGET_FILE"

echo "Synced UI dev auth bundle:"
echo "  source: $SOURCE_FILE"
echo "  target: $TARGET_FILE"
echo ""
echo "Open http://localhost:3000/auth/dev/bootstrap to import it."
