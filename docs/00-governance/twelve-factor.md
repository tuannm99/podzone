# Twelve-Factor App Rules

Podzone backend services (`internal/<service>` + `cmd/<service>`) must
follow the [twelve-factor app](https://12factor.net) methodology. This
document maps each factor to Podzone's actual stack (Go + Docker + Kafka +
Postgres/Mongo/Redis) and states the rule an agent must follow, not just
the theory.

## I. Codebase

One codebase in this monorepo, tracked in git, deployed as many independent
services (`auth`, `iam`, `onboarding`, `backoffice`, `partner`, `catalog`,
`gateway`, `grpcgateway`, `storefront`).

- Do not fork a service into a separate repo to "simplify" a change.
- Do not copy-paste shared logic across services instead of using `pkg/`.

## II. Dependencies

Go modules (`go.mod`) and npm (`frontend/package-lock.json`) declare every
dependency explicitly. Nothing may rely on a tool being present on the host
outside what's pinned.

- Use `go tool <name>` for pinned dev tools (`golangci-lint`, `gofumpt`,
  `mockery`, ...) — do not assume a global install.
- Do not add a dependency without an explicit requirement or architecture
  decision (see `docs/00-governance/agent-working-rule.md`).
- Do not silently widen a version constraint to work around an issue —
  fix the actual issue or open an ADR if the widen is the fix.

## III. Config

Config comes from the environment, loaded through `pkg/pdconfig`
(`koanf`: YAML file for structure/dev defaults, overridden by env vars,
`${VAR}` placeholders expanded from the environment). Verified current
state: `.env` is git-ignored (not tracked); committed YAML under
`deployments/docker/config/*.yml` holds dev-only placeholder values
(e.g. `jwt_secret: 'dev-secret'`), not real secrets.

- Never commit a real secret, API key, or production credential — dev-only
  placeholder values in `deployments/docker/config/*.yml` are fine, real
  ones are not.
- Do not hardcode an environment-specific value (host, port, feature flag)
  in Go code — read it through `pkg/pdconfig`/`toolkit.GetEnv`.
- Do not read `CONFIG_PATH` YAML as the source of truth for secrets —
  secrets must resolve through `${VAR}` env expansion.

## IV. Backing Services

Postgres, Mongo, Redis/Valkey, Kafka are attached resources, addressed via
config (connection string/host/port), swappable without code change.

- Do not hardcode a backing-service hostname/port — Docker service name
  vs. Kubernetes DNS vs. managed-cloud endpoint must all resolve through
  the same config path.
- A service must degrade to a clear error, not a panic, if a backing
  service is unreachable at startup — see `pkg/pdserver` lifecycle hooks.

## V. Build, Release, Run

Build (compile Go binary / bundle frontend), release (Docker image + config
for one environment), and run (process execution) stay strictly separate.

- Do not read `.env` or mutate config during the run stage — config is
  fixed at release time via injected environment.
- Do not bake environment-specific values into a Docker image at build
  time — inject at container start via env vars.

## VI. Processes

Services are stateless. Verified: no `os.WriteFile`/`ioutil.WriteFile` for
persistent local state found in `internal/` outside test fixtures. Any
state that must survive a request lives in Postgres/Mongo/Redis, never on
local disk or in-process memory across requests.

- Do not cache request-scoped or session data in an in-process map/global
  that isn't safe to lose on restart or diverge across replicas.
- Do not write files to local disk for anything except `/tmp`-style
  ephemeral scratch work (e.g. build caches) — no durable state on the
  container filesystem.

## VII. Port Binding

Each service exports its own HTTP/gRPC port; there is no implicit
app-server/container coupling. `grpcgateway` is a separate service that
front-ends the gRPC ports, not part of any individual service's binary.

- A new service must bind its own port via config, not hardcode one that
  collides with another service's default.

## VIII. Concurrency

Scale out via the process model — run more instances of a stateless
service, not more threads/goroutines pretending to be independent
processes for isolation purposes.

- Do not design a feature that only works correctly with exactly one
  running instance of a service (e.g. an in-memory leader-election
  substitute) without an explicit ADR justifying why horizontal scale-out
  doesn't apply.
- Background workers (`auth-worker`, `iam-worker`) are separate deployable
  processes from their owning service, consistent with this factor —
  keep new workers structured the same way, not embedded as goroutines in
  the API-serving process.

## IX. Disposability

Fast startup, graceful shutdown on `SIGTERM`. Verified:
`pkg/pdserver/http.go` wires shutdown through `fx.Lifecycle`/`OnStop`.

- Every new long-running service/worker must register its shutdown path
  through `fx.Lifecycle` (`pkg/pdserver` pattern), not a bare
  `signal.Notify` loop reinvented per service.
- Kafka consumers must commit offsets in a way that survives an abrupt
  kill without reprocessing storms or message loss — see
  `pkg/messaging`/`pkg/pdkafka` for the existing retry/dead-letter
  pattern; do not bypass it.

## X. Dev/Prod Parity

Keep development, staging, and production as similar as possible. Docker
Compose dev stack (`deployments/docker/`) should mirror production
topology in shape (same services, same backing-service types), differing
only in scale and secret values.

- Do not add a dev-only code path (`if env == "dev" { ... different
  behavior ... }`) for anything beyond config values (log level, mock
  external calls behind an explicit test double, not a silent branch in
  business logic).
- `deployments/docker/config/*.yml` structure should track what
  production config actually needs — do not let dev config drift into
  fields production doesn't have or vice versa.

## XI. Logs

Treat logs as an event stream to stdout, not a file to manage. Verified:
`pkg/pdlog/slog_provider.go` writes to `os.Stdout` by default.

- Do not add file-based logging (log rotation, `lumberjack`, writing to
  a `.log` file) to a service — stdout only, let the runtime (Docker/K8s)
  handle log collection.
- Use `pkg/pdlog` for all logging — do not use `fmt.Println`/`log.Println`
  directly in service code (test/debug scaffolding excepted).

## XII. Admin Processes

One-off admin/management tasks (migrations, backfills, bootstrap seeding)
run as one-off processes using the same codebase and config as the
long-running service, not as manually-run ad hoc scripts with their own
config loading.

- New admin tasks belong under `cmd/<service>` or `scripts/dev/`, reusing
  `pkg/pdconfig` for config — not a standalone script that duplicates
  connection-string parsing.
- `make dev-*` targets (`dev-backoffice-seed`, `dev-auth-bootstrap`, ...)
  are the existing pattern for admin processes — follow it for new ones.

## Verification

There is no automated twelve-factor conformance check today. When adding
or reviewing a service-level change, walk the factors above that the
change touches — this is a review checklist item
(`docs/05-process/review-checklist.md`), not a CI gate yet.
