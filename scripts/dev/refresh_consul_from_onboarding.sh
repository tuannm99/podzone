#!/usr/bin/env sh
set -eu

MONGO_URI="${MONGO_URI:-}"
MONGO_CONTAINER="${MONGO_CONTAINER:-podzone-mongo}"
CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
PG_HOST="${PG_HOST:-pgbouncer}"
PG_PORT="${PG_PORT:-6432}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_SSL_MODE="${PG_SSL_MODE:-disable}"

echo "Waiting for consul..."
until curl -fsS "${CONSUL_URL}/v1/status/leader" >/dev/null 2>&1; do
  sleep 1
done

echo "Rebuilding Consul tenant routes from onboarding placement allocations..."
mongo_eval='
const pgHost = process.env.PG_HOST || "pgbouncer";
const pgPort = Number(process.env.PG_PORT || "6432");
const pgUser = process.env.PG_USER || "postgres";
const pgPassword = process.env.PG_PASSWORD || "postgres";
const pgSSLMode = process.env.PG_SSL_MODE || "disable";
const clusters = new Set();
const allocations = db.placement_allocations.find({status: "ready"}).toArray();

for (const allocation of allocations) {
  const tenantID = allocation.tenant_id;
  const storeID = allocation.store_id || "";
  const clusterName = allocation.cluster_name || "pg-default";
  const dbName = allocation.db_name || "postgres";
  const schemaName = allocation.schema_name || "";
  const mode = allocation.mode || "schema";

  clusters.add(clusterName);

  const placement = {
    cluster_name: clusterName,
    mode,
    db_name: dbName,
    schema_name: schemaName,
  };
  print(`podzone/tenants/${tenantID}/placement|${JSON.stringify(placement)}`);

  const connection = {
    tenantID,
    infraType: "postgres",
    name: "default",
    endpoint: `postgres://${pgUser}:***@${pgHost}:${pgPort}/${dbName}`,
    status: "active",
    meta: {
      purpose: "backoffice",
      cluster: clusterName,
      store_id: storeID,
    },
    config: {
      driver: "postgres",
      pool: "pgbouncer",
    },
  };
  print(`podzone/tenants/${tenantID}/connections/postgres/default|${JSON.stringify(connection)}`);
}

for (const clusterName of clusters) {
  const cluster = {
    host: pgHost,
    port: pgPort,
    user: pgUser,
    password: pgPassword,
    ssl_mode: pgSSLMode,
  };
  print(`podzone/postgres/clusters/${clusterName}|${JSON.stringify(cluster)}`);
}
'

if [ -n "${MONGO_URI}" ]; then
  routes="$(
    PG_HOST="${PG_HOST}" \
      PG_PORT="${PG_PORT}" \
      PG_USER="${PG_USER}" \
      PG_PASSWORD="${PG_PASSWORD}" \
      PG_SSL_MODE="${PG_SSL_MODE}" \
      mongosh "${MONGO_URI}" --quiet --eval "${mongo_eval}"
  )"
else
  routes="$(
    docker exec \
      -e PG_HOST="${PG_HOST}" \
      -e PG_PORT="${PG_PORT}" \
      -e PG_USER="${PG_USER}" \
      -e PG_PASSWORD="${PG_PASSWORD}" \
      -e PG_SSL_MODE="${PG_SSL_MODE}" \
      "${MONGO_CONTAINER}" \
      mongosh -u podzone -p podzone123 --authenticationDatabase admin onboarding --quiet --eval "${mongo_eval}"
  )"
fi

if [ -z "${routes}" ]; then
  echo "No ready placement allocations found in onboarding."
  exit 0
fi

printf '%s\n' "${routes}" | while IFS= read -r route; do
  key="${route%%|*}"
  value="${route#*|}"
  curl -fsS -X PUT "${CONSUL_URL}/v1/kv/${key}" --data "${value}" >/dev/null
  echo "Published ${key}"
done

echo "Consul tenant routes refreshed."
