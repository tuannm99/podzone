# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Source Of Truth

Start from the numbered docs. Do not use old paths such as `docs/srs`,
`docs/process`, `docs/architecture`, `docs/refactor`, or `docs/requirements`.

Read in this order:

1. `agent/SKILL.md`
2. `docs/README.md`
3. `docs/00-project-vision/README.md`
4. `docs/01-srs/podzone-srs.md`
5. `docs/01-srs/traceability-matrix.md`
6. `docs/02-architecture-overall/README.md`
7. `docs/03-architecture-detail-design/README.md`
8. `docs/04-sprints/sprint-00-foundation.md`
9. `docs/05-process/sdlc-operating-model.md`
10. `docs/06-recovery/recovery-plan.md`

Area-specific docs:

- Frontend: `agent/SOLID_STYLE_GUIDE.md`
- Onboarding placement: `internal/onboarding/README.md`
- IAM: `docs/03-architecture-detail-design/iam-platform.md`
- DDD/Clean Architecture: `docs/03-architecture-detail-design/ddd-clean-architecture.md`
- FE refactor audit: `docs/03-architecture-detail-design/frontend-solid-audit.md`
- Design system: `docs/03-architecture-detail-design/design-system.md`

## Current Recovery Target

The first flow to stabilize is:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

Do not expand feature breadth until this backbone works in Docker dev without
manual database or KV edits.

## Agent Rules

- Do not start coding from broad requirements. Use a sprint slice or agent task
  that links to SRS, architecture, contracts, acceptance criteria, and
  verification.
- Keep changes scoped to the requested slice.
- Do not add services, packages, tables, or dependencies without an explicit
  requirement or architecture decision.
- Do not change public contracts without updating the matching docs.
- Do not move code into `pkg` unless it is genuinely shared and stable.
- Do not use frontend permission checks as security boundaries. Backend
  services must enforce authorization at inbound boundaries.

## Backend Commands

Prefer scoped checks first:

```bash
go test ./internal/<service>/...
go test ./pkg/...
GOCACHE=/tmp/podzone-gocache GOLANGCI_LINT_CACHE=/tmp/podzone-golangci-cache go tool golangci-lint run --timeout=5m ./internal/<service>/...
```

Common full checks:

```bash
make test
make coverage
make lint
make fmt
make mocks-gen
make proto
make gql-backoffice
```

Use tools pinned in `go.mod` through `go tool`.

## Frontend Commands

For `frontend/`, read `agent/SOLID_STYLE_GUIDE.md` first.

Run all four before finishing frontend work:

```bash
cd frontend
npm run format
npm run lint
npm run build
npm run format:check
```

Do not start another dev server unless explicitly asked. The user normally runs
Docker hot reload.

## Architecture Summary

- `controller/`: inbound adapters only. Parse request, call one usecase, map
  response, translate errors.
- `domain/<context>/`: aggregates, value objects, domain errors, events,
  input/output ports, interactors.
- `infrastructure/`: repository implementations, messaging runtime adapters,
  external clients.
- Cross-service calls go through gRPC adapters or Kafka events, never direct
  domain/interactor imports.
- Messaging runtime lives in `pkg/pdkafka` and `pkg/messaging`; service-specific
  event handlers live under `internal/<service>/controller/eventhandler`.

## Frontend Summary

- Route pages are thin composition roots.
- `createXViewModel` owns remote reads, mutations, loading, error, and actions.
- Panels render state and invoke actions; they do not call services directly.
- Remote reads use `createResource` or the shared pagination resource.
- Collections require search/filter/sort/loading/error/empty/pagination states.
- Internal navigation uses router/Link primitives, not `window.location.href`.

## Docs Updates

When behavior, architecture, contracts, or recovery scope changes, update the
matching numbered docs:

- Product/BA: `docs/00-project-vision`
- Requirements/traceability: `docs/01-srs`
- Overall architecture/C4: `docs/02-architecture-overall`
- Detail design/contracts: `docs/03-architecture-detail-design`
- Sprint slices: `docs/04-sprints`
- Process/templates: `docs/05-process`
- Recovery/inventory: `docs/06-recovery`
- Problem evidence: `docs/07-problems`
