# Platform Scaling And Runtime Topology — Known Problems

Identified from code and manifest review on 2026-07-23.
Scope: DB-cluster placement/capacity in `internal/onboarding/infrastructure/provisioning/`,
Kubernetes staging manifests in `deployments/kubernetes/staging/`, PgBouncer
configuration, and the shared-application runtime shape of the services.

Companion note: [Multi-tenancy issues, 2026-07-23](./multitenancy-2026-07-23.md)
covers the tenant isolation and `pdtenantdb` defects. This note covers whether
the platform **scales and survives** — cluster management, capacity admission,
and the k8s topology. The two overlap in two places, cross-referenced inline
(MT2 pool sizing, MT12 PgBouncer bypass).

Status: all items **open**. Derived from reading code and manifests — **no load
test was run and no cluster was inspected**. Numbers about throughput below are
reasoned from configuration, not measured. Re-check the referenced file/line
before treating any item as current.

---

## Summary

The architectural *shape* is right: placement resolution is separated from
business code, the cluster registry is KV-driven so clusters can be added
without redeploying, and the inventory/capacity/health entities have the right
seams for a real scheduler. What is missing is that almost none of the scaling
machinery is wired — the capacity model exists but does not count, the
multi-cluster model exists but always picks one cluster, PgBouncer is deployed
but bypassed, and every service runs `replicas: 1`.

The current topology is a development topology. It should not be read as
staging-capable.

---

## PS1 — Database cluster is selected by config name, not by load

**File:** `internal/onboarding/infrastructure/provisioning/provider/capacity.go:19-31`,
`internal/onboarding/infrastructure/provisioning/provider/provider.go:99`
**Severity:** High — multi-cluster in the data model, single-cluster in behaviour

```go
func selectDatabaseCluster(inventory entity.ResourceInventory, name string) *entity.DatabaseCluster {
	for i := range inventory.DBClusters {
		if inventory.DBClusters[i].Name == name { return &inventory.DBClusters[i] }
	}
	for i := range inventory.DBClusters {
		if inventory.DBClusters[i].Healthy { return &inventory.DBClusters[i] }
	}
	return nil
}
```

Called as `selectDatabaseCluster(inventory, p.cfg.ClusterName)`. Every tenant
lands on the cluster named in `store_provisioning.cluster_name`
(`pg-default`); the fallback is "first healthy cluster in list order", which is
still not a scheduling decision. There is no least-loaded selection, no
weighting, and no affinity/anti-affinity input.

**Fix:** replace with a real placement policy — rank candidate clusters by
remaining capacity (see PS2, which must be fixed first for any ranking to mean
anything), filter by health and region/tier requirements, and record the chosen
cluster plus the reason in the `PlacementPlan`.

---

## PS2 — Capacity counters are never incremented by placement, so admission control reads stale numbers

**File:** `internal/onboarding/domain/infrasmanager/resource_interactor.go:57-64`,
`internal/onboarding/infrastructure/provisioning/provider/capacity.go:88-99`,
`internal/onboarding/infrastructure/repository/infrasmanager/inventory_repository.go:234-236`
**Severity:** High — the capacity gate is decorative

`appendDBCapacityReasons` blocks placement on three counters:

```go
if cluster.MaxTenants > 0 && cluster.CurrentTenants >= cluster.MaxTenants { ... }
if cluster.MaxSchemas > 0 && cluster.CurrentSchemas >= cluster.MaxSchemas { ... }
if cluster.MaxConnections > 0 && cluster.CurrentConnections >= cluster.MaxConnections { ... }
```

But `current_tenants` / `current_schemas` / `current_connections` are written in
only two places: `UpdateDatabaseClusterHealth` (reached only from the manual
admin endpoint `CheckDatabaseClusterHealth`) and an admin `Upsert`. A
successful placement does **not** increment anything — a grep across
`internal/onboarding` for these fields finds no other write path.

So the gate compares against whatever the last manual health check recorded.
Between health checks it either blocks placements on a cluster that has since
drained, or admits placements onto a cluster that is already full.

**Fix:** increment the counters transactionally as part of placement
allocation, and keep the health check as a *reconciliation* pass that corrects
drift rather than as the only source of truth. Note `CheckDatabaseClusterHealth`
already counts schemas and connections live from the database
(`provider/health.go:66-93`), so the reconciler exists — it just needs to run
on a schedule instead of on demand.

---

## PS3 — No spill-over path when a cluster is full

**File:** `internal/onboarding/infrastructure/provisioning/provider/provider.go:62-75`
**Severity:** Medium — capacity exhaustion is a hard failure, not a rebalance

```go
if !capacity.CanPlace {
    return entity.PlacementPlan{}, fmt.Errorf("placement capacity unavailable: %s", ...)
}
```

Capacity is evaluated for exactly one pre-selected cluster (PS1). When it is
full, planning fails and store onboarding fails with it — there is no attempt
to evaluate a second candidate. Combined with PS2 this means the first real
capacity wall will present as "onboarding is broken" rather than "add a
cluster".

**Fix:** iterate candidate clusters in ranked order and only fail when every
candidate is exhausted; surface a distinguishable error so operators can tell
"no capacity anywhere" from "misconfigured cluster name".

---

## PS4 — `mode: database` can be resolved but never provisioned

**File:** `internal/onboarding/infrastructure/provisioning/provider/provider.go:170-187`,
`pkg/pdtenantdb/resolver_static.go:46-53`
**Severity:** Medium — the "large tenant gets a dedicated database" tier does not exist end to end

Provisioning rejects anything that is not schema mode:

```go
if plan.Mode != "schema" {
    return fmt.Errorf("unsupported postgres placement mode %q", plan.Mode)
}
```

Meanwhile `pdtenantdb` has a full read-side implementation of database-per-tenant
(`Placement.Mode == ModeDatabase`, dedicated pool accounting, `MaxDedicatedPools`,
the idle janitor, and the `BigTenants` branch producing `bo_<tenant>`). The
escape hatch for a noisy or large tenant is therefore half-built: backoffice
can route to a dedicated database, but nothing can create one.

Related: MT3 in the multi-tenancy note — the moment a `mode: database`
placement does exist, `SharedDB` misconfiguration corrupts pool accounting.

**Fix:** either implement dedicated-database provisioning
(`EnsurePostgresDatabase` already exists in `pkg/pdsql`), or explicitly
document the mode as unsupported and gate it at config validation so it cannot
be selected.

---

## PS5 — No tenant migration or rebalance path

**File:** `internal/onboarding/infrastructure/provisioning/` (absent capability)
**Severity:** Medium — placement is effectively permanent

There is no code to move a tenant's schema between databases or clusters. Once
a placement is published to `podzone/tenants/<t>/placement` it is where that
tenant lives. Combined with MT9 (2-minute placement cache with no
invalidation), even a manual move would have a window where writes land on the
old location.

**Fix:** treat this as a required capability before the platform takes real
load, and design it together with MT9 — a move needs a placement version, a
read-only/drain phase, and cache invalidation across all replicas.

---

## PS6 — PgBouncer is deployed and correctly configured, but nothing connects to it

**File:** `deployments/kubernetes/staging/persistent/pgbouncer/single.yml:16-35`,
`internal/onboarding/infrastructure/provisioning/router/placement_route_reader.go:129-155`
**Severity:** High — the pooling tier in the design is not in the data path

The PgBouncer config is right for this workload — `pool_mode = transaction`
matches the `SET LOCAL search_path` discipline in `pkg/pdtenantdb/manager.go:145`
exactly, with `max_client_conn = 500` and `default_pool_size = 20`. But the
cluster route published to KV is built from `admin_dsn`, i.e. the Postgres
service directly, so no application connection passes through it.

This is the deployment-side view of MT12 in the multi-tenancy note.

**Fix:** publish the PgBouncer endpoint as the cluster route host — but read
PS7 first, because switching naively will break queries.

---

## PS7 — Staging PgBouncer omits `max_prepared_statements`, which will break `lib/pq` under transaction pooling

**File:** `deployments/kubernetes/staging/persistent/pgbouncer/single.yml:16-35`
vs `deployments/docker/pgbouncer/pgbouncer.ini:11`
**Severity:** High (blocks PS6) — latent until PgBouncer is actually used

The docker config sets `max_prepared_statements = 100`; the staging k8s
ConfigMap does not set it at all, so it defaults to `0` — prepared statements
disabled. `lib/pq` uses the extended query protocol (implicit prepare) for
`Query`/`Exec` with arguments, which is exactly what every repository does via
squirrel. Under `pool_mode = transaction` with prepared statements disabled,
those queries fail.

So PS6 cannot be fixed on its own: pointing the cluster route at PgBouncer
today would take the service down.

**Fix:** add `max_prepared_statements` to the staging ConfigMap (matching the
docker value), or add `binary_parameters=yes` to the DSN built in
`pdtenantdb.buildDSN`. Verify with a real query path before flipping PS6.

---

## PS8 — Every service runs `replicas: 1` with no HPA and no PodDisruptionBudget

**File:** `deployments/kubernetes/staging/services/backoffice.yml:8` and the
same line in `auth.yml`, `catalog.yml`, `onboarding.yml`, `storefront.yml`,
`grpcgateway.yml`, `frontend.yml`
**Severity:** High — contradicts the shared-application multi-tenant model

A shared application serving all tenants from a single instance means every
pod restart, node drain, or rollout is a full outage for every tenant
simultaneously. There is no `HorizontalPodAutoscaler` and no
`PodDisruptionBudget` anywhere under `deployments/kubernetes/`.

**Fix:** `replicas: 2` minimum plus a PDB with `minAvailable: 1` for every
stateless service, then an HPA once resource limits are realistic (PS9). Note
this depends on PS10 — scaling out without working readiness probes makes
rollouts *worse*, not better.

---

## PS9 — CPU limit of `50m` for every Go service

**File:** `deployments/kubernetes/staging/services/backoffice.yml:38-43`
(identical block in `auth.yml`, `catalog.yml`, `onboarding.yml`,
`storefront.yml`, `grpcgateway.yml`)
**Severity:** High — severe CFS throttling, and it will be misdiagnosed

```yaml
resources:
  requests: { cpu: '20m',  memory: '50Mi' }
  limits:   { cpu: '50m',  memory: '128Mi' }
```

0.05 of a core for a service that per request does GraphQL parse/validate, two
to four gRPC round trips (`GetSession`, `GetTenantMembership`, one
`CheckPermission` per mapped field — MT14), and at least two Postgres
transactions (MT8). Requests will spend most of their wall time in CFS
throttle.

The diagnostic trap is worth calling out: throttled CPU presents as elevated
latency on the *downstream* calls, so this will look like "the database is
slow" or "IAM is slow" rather than like a CPU limit.

**Fix:** measure actual usage, then set limits with real headroom (order of
`500m`–`1` core for backoffice) before enabling an HPA — an HPA on a throttled
pod scales on a meaningless signal.

---

## PS10 — Readiness and liveness probes are commented out on all six services, for two different reasons

**File:** `deployments/kubernetes/staging/services/backoffice.yml:45-55` and the
equivalent blocks in `auth.yml`, `catalog.yml`, `onboarding.yml`,
`storefront.yml`, `grpcgateway.yml`
**Severity:** High — rollouts route traffic to pods that are not ready

Three separate defects stacked here:

1. **Wrong port.** The commented probes target `port: 8080`; every container
   listens on `8000`. Uncommenting them as-is produces a crashloop.
2. **Backoffice has no `/healthz` at all.** `cmd/backoffice/main.go:16-37` is
   the only HTTP service that does not wire
   `pdglobalmiddleware.CommonGinMiddlewareModule` — `auth`, `catalog`,
   `grpcgateway`, `iam`, `onboarding` and `partner` all do. Backoffice
   therefore also loses the shared request-logging middleware that module
   provides.
3. **The handler is not a readiness check.** Where it does exist, it returns a
   static 200 (`pkg/pdglobalmiddleware/modulehttpgin.go:38-45`) without
   verifying Mongo, Postgres, or the auth/IAM gRPC connections — so even
   enabled, it would not have protected a rollout.

**Fix:** wire `pdglobalmiddleware` into `cmd/backoffice/main.go`; correct the
probe port to `8000`; split `/healthz` (liveness, static) from `/readyz`
(readiness, checks dependencies) and point the probes at the right one.

---

## PS11 — Postgres runs as a single-replica `Deployment` on an RWO volume

**File:** `deployments/kubernetes/staging/persistent/postgres/single.yml:1-53`
**Severity:** High — the database backing every tenant has no HA and a deadlocking rollout

```yaml
kind: Deployment          # line 15 — not a StatefulSet
replicas: 1               # line 19
resources:
  limits: { cpu: '100m', memory: '256Mi' }   # lines 47-49
```

with a `ReadWriteOnce` PVC of `2Gi` and no `strategy: Recreate`. Consequences:

- Default `RollingUpdate` on an RWO volume deadlocks — the new pod cannot mount
  until the old one terminates, and the old one is not terminated until the new
  one is ready.
- `Deployment` gives no stable identity or ordered startup, both of which a
  database wants.
- `cpu: 100m` / `memory: 256Mi` is a dev sandbox allocation for the cluster that
  every tenant schema lives in. `max_connections` is untuned (default 100),
  which also caps what PgBouncer can ever do (PS6/PS7).
- No replica, no backup job, no PITR anywhere under `deployments/`.

Also: `POSTGRES_USER` / `POSTGRES_PASSWORD` are hardcoded as `postgres`/`postgres`
in the manifest (lines 34-37) rather than sourced from a Secret — the same
credential that MT1 shows is used as the application runtime credential.

**Fix:** `StatefulSet` with `strategy` appropriate for storage, realistic
resources, credentials from a Secret, and a backup strategy. If this is meant
to remain a dev-only manifest, say so explicitly in the file and keep a
separate staging manifest — right now nothing marks it as non-production.

---

## PS12 — No per-tenant quota, rate limit, or noisy-neighbour control

**File:** `internal/backoffice/` (absent capability), `pkg/pdtenantdb/manager.go:47-75`
**Severity:** High — one tenant can starve all others

Shared application, shared connection pool (10 connections by default — MT2),
no per-tenant concurrency cap, no query timeout, no rate limit at the gateway
or in the tenant middleware. A single tenant running heavy collection queries
saturates the pool and every other tenant times out. There is no mechanism to
detect which tenant caused it either (MT16 — no per-tenant metrics).

For shared-application multi-tenancy this is a required control, not an
optimisation.

**Fix:** per-tenant concurrency limiting in the tenant middleware, statement
timeouts on the tenant transaction (`SET LOCAL statement_timeout`), and
per-tenant request metrics so the offender is identifiable.

---

## Suggested Order Of Work

These split into two tracks that are largely independent and can run in
parallel.

**Track A — make the application survivable in k8s** (no tenancy changes):

1. PS10 — wire `pdglobalmiddleware` into backoffice, fix the probe port, add a
   real `/readyz`.
2. PS9 — realistic CPU/memory limits.
3. PS8 — `replicas: 2` + PDB, then HPA.
4. PS11 — Postgres to StatefulSet with real resources and secrets.

**Track B — make tenancy scale:**

1. MT1 (multi-tenancy note) — split the runtime credential from the admin DSN.
2. MT2 / MT3 — bind `pdtenantdb.Config`, size the pool.
3. MT8 — migrations off the request path.
4. PS7 then PS6 — fix `max_prepared_statements`, then route through PgBouncer.
5. PS2 — make capacity counters real.
6. PS1 / PS3 — cluster ranking and spill-over.
7. PS12 — per-tenant limits.

Two ordering constraints are real and not negotiable:

- **PS7 before PS6.** Switching the cluster route to PgBouncer without fixing
  prepared statements takes the service down.
- **MT1 before PS6.** Both change how the cluster route payload is produced
  (`clusterRoutePayload`); doing them in the other order means writing that
  code twice.

PS1, PS2, PS4 and PS5 change placement behaviour and PS11 changes the storage
topology — each needs its own sprint slice or PZEP with a migration plan. Do
not implement them straight from this note.
