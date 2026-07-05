#!/usr/bin/env sh
set -eu

MONGO_URI="${MONGO_URI:-}"
MONGO_CONTAINER="${MONGO_CONTAINER:-podzone-mongo}"
PG_HOST="${PG_HOST:-pgbouncer}"
PG_PORT="${PG_PORT:-6432}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-postgres}"
PG_SSL_MODE="${PG_SSL_MODE:-disable}"

echo "Rebuilding Mongo runtime KV from onboarding placement allocations..."
mongo_eval='
const pgHost = process.env.PG_HOST || "pgbouncer";
const pgPort = Number(process.env.PG_PORT || "6432");
const pgUser = process.env.PG_USER || "postgres";
const pgPassword = process.env.PG_PASSWORD || "postgres";
const pgSSLMode = process.env.PG_SSL_MODE || "disable";
const runtimeKV = db.getCollection("runtime_kv");
const clusters = new Set();
const allocations = db.placement_allocations.find({status: "ready"}).toArray();

function put(key, value) {
  runtimeKV.updateOne(
    {_id: key},
    {$set: {value: new Binary(Buffer.from(JSON.stringify(value))), updated_at: new Date()}},
    {upsert: true},
  );
  print(`Published ${key}`);
}

for (const allocation of allocations) {
  const tenantID = allocation.tenant_id;
  const storeID = allocation.store_id || "";
  const clusterName = allocation.cluster_name || "pg-default";
  const dbName = allocation.db_name || "postgres";
  const schemaName = allocation.schema_name || "";
  const mode = allocation.mode || "schema";

  clusters.add(clusterName);
  put(`podzone/tenants/${tenantID}/placement`, {
    cluster_name: clusterName,
    mode,
    db_name: dbName,
    schema_name: schemaName,
  });
  put(`podzone/tenants/${tenantID}/connections/postgres/default`, {
    tenantID,
    infraType: "postgres",
    name: "default",
    endpoint: `postgres://${pgUser}:***@${pgHost}:${pgPort}/${dbName}`,
    status: "active",
    meta: {purpose: "backoffice", cluster: clusterName, store_id: storeID},
    config: {driver: "postgres", pool: "pgbouncer"},
  });
}

for (const clusterName of clusters) {
  put(`podzone/postgres/clusters/${clusterName}`, {
    host: pgHost,
    port: pgPort,
    user: pgUser,
    password: pgPassword,
    ssl_mode: pgSSLMode,
  });
}

if (allocations.length === 0) {
  print("No ready placement allocations found in onboarding.");
}
'

if [ -n "${MONGO_URI}" ]; then
  PG_HOST="${PG_HOST}" \
    PG_PORT="${PG_PORT}" \
    PG_USER="${PG_USER}" \
    PG_PASSWORD="${PG_PASSWORD}" \
    PG_SSL_MODE="${PG_SSL_MODE}" \
    mongosh "${MONGO_URI}" --quiet --eval "${mongo_eval}"
else
  docker exec \
    -e PG_HOST="${PG_HOST}" \
    -e PG_PORT="${PG_PORT}" \
    -e PG_USER="${PG_USER}" \
    -e PG_PASSWORD="${PG_PASSWORD}" \
    -e PG_SSL_MODE="${PG_SSL_MODE}" \
    "${MONGO_CONTAINER}" \
    mongosh -u podzone -p podzone123 --authenticationDatabase admin onboarding --quiet --eval "${mongo_eval}"
fi

echo "Mongo runtime KV refreshed."
