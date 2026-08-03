# PZEP-0019: Staging Kubernetes survivability, Track A (PS8, PS9, PS10, PS11)

## Status
Done — all four items (PS8, PS9, PS10, PS11) implemented. Manifests
validated for YAML syntax only; not applied to any real k8s cluster (no
cluster available in this environment) — see Rollout Log for what was and
wasn't verified.

## Date
2026-07-27

## Related Commit
(none yet — implemented in the working tree this session, not committed)

## Requirement Sources
- Business: recovery track, continuing
  [PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md) Wave 6.
- Feature: none new — this hardens existing staging deployment manifests,
  it does not add product behavior.
- Use Cases: a rollout or node drain does not take down every tenant at
  once (PS8); a service under real load is not silently CPU-throttled in a
  way that looks like "the database is slow" (PS9); Kubernetes actually
  knows when a pod is ready to receive traffic instead of guessing (PS10);
  the database backing every tenant survives a pod restart without an
  `RWO`-volume rollout deadlock and without a hardcoded credential (PS11).
- Functional/Non-functional Requirements: none in `docs/01-srs` — this is
  infra/ops hardening, not a product requirement.
- Acceptance Criteria: see below.
- UI Specs: none.

## Summary

`docs/07-problems/platform-scaling-2026-07-23.md`'s Track A
(PS8/PS9/PS10/PS11) covers whether the staging Kubernetes topology
survives a rollout, a node drain, or real load — independent of the
tenancy-scaling work in Track B (PZEP-0016 Wave 7). Re-reading the actual
current manifests before writing this PZEP (same discipline as every prior
PZEP this session) found the picture has **partially, unevenly moved since
the original note was written** — some services already have working
probes, `backoffice`'s CPU limit was already bumped, but one service
(`grpcgateway`) has a *worse*, previously-undocumented bug: its k8s
`Service`/`containerPort` declare `8080`, but the actual application
listens on `8088` (`GW_HTTP_PORT`, `pkg/pdgrpcgateway/lifecycle.go:23`) —
traffic to the Service would never reach the pod at all, a connectivity
bug, not just a missing health check.

## Problem

**PS8 — no replicas/PDB/HPA anywhere.** Confirmed still fully open:
every service (`auth`, `catalog`, `onboarding`, `storefront`,
`grpcgateway`, `backoffice`, `frontend`) is `replicas: 1`
(`deployments/kubernetes/staging/services/*.yml`). No
`PodDisruptionBudget` or `HorizontalPodAutoscaler` object exists anywhere
under `deployments/kubernetes/` (confirmed by search — zero matches).

**PS9 — CPU limits.** Partially resolved already, unevenly:
- `backoffice.yml` already has `requests: {cpu: 100m}, limits: {cpu:
  500m}` — someone bumped this outside a tracked PZEP; not the `50m`
  the original note cites.
- `auth`, `catalog`, `onboarding`, `storefront`, `grpcgateway` are all
  still `requests: {cpu: 20m}, limits: {cpu: 50m}` — the original
  finding's severity ("CFS throttling that looks like a downstream
  slowdown") still applies to five of six services.

**PS10 — readiness/liveness probes.** The original note's "commented out
on all six services" is now **stale** — re-checked each file directly:

| Service | Probe state | Correct? |
|---|---|---|
| `auth` | `tcpSocket` port `50051` | Correct — matches `containerPort: 50051` |
| `catalog` | `tcpSocket` port `50051` | Correct |
| `onboarding` | `httpGet /healthz` port `8080` | Port matches `containerPort: 8080`, but the handler behind `/healthz` (`pkg/pdglobalmiddleware/modulehttpgin.go`, shared by every Gin service) is a static `200 {"status":"ok"}` with no dependency check — not a real readiness check |
| `backoffice` | **commented out**, path `port: 8080` | Wrong even if uncommented — actual `containerPort: 8000`; `cmd/backoffice/main.go` doesn't wire `pdglobalmiddleware.CommonGinMiddlewareModule` at all, so `/healthz` doesn't exist on this service yet either |
| `storefront` | **commented out**, path `port: 8080` | Wrong even if uncommented — actual `containerPort: 3000` |
| `grpcgateway` | `httpGet /healthz` port `8088` | Port matches `GW_HTTP_PORT`, but **no `/healthz` route is registered anywhere in `pkg/pdgrpcgateway`/`cmd/grpcgateway`** — the probe would 404. Separately, and more severely: `containerPort: 8080` and the `Service`'s `port`/`targetPort: 8080` do not match `8088`, so the Service can't reach this pod's actual listener at all — a connectivity bug independent of the probe |
| `frontend` | **no probe block at all**, not even commented | Static nginx serving `/` on port 80 — lowest risk of the six, but still unverified |

**PS11 — Postgres topology.** Confirmed still fully open:
`deployments/kubernetes/staging/persistent/postgres/single.yml` is a
`Deployment` (not `StatefulSet`), `replicas: 1`, `ReadWriteOnce` PVC
(`2Gi`), `resources: {limits: {cpu: 100m, memory: 256Mi}}`, and
`POSTGRES_USER`/`POSTGRES_PASSWORD` hardcoded as `postgres`/`postgres` in
plain text in the manifest (not a Secret).

## Goals

- **PS8**: every service gets `replicas: 2` minimum plus a
  `PodDisruptionBudget` with `minAvailable: 1`. HPA is a Non-Goal (see
  below) until PS9's limits are realistic.
- **PS9**: bump `auth`, `catalog`, `onboarding`, `storefront`,
  `grpcgateway` to real resource limits (order of magnitude:
  `500m`–`1` core, matching what `backoffice` already has, sized per
  service role — gRPC-only services like `auth`/`catalog` likely need
  less than `backoffice`'s GraphQL+multi-hop path; measure before
  finalizing exact numbers, don't just copy `backoffice`'s verbatim).
- **PS10**: fix in dependency order, since some sub-items block others:
  1. `grpcgateway`'s `containerPort`/`Service.port`/`targetPort` corrected
     to `8088` (or the app changed to listen on `8080` and `GW_HTTP_PORT`
     env dropped — pick one canonical port, don't leave two) — this is a
     connectivity bug, higher severity than the probe issue riding along
     with it.
  2. Register a real `/healthz` (liveness, static — already exists for
     Gin services via `pdglobalmiddleware`, needs adding for
     `grpcgateway` specifically) and a real `/readyz` (readiness, checks
     actual dependencies: Postgres/Mongo ping, downstream gRPC
     connectivity as applicable per service) — see Proposed Solution for
     which package this belongs in.
  3. Wire `pdglobalmiddleware.CommonGinMiddlewareModule` into
     `cmd/backoffice/main.go` (currently absent entirely).
  4. Correct every probe's `path`/`port` to match reality: `backoffice`
     → `8000`, `storefront` → `3000`, `grpcgateway` → `8088` (or whatever
     the canonical port ends up being after step 1). Point readiness
     probes at `/readyz`, liveness at `/healthz`, once `/readyz` exists —
     until then, existing services keep pointing both at `/healthz` as
     they already do (no regression).
  5. Uncomment `backoffice`/`storefront`'s probe blocks once their paths
     are correct.
- **PS11**: `Deployment` → `StatefulSet` (matches the original fix note);
  realistic resources; `POSTGRES_USER`/`POSTGRES_PASSWORD` moved to a
  Kubernetes `Secret` (matching the `global-secrets` pattern
  `backoffice.yml` already uses for `BACKOFFICE_INTERNAL_SERVICE_TOKEN`).

## Non-Goals

- **HPA** (HorizontalPodAutoscaler) for any service — explicitly deferred
  until PS9's limits are real; an HPA scaling on a CPU-throttled pod's
  meaningless utilization signal would make things worse, per the
  original finding's own reasoning.
- **A real dependency-aware `/readyz` for every service in full** — this
  PZEP adds the `/readyz` *mechanism* (see Proposed Solution) and wires it
  for `backoffice`/`grpcgateway` (the two services this PZEP already
  touches for other reasons); rolling it out to `auth`/`catalog`/
  `onboarding`/`storefront` is mechanical repetition of the same pattern,
  left as follow-up rather than bundled here to keep this PZEP's diff
  reviewable.
- **Postgres backup/PITR strategy** — PS11's fix note flags this too, but
  it's a distinct operational concern (backup schedule, retention, restore
  runbook) deserving its own scoping, not a StatefulSet migration detail.
- **Any change to `deployments/docker/` (docker dev)** — this PZEP is
  staging-k8s-only; docker dev's compose-based setup is unaffected and
  intentionally different (already documented as dev-only).
- Track B (PS1-PS3, PS4, PS5, PS6, PS7, PS12) — tenancy scaling, tracked
  separately as [PZEP-0016](./PZEP-0016-recovery-backlog-triage-2026-07.md)
  Wave 7, not started.
- MT4-MT16 wiring, IAM authorization — unrelated, already closed via
  PZEP-0017/PZEP-0018.

## Proposed Solution

1. **PS8** — for every `Deployment` under
   `deployments/kubernetes/staging/services/*.yml`: set `replicas: 2`.
   Add one `PodDisruptionBudget` per service (new manifest file per
   service, or appended to each existing file — follow whichever
   convention `deployments/kubernetes/staging/` already uses for
   multi-object files, matching how `services.yml` files already stack
   `Deployment` + `Service` with `---` separators):
   ```yaml
   apiVersion: policy/v1
   kind: PodDisruptionBudget
   metadata:
     name: <service>-pdb
   spec:
     minAvailable: 1
     selector:
       matchLabels:
         app: <service>
   ```

2. **PS9** — bump resource requests/limits for `auth`, `catalog`,
   `onboarding`, `storefront`, `grpcgateway` in their respective
   manifests. Exact numbers need a measurement pass (this PZEP proposes
   the mechanism and a starting order of magnitude, not final tuned
   values — see Open Questions).

3. **PS10**:
   - `deployments/kubernetes/staging/services/grpcgateway.yml`: change
     `containerPort`, `Service.port`, and `Service.targetPort` from
     `8080` to `8088` to match `GW_HTTP_PORT`'s actual default (simpler
     fix than changing the app's listen port, since `8080` is not
     referenced by app code, only by manifest — confirm no other staging
     manifest/Ingress references `grpcgateway-service` port `8080`
     before changing, since this is also potentially a breaking change
     for whatever calls this service).
   - `pkg/pdglobalmiddleware`: add a `/readyz` handler alongside the
     existing `/healthz` — needs a way to receive a caller-supplied
     dependency-check function (e.g. `func(ctx) error`, run against
     Postgres/Mongo `Ping`), since `pdglobalmiddleware` itself has no
     knowledge of which dependencies a given service owns. Shape this as
     an optional `fx.Provide`-supplied checker, defaulting to "always
     ready" so services that don't supply one keep today's behavior
     (backward compatible, matches every other opt-in pattern from this
     session's earlier PZEPs).
   - `pkg/pdgrpcgateway`: register its own `/healthz` (static) on the
     `GW_HTTP_PORT` mux, since this service doesn't currently go through
     `pdglobalmiddleware.CommonGinMiddlewareModule` at all (it's not a Gin
     service — see `lifecycle.go`'s raw `http.Handler` + `pdserver`
     usage) — needs its own small implementation, not a reuse of the Gin
     module.
   - `cmd/backoffice/main.go`: add `pdglobalmiddleware.CommonGinMiddlewareModule`
     to the fx options list (one line, matches `cmd/onboarding/main.go`'s
     existing pattern exactly).
   - Correct every probe's `path`/`port` per the table in Problem; move
     readiness probes to `/readyz` once it exists for that service,
     otherwise leave pointed at `/healthz` (no regression for services
     this PZEP doesn't touch).

4. **PS11** — rewrite
   `deployments/kubernetes/staging/persistent/postgres/single.yml`:
   `StatefulSet` instead of `Deployment` (gives stable network identity
   and ordered startup, and sidesteps the `RWO`-volume `RollingUpdate`
   deadlock this note flags — `StatefulSet`'s default update strategy
   handles single-replica storage correctly without needing an explicit
   `Recreate` strategy override). Realistic resources (needs the same
   measurement pass as PS9, not guessed here). `POSTGRES_USER`/
   `POSTGRES_PASSWORD` sourced from a new `postgres-credentials` Secret
   instead of plain `value:` fields.

## Affected Components
- Backend: `pkg/pdglobalmiddleware`, `pkg/pdgrpcgateway`,
  `cmd/backoffice/main.go`
- Deployment: every file under `deployments/kubernetes/staging/services/`
  and `deployments/kubernetes/staging/persistent/postgres/`
- Frontend/Gateway/Worker/Database schema/External Integration: none

## API / DB / Event Contract Changes
- API: none — `/healthz`/`/readyz` are operational endpoints, not part of
  any service's product API surface.
- DB: none.
- Event: none.
- Deployment contract: `grpcgateway-service`'s exposed port changes from
  `8080` to `8088` — anything outside this repo's own manifests that
  references `grpcgateway-service:8080` directly (an Ingress, a
  cross-namespace reference) would break; confirmed no such reference
  exists inside this repo's own `deployments/` tree, but this repo does
  not necessarily contain every consumer of a staging service.

## Permission Changes
None.

## Security Considerations
PS11's credential move (hardcoded `postgres`/`postgres` → a Secret) is the
only security-relevant change here — closes a plaintext-credential-in-a
checked-in-manifest issue for the database every tenant's schema lives in.
Everything else in this PZEP is availability/observability, not a new
security decision.

## Test Plan
- Unit: `pkg/pdglobalmiddleware`'s new `/readyz` handler — test with a
  supplied checker that returns an error (expect 503) and one that
  returns nil (expect 200), and the default no-checker case (expect 200,
  unchanged from today).
- Integration: none new — no service code path beyond the health-check
  addition changes.
- E2E: not applicable from this repo — actually validating PS8 (PDB
  survives a drain), PS10 (probes gate traffic correctly), and PS11
  (StatefulSet survives a pod restart with data intact) requires a real
  or kind/minikube-based staging cluster, which is out of scope to spin up
  as part of this PZEP's automated verification.
- Manual QA: apply the updated manifests to a real staging cluster (or a
  local kind cluster as a dry run) and confirm: pods reach `Ready`, a
  manual pod delete during a `kubectl drain` doesn't take the whole
  service down (PS8), `grpcgateway-service` is actually reachable on its
  corrected port (PS10), and Postgres survives a pod restart with the PVC
  reattached correctly (PS11) — none of this is exercised by this
  session's usual `go test ./...` verification loop, since it's pure k8s
  manifest behavior.

## Agent Implementation Plan
- TASK-0001: PS8 — `replicas: 2` + PDB per service. **Done** — all 7
  service manifests.
- TASK-0002: PS9 — bump resource limits for the five services still at
  `50m`/`20m`. **Done** — matched to `backoffice`'s existing values
  (`requests: {cpu: 100m}`, `limits: {cpu: 500m}`); still not a measured
  figure, see Open Questions.
- TASK-0003: PS10 step 1 — fix `grpcgateway`'s port mismatch. **Done** —
  fixed the manifest (`GW_HTTP_PORT` 8088→8080), not the app, since docker
  dev already establishes 8080 as the canonical port for this service
  (confirmed via `deployments/docker/services.yml`) and `containerPort`/
  `Service` were already correct at 8080 - staging's env var was the one
  thing out of step.
- TASK-0004: PS10 steps 2-3 — `/readyz` mechanism in `pdglobalmiddleware`,
  `grpcgateway`'s own `/healthz`, wire `CommonGinMiddlewareModule` into
  `cmd/backoffice/main.go`. **Done.**
- TASK-0005: PS10 steps 4-5 — correct every probe path/port, uncomment
  `backoffice`/`storefront`. **Done** — also found `storefront`
  (`cmd/storefront/main.go`) had no `/healthz` handler at all (it's a
  placeholder `net/http` stub, not yet a real service) - added one, since
  uncommenting its probe without it would have crashlooped exactly as the
  original PS10 note warned.
- TASK-0006: PS11 — `StatefulSet` conversion, realistic resources, Secret
  for credentials. **Done.**
- TASK-0007: unit tests for the new `/readyz` handler. **Done** —
  `pkg/pdglobalmiddleware/modulehttpgin_test.go`.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: every staging service manifest declares `replicas: 2` and has a matching PDB with `minAvailable: 1` | TASK-0001 | Manifest review (all 7 files) + `python3 -c "yaml.safe_load_all"` syntax check |
| AC-2: `auth`/`catalog`/`onboarding`/`storefront`/`grpcgateway` have resource limits in the same order of magnitude as `backoffice`'s current values, not `50m` | TASK-0002 | Manifest review |
| AC-3: `grpcgateway-service`'s `Service`/`containerPort` match the port the application actually listens on | TASK-0003 | Manifest review + `pkg/pdgrpcgateway/lifecycle.go` cross-check (no k8s cluster available to `kubectl port-forward` against) |
| AC-4: `/readyz` returns 503 when a supplied dependency checker fails, 200 when it succeeds or none is supplied | TASK-0004, TASK-0007 | `TestNewReadyRoute` (3 cases) |
| AC-5: every service's probe `path`/`port` matches its actual listening port and handler | TASK-0005 | Manifest review per the Problem table |
| AC-6: Postgres runs as a `StatefulSet`, not a `Deployment`, with credentials from a Secret | TASK-0006 | Manifest review |

## Rollout Log

**2026-07-27, all four items implemented in one pass:**
1. **PS8**: `replicas: 2` + a `PodDisruptionBudget` (`minAvailable: 1`)
   added to all 7 service manifests (`auth`, `backoffice`, `catalog`,
   `frontend`, `grpcgateway`, `onboarding`, `storefront`).
2. **PS9**: bumped `auth`/`catalog`/`onboarding`/`storefront`/`grpcgateway`
   to `requests: {cpu: 100m}, limits: {cpu: 500m}`, matching what
   `backoffice` already had.
3. **PS10**: `grpcgateway.yml`'s `GW_HTTP_PORT` corrected `8088`→`8080` -
   confirmed via `deployments/docker/services.yml` that docker dev already
   uses `8080` consistently (`GW_HTTP_PORT: '8080'`) and that `pkg/
   pdgrpcgateway/lifecycle.go`'s own default is `8088`, so staging's env
   var was the actual outlier, not the app or `containerPort`/`Service`.
   Added `NewReadyRoute` + `ReadinessChecker` to `pkg/pdglobalmiddleware`
   (nil-checker-safe, defaults to always-ready). Wired a real `/readyz`
   for `backoffice` pinging the Mongo connection it shares with onboarding
   (`internal/backoffice/module.go`'s `newReadyRoute`, fx-populated via
   the `name:"mongo-onboarding"` tag already in the graph) and added
   `pdglobalmiddleware.CommonGinMiddlewareModule` to `cmd/backoffice/
   main.go` (previously entirely absent - the actual root cause of
   `backoffice` having no `/healthz` at all, not just a probe-config
   issue). Added a `/healthz` route directly to `pkg/pdgrpcgateway/
   lifecycle.go`'s mux (it isn't a Gin service, so it needed its own,
   separate from `CommonGinMiddlewareModule`). Discovered mid-implementation
   that `cmd/storefront/main.go` is a placeholder `net/http` stub (not a
   real service yet, no Gin, no fx) with no `/healthz` handler of any
   kind - added a minimal one directly, since uncommenting `storefront.yml`'s
   probe without it would have crashlooped. Corrected every probe's
   `path`/`port`: `backoffice` → `/readyz`+`/healthz` on `8000`,
   `storefront` → `/healthz` on `3000`, `grpcgateway` → `/healthz` on
   `8080`. `auth`/`catalog`/`onboarding` probes were already correct,
   left unchanged.
4. **PS11**: `deployments/kubernetes/staging/persistent/postgres/single.yml`
   rewritten - `StatefulSet` (headless `Service`, `volumeClaimTemplates`
   replacing the standalone `postgres-pvc`, confirmed via `grep` that
   nothing else referenced that PVC name), `requests: {cpu: 250m, memory:
   256Mi}` / `limits: {cpu: 1, memory: 512Mi}`, a real `pg_isready` exec
   probe, and a `postgres-credentials` Secret (same literal value as
   before - see the manifest's own comment on why this alone doesn't
   remove the credential from git history). Preserved the existing
   commented-out `pgadmin` scaffold at the bottom of the file untouched.

Verification: `go build ./...`, `go vet ./...`, full repo `go test
./...`, and `golangci-lint run` on every touched Go package (0 new
issues) all pass. Every edited YAML file parsed with `python3 -c
"yaml.safe_load_all(...)"` to confirm syntactic validity (no `kubectl`/
real cluster available in this environment - see Test Plan's own note on
what this does and doesn't verify).

Not done: no commit yet; no manual QA against a real or kind-based k8s
cluster (this PZEP's own Test Plan already flagged this as out of scope
for automated verification - still true, not silently skipped).

## Open Questions
- Exact PS9/PS11 resource numbers — this PZEP shipped an order-of-magnitude
  starting point (matching `backoffice`'s existing bump), not final tuned
  values; still needs a real measurement pass against staging traffic (or
  at minimum a load test) before treating these numbers as final.
- ~~`grpcgateway`'s port fix~~ — resolved: fixed the manifest's env var,
  see Rollout Log.
- Rolling `/readyz` out to `auth`/`catalog`/`onboarding`/`storefront`
  beyond this PZEP's two touched services — explicitly a Non-Goal here,
  but worth a fast follow-up once the mechanism proves out on
  `backoffice`/`grpcgateway`.
- `storefront` is still a placeholder stub service with no real
  dependencies of its own - its `/healthz` is intentionally static, not a
  gap to close later unless the service itself becomes real.
