# Go Service Conventions For Podzone

Use this guide when changing Go code in this repo.

## Architecture

- Follow clean architecture with service-local ownership.
- Keep business rules in `internal/<service>/domain` or `internal/<service>/interactor`.
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
- Runtime subscribers, workers, clients, persistence, and external SDK wiring belong in `infrastructure`.
- Do not import another service's interactor/domain directly. Cross-service calls must go through:
  - gRPC client adapters
  - Kafka events / projections

## Messaging Placement

- Generic Kafka/runtime behavior lives in:
  - `pkg/pdkafka`
  - `pkg/messaging`
- Service-specific event handlers live in:
  - `internal/<service>/controller/eventhandler/...`
- Service-specific consumer workers live in:
  - `internal/<service>/infrastructure/messaging/...`
- Projection tables are read models, not source of truth.

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

## Lint

- Run:

```bash
golangci-lint run ./...
```

- Fix `staticcheck` issues instead of suppressing them where practical.

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
