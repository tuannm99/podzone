#!/bin/sh

set -eu

ADMIN_URL="${APISIX_ADMIN_URL:-http://apisix:9180/apisix/admin}"
ADMIN_KEY="${APISIX_ADMIN_KEY:-edd1c9f034335f136f87ad84b625c8f1}"

wait_for() {
  url="$1"
  name="$2"
  attempts="${3:-60}"
  i=0
  until curl -sS "$url" >/dev/null 2>&1; do
    i=$((i + 1))
    if [ "$i" -ge "$attempts" ]; then
      echo "Timed out waiting for $name at $url" >&2
      exit 1
    fi
    sleep 2
  done
}

put_admin() {
  path="$1"
  payload="$2"
  curl -fsS -X PUT \
    "$ADMIN_URL/$path" \
    -H "X-API-KEY: $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    --data "$payload" >/dev/null
}

wait_for "$ADMIN_URL/routes" "APISIX admin"
wait_for "http://grpc-gateway:8080/healthz" "gRPC gateway" 120
wait_for "http://backoffice-service:8000/query" "backoffice graphql" 120 || true
wait_for "http://ui-podzone:3000" "ui podzone" 120

put_admin "plugin_configs/9000" '{
  "desc": "Podzone shared edge defaults",
  "plugins": {
    "cors": {
      "allow_origins": "*",
      "allow_methods": "*",
      "allow_headers": "*"
    },
    "request-id": {}
  }
}'

put_admin "services/100" '{
  "name": "podzone-grpcgateway",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "grpc-gateway:8080": 1
    }
  }
}'

put_admin "services/110" '{
  "name": "podzone-backoffice-graphql",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "backoffice-service:8000": 1
    }
  }
}'

put_admin "services/120" '{
  "name": "podzone-ui",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "ui-podzone:3000": 1
    }
  }
}'

put_admin "consumers/podzone-dev-edge" '{
  "username": "podzone-dev-edge",
  "plugins": {
    "jwt-auth": {
      "key": "podzone-dev-edge",
      "secret": "podzone-dev-edge-secret"
    }
  }
}'

put_admin "routes/1000" '{
  "name": "podzone-api",
  "uri": "/api/*",
  "service_id": 100,
  "plugin_config_id": 9000,
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/api/(.*)", "/$1"]
    }
  }
}'

put_admin "routes/1010" '{
  "name": "podzone-graphql",
  "uri": "/query*",
  "service_id": 110,
  "plugin_config_id": 9000
}'

put_admin "routes/1020" '{
  "name": "podzone-ui",
  "uri": "/*",
  "service_id": 120,
  "plugin_config_id": 9000
}'

put_admin "routes/1030" '{
  "name": "podzone-edge-jwt-probe",
  "uri": "/edge/protected/*",
  "service_id": 100,
  "plugin_config_id": 9000,
  "plugins": {
    "jwt-auth": {},
    "proxy-rewrite": {
      "regex_uri": ["^/edge/protected/(.*)", "/$1"]
    }
  }
}'

echo "APISIX routes, services, and edge plugins seeded."
