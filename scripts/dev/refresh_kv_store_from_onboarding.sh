#!/usr/bin/env sh
set -eu

MONGO_URI="${MONGO_URI:-}"
MONGO_CONTAINER="${MONGO_CONTAINER:-podzone-mongo}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-podzone-postgres}"
PLACEMENT_AUDIT_STRICT="${PLACEMENT_AUDIT_STRICT:-false}"
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
const allocations = db.placement_allocations.aggregate([
  {$match: {status: "ready"}},
  {$sort: {updated_at: -1, _id: -1}},
  {$group: {_id: "$tenant_id", allocation: {$first: "$$ROOT"}}},
  {$replaceRoot: {newRoot: "$allocation"}},
]).toArray();

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

if [ -z "${MONGO_URI}" ]; then
  iam_tenants="$(
    docker exec "${POSTGRES_CONTAINER}" \
      psql -U "${PG_USER}" -d iam -Atc \
      "SELECT id || '|' || slug || '|' || name FROM tenants ORDER BY id"
  )"
  ready_tenants="$(
    docker exec "${MONGO_CONTAINER}" \
      mongosh -u podzone -p podzone123 --authenticationDatabase admin onboarding --quiet \
      --eval 'db.placement_allocations.distinct("tenant_id", {status: "ready"}).sort().forEach((id) => print(id))'
  )"
  orphan_tenants="$(
    printf '%s\n' "${iam_tenants}" | while IFS='|' read -r tenant_id tenant_slug tenant_name; do
      if [ -n "${tenant_id}" ] && ! printf '%s\n' "${ready_tenants}" | grep -Fxq "${tenant_id}"; then
        printf '%s|%s|%s\n' "${tenant_id}" "${tenant_slug}" "${tenant_name}"
      fi
    done
  )"

  if [ -n "${orphan_tenants}" ]; then
    echo "IAM tenants without a ready onboarding placement:" >&2
    printf '%s\n' "${orphan_tenants}" >&2
    echo "Enroll these tenants through onboarding; runtime KV refresh cannot invent a store or placement." >&2
    if [ "${PLACEMENT_AUDIT_STRICT}" = "true" ]; then
      exit 1
    fi
  else
    echo "Placement audit passed: every IAM tenant has a ready allocation."
  fi
fi
