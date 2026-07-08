# AI Agent SDLC

This document defines how Podzone should use AI coding agents without losing
architecture consistency.

## Goal

AI agents should accelerate delivery, not create uncontrolled refactors.

Every agent task must have:

- clear business goal;
- bounded technical scope;
- known architecture rules;
- explicit verification commands;
- handoff notes.

Detailed process templates live in:

- `docs/05-process/sdlc-operating-model.md`
- `docs/05-process/spec-first-vertical-slice.md`
- `docs/05-process/feature-spec-template.md`
- `docs/05-process/ui-state-spec-template.md`
- `docs/05-process/api-contract-template.md`
- `docs/05-process/db-contract-template.md`
- `docs/05-process/vertical-slice-breakdown-template.md`
- `docs/05-process/agent-task-template.md`
- `docs/05-process/review-checklist.md`

## Source Of Truth

Agents must read the current context before coding:

1. `STATUS_CURENT.md`
2. `CLAUDE.md`
3. `agent/SKILL.md`
4. `agent/SOLID_STYLE_GUIDE.md` for frontend work
5. Domain-specific docs:
   - `internal/onboarding/README.md`
   - `docs/03-architecture-detail-design/iam-platform.md`
   - `docs/03-architecture-detail-design/ddd-clean-architecture.md`
   - `docs/03-architecture-detail-design/frontend-solid-audit.md`

If docs and code disagree, inspect code and update the relevant docs in the same
handoff or explicitly report the mismatch.

## Development Flow

Use this flow for every non-trivial change:

```text
Requirement
  -> Architecture note
  -> Task slice
  -> Implementation
  -> Verification
  -> Status update
  -> Commit
```

Do not jump from requirement straight to broad code edits when the affected
domain, contract, or done criteria are unclear.

## Task Definition

A good AI-agent task includes:

- goal;
- affected service/module;
- in-scope files or boundaries;
- out-of-scope areas;
- API, DB, proto, GraphQL, or UI contract changes;
- expected behavior;
- verification commands;
- done criteria.

Example:

```text
Goal:
Onboarding worker verifies placement route after provisioning.

Scope:
internal/onboarding only.

Contracts:
No public API change unless adding a readiness endpoint.

Done:
- readiness status is persisted;
- failure reason is visible to operators;
- store is selectable only when route resolves;
- go test ./internal/onboarding/... ./cmd/onboarding passes.
```

## Agent Roles

Prefer assigning agents by role instead of letting every agent edit everything.

### Architect Agent

- Reads docs and current code.
- Writes plans, ADRs, and task slices.
- Defines boundaries and done criteria.
- Avoids large code changes.

### Backend Agent

- Implements Go service slices.
- Follows Clean Architecture and DDD rules.
- Updates mocks and tests.
- Avoids frontend or product copy changes unless required by the task.

### Frontend Agent

- Implements one Solid module or workflow at a time.
- Follows the functional ViewModel pattern.
- Does not guess backend authorization or pagination contracts.
- Keeps expected permission and validation errors inside the current screen.

### Reviewer Agent

- Reviews diffs for bugs, regressions, architecture drift, missing tests, and
  unsafe assumptions.
- Does not rewrite the feature while reviewing unless explicitly asked.

### Docs Agent

- Updates status, requirements, architecture notes, and handoff docs.
- Does not invent runtime behavior without checking code.

## Safe Parallel Work

Good parallel lanes:

- onboarding provider/readiness/reconciliation;
- IAM role catalog and permission matrix contracts;
- Backoffice DDD context cleanup;
- frontend UX refactor for one module.

Avoid parallel edits to:

- generated mocks;
- proto contracts;
- GraphQL generated files;
- shared Solid primitives;
- broad package moves;
- service module wiring used by multiple current tasks.

If two agents need the same shared file, one agent should own that file for the
phase.

## Architecture Guardrails

### Backend

- Controllers are inbound adapters only.
- Interactors own usecase orchestration.
- Infrastructure implements output ports.
- Cross-service calls use gRPC clients, events, or projections.
- Do not import another service's domain or interactor directly.
- Transactional outbox is for commit-coupled side effects, not every job.
- Use generated mockery/testify mocks for ports unless a reusable testkit fake
  already exists.

### Frontend

- `Page.tsx` is a route/composition wrapper.
- `XView.tsx` renders layout.
- `createXViewModel.ts` owns state, resources, forms, actions, and derived
  values.
- Panels render and call ViewModel actions; panels do not call services.
- Operational lists require backend pagination, search, filters, sort, loading,
  error, empty, table/list, and pagination states.
- Frontend must not call IAM permission-check endpoints as authorization probes.
- Permission and validation errors stay in the current screen.

## Verification

Run scoped checks before full repo checks.

For Go:

```bash
go tool gofumpt -w <touched go files>
GOCACHE=/tmp/podzone-gocache go test ./internal/<service>/... ./cmd/<service>
GOCACHE=/tmp/podzone-gocache GOLANGCI_LINT_CACHE=/tmp/podzone-golangci-cache go tool golangci-lint run --timeout=5m ./internal/<service>/... ./cmd/<service>
git diff --check
```

For frontend:

```bash
cd internal/ui-podzone
npm run format
npm run lint
npm run build
npm run format:check
git diff --check
```

Use full checks when the task touches shared packages, generated contracts, or
cross-service wiring:

```bash
GOCACHE=/tmp/podzone-gocache go test ./...
make lint
```

`make coverage` is a target gate, not required for every task until coverage is
raised consistently.

## Handoff Format

Every agent handoff should include:

```text
Changed:
- files/modules touched

Behavior:
- user-visible or runtime behavior changed

Verification:
- commands run and result

Known gaps:
- what remains unfinished or risky

Suggested commit:
- type(vibe): scope - summary
```

If tests were not run, say why.

## Commit Strategy

Keep commits small and themed.

Good examples:

```text
fix(vibe): onboarding - repair placement route reconciliation
feat(vibe): onboarding - add resource health checks
refactor(vibe): ui - split admin settings view models
docs(vibe): process - document ai agent sdlc
```

Avoid commits that mix unrelated backend, frontend, generated, and docs changes.

## Roadmap Discipline

Use vertical slices with explicit exit criteria.

Current priority lanes:

1. Onboarding backbone:
   - request lifecycle;
   - placement allocation;
   - readiness checks;
   - route projection;
   - operator retry/cancel visibility.
2. IAM usability and platform extraction:
   - permission catalog;
   - role catalog;
   - role-permission matrix;
   - generic Decision API;
   - SDK and gateway adapters later.
3. Backoffice POD workflows:
   - product setup readiness;
   - partner connection;
   - first order routing;
   - fulfillment exceptions;
   - margin and settlement.
4. Frontend quality:
   - resource-owned reads;
   - typed route search;
   - list/detail or drawer workflows;
   - bounded operational tables;
   - consistent permission and validation errors.

Do not start broad refactors unless they directly support one of these lanes.
