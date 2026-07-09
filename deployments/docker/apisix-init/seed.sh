#!/bin/sh
set -eu

ADMIN_URL="${APISIX_ADMIN_URL:-http://apisix:9180/apisix/admin}"

ADMIN_KEY="${APISIX_ADMIN_KEY:-edd1c9f034335f136f87ad84b625c8f1}"

wait_for() {
  url="$1"
  name="$2"
  attempts="${3:-60}"
  i=0


  echo "Waiting for $name at $url ..."
  until curl -fsS "$url" >/dev/null 2>&1; do
    i=$((i + 1))

    if [ "$i" -ge "$attempts" ]; then
      echo "Timed out waiting for $name at $url" >&2
      return 1
    fi

    echo "  [$i/$attempts] $name not ready, retrying..."
    sleep 2
  done

  echo "  $name is ready."
}

wait_for_admin() {
  url="$1"
  name="$2"
  attempts="${3:-60}"
  i=0

  echo "Waiting for $name at $url ..."
  until curl -fsS \
    "$url" \
    -H "X-API-KEY: $ADMIN_KEY" >/dev/null 2>&1; do
    i=$((i + 1))
    if [ "$i" -ge "$attempts" ]; then
      echo "Timed out waiting for $name at $url" >&2

      return 1
    fi

    echo "  [$i/$attempts] $name not ready, retrying..."
    sleep 2
  done

  echo "  $name is ready."
}

wait_for_graphql() {
  url="$1"
  name="$2"
  attempts="${3:-60}"
  i=0

  echo "Waiting for $name at $url ..."
  until curl -fsS -X POST \
    "$url" \
    -H "Content-Type: application/json" \
    --data '{"query":"query { __typename }"}' >/dev/null 2>&1; do
    i=$((i + 1))
    if [ "$i" -ge "$attempts" ]; then
      echo "Timed out waiting for $name at $url" >&2
      return 1
    fi

    echo "  [$i/$attempts] $name not ready, retrying..."
    sleep 2
  done

  echo "  $name is ready."
}

put_admin() {
  path="$1"
  payload="$2"

  echo "PUT $ADMIN_URL/$path"


  response="$(curl -sS -w '\n%{http_code}' -X PUT \
    "$ADMIN_URL/$path" \
    -H "X-API-KEY: $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    --data "$payload")"

  status="$(printf '%s\n' "$response" | tail -n 1)"
  body="$(printf '%s\n' "$response" | sed '$d')"

  if [ "$status" -lt 200 ] || [ "$status" -ge 300 ]; then
    echo "APISIX admin error: HTTP $status" >&2
    echo "$body" >&2
    return 1

  fi
}

wait_for_admin "$ADMIN_URL/routes" "APISIX admin"
wait_for "http://grpc-gateway:8080/healthz" "gRPC gateway" 120

# Backoffice /query only supports POST and requires auth at business level.
# This check only verifies that the GraphQL HTTP endpoint is reachable.
wait_for_graphql "http://backoffice-service:8000/query" "backoffice graphql" 120 || true

# wait_for "http://frontend:3000" "ui podzone" 120

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
      "frontend:3000": 1
    }
  }
}'

put_admin "services/130" '{
  "name": "podzone-backoffice-remote",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "backoffice-remote:3001": 1
    }
  }
}'

put_admin "services/140" '{
  "name": "podzone-iam-remote",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "iam-remote:3002": 1
    }
  }
}'

put_admin "services/150" '{
  "name": "podzone-onboarding-remote",
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "onboarding-remote:3003": 1
    }
  }
}'

put_admin "consumers/podzone_dev_edge" '{
  "username": "podzone_dev_edge",
  "plugins": {
    "jwt-auth": {
      "key": "podzone_dev_edge",
      "secret": "podzone-dev-edge-secret"
    }
  }
}'

put_admin "routes/1000" '{
  "name": "podzone-api",
  "uri": "/api/*",
  "service_id": "100",

  "plugin_config_id": "9000",
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/api/(.*)", "/$1"]
    }
  }
}'

put_admin "routes/1010" '{
  "name": "podzone-graphql",
  "uri": "/query*",
  "service_id": "110",
  "plugin_config_id": "9000"
}'

put_admin "routes/1020" '{
  "name": "podzone-ui",
  "uri": "/*",
  "service_id": "120",
  "plugin_config_id": "9000"
}'

put_admin "routes/1025" '{
  "name": "podzone-backoffice-remote",
  "uri": "/mfe/backoffice/*",
  "service_id": "130",
  "plugin_config_id": "9000",
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/mfe/backoffice/(.*)", "/$1"]
    }
  }
}'

put_admin "routes/1026" '{
  "name": "podzone-iam-remote",
  "uri": "/mfe/iam/*",
  "service_id": "140",
  "plugin_config_id": "9000",
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/mfe/iam/(.*)", "/$1"]
    }
  }
}'

put_admin "routes/1027" '{
  "name": "podzone-onboarding-remote",
  "uri": "/mfe/onboarding/*",
  "service_id": "150",
  "plugin_config_id": "9000",
  "plugins": {
    "proxy-rewrite": {
      "regex_uri": ["^/mfe/onboarding/(.*)", "/$1"]
    }
  }
}'

put_admin "routes/1030" '{
  "name": "podzone-edge-jwt-probe",
  "uri": "/edge/protected/*",
  "service_id": "100",
  "plugin_config_id": "9000",
  "plugins": {
    "jwt-auth": {},
    "proxy-rewrite": {
      "regex_uri": ["^/edge/protected/(.*)", "/$1"]
    }
  }
}'

echo "APISIX routes, services, and edge plugins seeded."
