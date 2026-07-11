# Agent Onboarding Brief

This brief is for additional coding agents joining Podzone.

## Required Reading

Read these before changing code:

1. `agent/SKILL.md`
2. `agent/SOLID_STYLE_GUIDE.md` before touching `frontend/`
3. `docs/03-architecture-detail-design/03-ddd-clean-architecture.md` before DDD changes
4. `internal/onboarding/README.md` before onboarding or placement changes
5. `docs/03-architecture-detail-design/11-iam-platform.md` before IAM architecture changes
6. `docs/03-architecture-detail-design/13-frontend-solid-audit.md` before frontend refactors

## Current Architecture Rules

- Keep Clean Architecture boundaries: `entity`, `inputport`, `outputport`,
  `interactor`, `controller`, `infrastructure`.
- Controllers are inbound adapters only. They parse requests, call one usecase,
  map responses, and translate errors.
- Business rules stay in service-local domain/interactor packages.
- Infrastructure adapters implement output ports. Do not call providers or DB
  clients directly from controllers.
- Do not import another service's domain or interactor directly. Use gRPC
  clients, events, or explicit projections.
- Frontend must not call IAM permission-check endpoints as authorization probes.
  Backend services enforce authorization through IAM gRPC guards.
- New Go tests should use mockery/testify-generated mocks for ports unless a
  reusable testkit fake is intentional.
- Use `go tool gofumpt -w` for touched Go files and run scoped
  `golangci-lint` before handoff.
- Frontend changes must run `npm run format`, `npm run lint`, `npm run build`,
  and `npm run format:check` in `frontend/`.
- The user runs Docker hot reload. Do not start an extra frontend dev server
  unless explicitly asked.

## Product Direction

Podzone is a multi-store POD operating platform, not just an admin console.

The desired user flow is:

1. Sign in.
2. Choose workspace.
3. Request or select a store.
4. If the store is pending, show onboarding/approval state.
5. If the store is ready and placement is resolvable, enter Backoffice.

The UI should make store operations, product setup, partner connection, orders,
fulfillment, margin, and settlement visible. IAM and infrastructure should be
powerful but not dominate merchant-facing screens.

## Current Roadmap

### P0: Stabilize Onboarding Backbone

- Treat resource inventory as source of truth.
- Treat Mongo runtime KV as a rebuildable route projection only.
- Store creation is a lifecycle workflow, not a synchronous create call.
- Persist store requests, placement plans, allocations, transitions, failures,
  and readiness results.
- Fail closed when DB, Kubernetes, namespace, runtime-pool, or placement
  capacity cannot be determined.
- Keep route reconciliation and health checks operator-visible.
- Continue adding real provider checks for Kubernetes/runtime pools; Terraform
  remains a future adapter.

### P1: IAM As Reusable Authorization Platform

- Keep Podzone's current IAM working while moving toward generic IAM concepts:
  organization, application, principal, action, resource, relationship, policy,
  and decision.
- Do not encode Podzone store/order/onboarding concepts in IAM core.
- Keep management APIs and decision APIs separable.
- Keep service-side PEP enforcement authoritative. Gateway checks are
  complementary, not the only security boundary.
- Replace hardcoded UI role/permission inputs with IAM-provided catalogs,
  selectors, and role-permission matrix APIs.
- Add generic Decision API, SDK, and later Envoy/forward-auth adapters before
  extracting IAM to a separate repository.

### P2: Backoffice DDD And Product Workflows

- Backoffice remains one deployable service but should use DDD boundaries.
- Current contexts: store, catalog, routing, fulfillment, settlement, activity,
  and exception.
- GraphQL is an inbound adapter. It must not own business rules.
- Repositories belong behind context output ports.
- Make the product surface feel like store operations: catalog readiness, order
  routing, partner assignment, fulfillment status, exceptions, margin, and
  settlement.

### P3: Frontend Module Boundary And UX

- Keep `src/modules/<feature>` for feature routes and state.
- Keep `src/services/<feature>` for DTOs, HTTP/GraphQL calls, and mapping.
- Keep `src/solid` for domain-neutral shared primitives.
- Operational lists require backend pagination, search, filters, sorting,
  loading, error, empty, table/list, and pagination states.
- Expected permission and validation errors stay inside the current screen.
  They must not fall through to the route error boundary.
- Prefer tabs, bounded scroll regions, drawers/modals, and list/detail flows
  over long always-mounted pages.
- Remaining FE gaps are tracked in `docs/03-architecture-detail-design/13-frontend-solid-audit.md`.

## Safe Task Split For Multiple Agents

Good parallel work:

- One agent on onboarding provider/readiness/reconciliation.
- One agent on IAM catalog/role matrix contracts.
- One agent on Backoffice DDD/context-specific usecases.
- One agent on frontend UX refactor for a single module.

Avoid parallel edits to the same generated mocks, proto contracts, GraphQL
generated files, or shared Solid primitives unless one agent owns the merge.

## Handoff Expectations

Every handoff should state:

- files changed;
- behavior changed;
- commands run and results;
- known gaps;
- suggested commit message.

Do not claim completion without verification. If full repo checks are too broad
for the task, run scoped checks and say exactly what was not run.
