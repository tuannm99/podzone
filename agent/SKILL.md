# Go Service Conventions For Podzone

Use this guide when changing Go code in this repo.

## Architecture

- Follow clean architecture with service-local ownership.
- Keep business rules in `internal/<service>/domain` or `internal/<service>/interactor`.
- When applying DDD, follow `docs/architecture/ddd-clean-architecture.md`.
- DDD names business boundaries; Clean Architecture controls dependency direction.
- Use package shape consistently:
  - `entity`
  - `inputport`
  - `outputport`
  - `interactor`
  - `controller/...`
  - `infrastructure/...`
- Inbound transport adapters belong in `controller`:
  - gRPC handlers
  - GraphQL resolvers
  - event handlers for Kafka projections
- Keep transport mapping in dedicated mapper files/packages under `controller`, not inline in handler/resolver files.
  - Good:
    - `internal/auth/controller/mapper/auth_mapper.go`
    - `internal/iam/controller/mapper/iam_mapper.go`
  - Handler/resolver files should only:
    - parse request
    - call interactor/usecase
    - call mapper
    - translate status/error
- Runtime subscribers, workers, clients, persistence, and external SDK wiring belong in `infrastructure`.
- Do not import another service's interactor/domain directly. Cross-service calls must go through:
  - gRPC client adapters
  - Kafka events / projections

## DDD + Clean Architecture

- A bounded context owns its language, commands, events, aggregates, input ports, output ports, and interactor.
- Prefer this context shape when a domain grows:
  - `internal/<service>/domain/<context>/entity`
  - `internal/<service>/domain/<context>/inputport`
  - `internal/<service>/domain/<context>/outputport`
  - `internal/<service>/domain/<context>/interactor`
- Aggregates expose behavior methods for state transitions; handlers and repositories must not mutate business state directly.
- Use value objects for validated business concepts and private fields when aggregate invariants need protection.
- Repository implementations should rehydrate aggregates from snapshots/documents without emitting domain events.
- Domain events are past-tense facts, e.g. `CustomerOrderReceived`, `PartnerRecommended`, `ShipmentDelivered`.
- Commands and queries are explicit structs. Split command/query input ports when read/write dependencies differ meaningfully.
- Query usecases may return read models/views directly; command usecases should use aggregates when enforcing invariants.
- Interactors orchestrate load -> domain behavior -> save -> event append/publish. They must not parse transport DTOs, know HTTP/gRPC status codes, or contain SQL.
- Interactors own transaction boundaries through output/application ports such as `TxManager`; do not inject DB clients into domain/interactor code.
- Repositories are output ports owned by the consuming context. Infrastructure implements them under `infrastructure/repository/<context>`.
- Avoid generic CRUD repositories at domain/application boundaries; use aggregate-specific repositories.
- Controllers are inbound adapters only: extract scope, map DTOs, call one usecase, map response, translate errors.
- Transport mapping belongs in `controller/mapper`; mapper code must not call repositories, usecases, or external services.
- Cross-context communication uses input ports, domain events/process policies, or explicit projections. Do not import another context's repository or interactor as a shortcut.
- Fx modules should mirror context boundaries and expose split command/query modules when the runtime boundary exists.
- Name Fx modules by their real dependency graph; if command runtime requires query dependencies, do not call it command-only.
- Every struct exposed via `fx.As(...)` must have a compile-time assertion near the concrete type.
- Constructors should return concrete types by default; use `fx.As(...)` at module boundaries.
- Add `fx.ValidateApp` tests for deployable modules and important split modules.
- New tests should use mockery/testify-generated mocks for ports, not hand-written fakes, unless the fake is a reusable testkit implementation.

## Messaging Placement

- Generic Kafka/runtime behavior lives in:
  - `pkg/pdkafka`
  - `pkg/messaging`
- Service-specific event handlers live in:
  - `internal/<service>/controller/eventhandler/...`
- Service-specific consumer workers live in:
  - `internal/<service>/infrastructure/messaging/...`
- Projection tables are read models, not source of truth.

## Messaging Architecture

- Default shape is:
  - `1 worker binary = 1 consumer runtime`
  - `1 consumer runtime = 1 registry`
  - `1 registry = many typed handlers`
- Do not start with a multi-consumer supervisor unless one binary truly owns multiple independent consumers.
- Keep API runtimes and worker runtimes separate once Kafka work is real:
  - `cmd/<service>` for API
  - `cmd/<service>-worker` for projections, outbox relays, sagas, and background consumers
- If one worker later owns multiple consumers, run each consumer in its own goroutine under a supervisor.

### Event Handler Shape

- Do not grow one big `switch envelope.Type` as event surface expands.
- Use `pkg/messaging.Registry` plus `messaging.TypedHandler`.
- Each event belongs in its own file.
- `handler.go` should only assemble the registry.

Recommended shape:

```text
internal/<service>/controller/eventhandler/<consumer>/
  handler.go
  <event_one>.go
  <event_two>.go
```

### Consumer Concurrency

- Keep message handling sequential inside a partition claim by default.
- Do not spawn goroutines per message unless offset management and ordering semantics are redesigned explicitly.
- Safe concurrency points are:
  - partitions
  - separate consumer groups
  - separate consumers under a worker supervisor
- Unsafe defaults to avoid:
  - fire-and-forget goroutines inside `Handle(...)`
  - marking offsets before background work is complete

## Config

- Prefer package-owned config loaders with defaults.
- Messaging runtime config lives under:
  - `messaging.<service>.consumers.<consumer_name>`
  - `messaging.kafka.<service>.topics`
- Core Kafka client config lives under:
  - `kafka.<service>`
- If config is missing, constructors should apply safe defaults.

## Formatting

- Prefer `gofumpt -w` for touched Go files.
- If `gofumpt` is unavailable in the environment, at least run `gofmt -w`.
- Follow the repo line-length rule enforced by `golangci-lint`:
  - max line length: `120`
  - break long call sites, composite literals, and chained conditions cleanly
  - acceptable exceptions are rare and should stay readable even when long

## Go Style

Follow the practical subset of the Uber Go Style Guide, adapted for this repo:

- Keep package names short, lower-case, and unsurprising.
- Prefer small interfaces at boundaries only.
- Do not define interfaces in the consumer package unless the boundary is truly owned there.
- Accept interfaces, return concrete types unless a boundary requires otherwise.
- Keep constructors explicit:
  - `NewX(...) *X`
  - return `(T, error)` when validation or dependency checks can fail
- Avoid `panic` in application code except unrecoverable bootstrap/programmer-error cases.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Do not use dot imports.
- Avoid stutter in names:
  - `auth.Config`, not `auth.AuthConfig` unless needed for clarity outside package boundaries
- Keep zero values meaningful where possible.
- Minimize mutable global state.
- Prefer table-driven tests for branch-heavy logic.
- Use compile-time interface assertions for important adapters and implementations:

```go
var _ outputport.SomeRepository = (*RepositoryImpl)(nil)
```

- Group imports in standard order:
  - standard library
  - third-party
  - local repo imports
- Keep functions short enough that intent is obvious; split orchestration from helpers.
- Prefer explicitness over cleverness.
- If a struct carries configuration/runtime knobs, document defaults in code.

### Naming

- Use descriptive names at boundaries, shorter names in tight local scope.
- Boolean names should read like predicates:
  - `enabled`
  - `isActive`
  - `hasProjection`
- Avoid vague names like:
  - `data`
  - `obj`
  - `info`
    unless the scope is trivially small.
- Error vars should start with `Err...`.
- Sentinel errors should be package-level `var`, not re-created dynamically.
- Acronyms should follow Go casing:
  - `ID`, `URL`, `HTTP`, `JWT`, `IAM`

### Package Design

- Keep packages cohesive.
- Avoid “utils” packages unless the content is truly cross-cutting and stable.
- If code is service-specific, keep it inside that service under `internal/<service>`.
- `pkg/*` is only for infrastructure or cross-service runtime that is genuinely shared.
- Do not move code into `pkg` just to avoid imports.

### Structs and Methods

- Use pointer receivers when:
  - the method mutates state
  - the struct is large
  - consistency across methods matters
- Avoid mixing value and pointer receivers on the same type without reason.
- Prefer constructors over exposing partially initialized structs.
- Keep struct fields ordered logically:
  - dependencies
  - config
  - runtime/cache state

### Embedding

- Do not embed types just for “shortcut syntax”.
- Use embedding when it expresses a real relationship or shared behavior.
- Prefer explicit fields over anonymous embedding in service code.

### Interfaces

- Core rule: accept interfaces, return concrete types.
- Define interfaces where the boundary is consumed or owned intentionally.
- Do not create an interface for a type that only has one implementation unless:
  - it is a boundary
  - it is used for testing/mocking
  - it isolates external transport/storage concerns
- Keep interfaces small and capability-based.
- Function parameters may use interfaces when the caller benefits from substitution.
- Function return values should prefer concrete types unless returning an interface is required by an architectural boundary.
- Do not hide concrete behavior behind an interface return “just in case”.
- Prefer:

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, key string, msg Envelope) error
}
```

over large “god interfaces”.

Examples:

```go
func NewRelay(store messaging.OutboxStore, publisher messaging.Publisher, limit int) *Relay
```

Good:

- accepts interfaces for dependencies
- returns concrete `*Relay`

```go
func NewRepository(db *sqlx.DB) outputport.Repository
```

Use only when the constructor is explicitly wiring an architectural boundary and callers should depend on that boundary.

### Nil / Error Handling

- Validate nil dependencies at construction or first use.
- Return early on invalid input.
- Do not silently swallow important errors unless the behavior is intentionally best-effort and documented.
- Use `errors.Is` / `errors.As` for matching wrapped errors.
- Log-or-return, not log-and-return at every layer.
  - Prefer logging at process/transport/worker boundaries.
  - Prefer returning wrapped errors inside domain/interactor code.

### Error Messages

- Error strings should start lower-case and avoid trailing punctuation.
- Add operation context:

```go
return fmt.Errorf("load tenant membership: %w", err)
```

- Avoid ambiguous wrappers like `operation failed`.

### Context Usage

- Pass `context.Context` as the first parameter for request/runtime scoped work.
- Do not store context in structs.
- Do not use context for optional parameters.
- Honor cancellation at I/O boundaries and long-running loops.
- Do not pass `nil` contexts.

### Concurrency

- Be explicit about ownership and lifecycle of goroutines/workers.
- Every long-running worker should have a clear start/stop path.
- Prefer lifecycle-managed workers over ad-hoc goroutines in constructors.
- Protect shared mutable state explicitly.
- Prefer channels for ownership transfer, not as a substitute for every synchronization problem.
- Avoid spawning goroutines inside libraries unless lifecycle is very clear.

### Time and Duration

- Use `time.Duration` values, not raw ints, for delays/timeouts.
- Store timestamps in UTC unless there is a strong reason not to.
- Capture `now := clock()` once when one operation should share a consistent timestamp.
- Make clocks injectable when tests need deterministic time behavior.

### Slices and Maps

- Preallocate slices when size is known or bounded.
- Return empty slices instead of `nil` when it simplifies API use, unless `nil` has semantic meaning.
- Be explicit when cloning maps/slices passed across boundaries.
- Never mutate caller-owned maps/slices unless documented.

### Constants

- Use typed constants when they model a domain concept.
- Group related constants together.
- Prefer enums via `type X string` over untyped string literals repeated across code.

### Comments and Docs

- Public symbols should have doc comments when the package is exported outside service-local code.
- Comment why, not what, when code is already readable.
- Keep TODOs actionable and specific.
- When behavior is non-obvious, document invariants close to the code.

### Logging

- Structured logging only.
- Log stable keys:
  - `tenant_id`
  - `user_id`
  - `order_id`
  - `message_id`
  - `consumer`
  - `topic`
- Do not log secrets, tokens, or full credentials.
- Prefer one meaningful log at the operational boundary over noisy logs at every helper.

### Dependency Injection

- Constructors should be deterministic and side-effect free where possible.
- Put side-effectful start/stop logic into lifecycle hooks or workers.
- `fx.Module` should describe composition, not business logic.
- Keep named providers and tags consistent across services.

### Clean Architecture Boundaries

- `controller`:
  - gRPC
  - GraphQL
  - HTTP
  - inbound event handlers
  - scheduler root
- `interactor`:
  - application usecases
  - orchestration of business rules
- `entity`:
  - domain types and rules with minimal framework coupling
- `outputport`:
  - storage/external dependency contracts
- `infrastructure`:
  - repositories
  - gRPC clients
  - Kafka workers
  - SQL/Mongo/Redis adapters

- Workers that subscribe to Kafka belong in `infrastructure`.
- Event handlers that translate consumed events into application actions belong in `controller/eventhandler`.
- Do not put transport/storage code in `interactor`.

### Testing Style

- Prefer table-driven tests for branching logic.
- Use small focused tests over giant end-to-end tests by default.
- Test both happy path and failure path.
- Assert behavior, not implementation details, where possible.
- For time-sensitive code, inject clock functions.
- For concurrency-sensitive code, keep tests deterministic and bounded.

### Generated Code and Boundaries

- Never hand-edit generated protobuf or gqlgen output.
- Keep custom mapping/support code outside generated files.
- Regenerate after proto/schema changes in the same batch.

## Lint

- Run:

```bash
golangci-lint run ./...
```

- Current repo lint set explicitly enables:
  - `govet`
  - `staticcheck`
  - `ineffassign`
  - `unused`
  - `lll`
  - `copyloopvar`
  - `errname`
  - `wastedassign`
  - `whitespace`
  - `testifylint`
  - `tparallel`
  - `unconvert`
  - `unparam`
  - `nilerr`
  - `asciicheck`
- `errcheck` is currently disabled in repo config, but unchecked errors should still be a conscious decision.
- Fix `staticcheck` and other enabled linter issues instead of suppressing them where practical.
- Code should be written to satisfy the enabled linter set by default, not patched after the fact.

## Tests

- Narrow tests first for touched packages.
- Then run broader service/package suites.
- Useful commands:

```bash
go test ./pkg/...
go test ./internal/<service>/...
go test ./cmd/<service>
make test
```

- Domain tests should avoid real DB/container usage unless the package is explicitly integration-oriented.
- Repository tests may use `pkg/testkit` and testcontainers.

## Mocks

- Use `mockery`.
- Update interfaces before regenerating mocks.
- Generate with:

```bash
make mocks
```

- Keep mocks next to the package boundary they serve.

## gRPC / Proto

- Proto changes go through `api/proto/...`.
- Regenerate with:

```bash
buf generate api/proto
```

- Keep RPC ownership strict:
  - `AuthService` for identity/session/token lifecycle
  - `IAMService` for tenant/membership/policy/authorization

## GraphQL

- Backoffice GraphQL generation:

```bash
go run github.com/99designs/gqlgen generate
```

- Keep resolvers thin. Mapping and orchestration can live in resolver support files, but business rules should stay in domain/interactor packages.

## Docs

- Architecture docs live under `docs/architecture`.
- Use Mermaid for diagrams.
- Follow existing C4 split:
  - context
  - containers
  - modules
  - sequences
  - deployment
- When changing runtime boundaries, update docs in the same batch.

## Runtime / Deployment

- Docker local stack lives under `deployments/docker`.
- Kubernetes manifests live under `deployments/kubernetes`.
- APISIX local bootstrap lives under:
  - `deployments/docker/apisix-init`
- Future infra notes for Terraform/AWS should be documented, even if not fully implemented yet.

## Review Checklist

- No cross-service code imports for business logic.
- Config has sane defaults.
- New interfaces have compile-time implementation assertions where useful.
- Tests cover failure paths, not only happy paths.
- Docs are updated if service boundary, messaging flow, or deployment wiring changed.
