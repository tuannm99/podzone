#!/bin/sh
set -e

# Reinstall if package.json changed since the volume was last populated
stamp=/app/node_modules/.install-stamp

if [ ! -f "$stamp" ] || ! cmp -s /app/package.json "$stamp" 2>/dev/null; then
    echo "[entrypoint] node_modules outdated or missing, running npm install..."
    npm install
    cp /app/package.json "$stamp"
fi

exec "$@"
