# Spec-First Vertical Slice Delivery

Podzone must stop growing from agent-first implementation. New work should move
through a locked delivery protocol before an agent writes code.

## Core Rule

No coding without a clear spec, contract, task scope, and done criteria.

Every change must answer:

- Which requirement is being implemented?
- Which contract does it follow?
- Which vertical slice does it belong to?
- Which files or boundaries may be changed?
- What proves it is done?

If those questions cannot be answered, the task is not ready.

## Delivery Pipeline

Use this order:

```text
Product requirement
  -> Figma or UI state spec
  -> Domain, API, and DB contract
  -> Architecture decision when needed
  -> Vertical task slice
  -> Agent implementation
  -> Human review
  -> Verified commit
```

For the full phase/gate model, use `docs/05-process/sdlc-operating-model.md`.

Do not ask agents to implement broad feature names such as `tenant management`
or `store onboarding`. Break them into vertical slices.

## Vertical Slice

A vertical slice is one user-visible flow across all required layers:

```text
UI state
  -> API contract
  -> usecase
  -> repository or external port
  -> DB migration or projection if needed
  -> tests
  -> docs/status update
```

Example:

```text
Create store request
  -> form validation
  -> POST request API
  -> onboarding usecase
  -> request persistence
  -> worker queue state
  -> visible pending state in UI
```

Avoid horizontal tasks such as `create all DB tables first` or `build all UI
pages first` unless the work is strictly a contract/schema phase.

## Definition Of Ready

A task is ready for an agent only when it has:

- requirement file or status reference;
- actor and user flow;
- allowed files or module boundaries;
- explicit out-of-scope list;
- API/proto/GraphQL contract if touched;
- DB contract or migration decision if touched;
- permission and tenant scope rules;
- error, loading, empty, and success states for UI work;
- acceptance criteria;
- verification commands.

## Definition Of Done

A task is done only when:

- code compiles;
- scoped tests pass;
- lint or build checks pass for touched area;
- architecture boundaries are preserved;
- handlers/components do not contain business logic;
- repository queries are tenant-aware where required;
- API and error responses follow the documented contract;
- migrations exist for DB changes;
- docs/status are updated for behavior or architecture changes;
- handoff includes changed files, behavior, verification, gaps, and commit
  message.

## Canonical Documentation Layers

Use the existing `docs/` structure. Do not create a parallel docs kit.

```text
docs/README.md
  canonical index and reading order

docs/00-project-vision/
  BA/product/domain requirements

docs/01-srs/
  system requirements and traceability

docs/02-architecture-overall/
  C4, system context, containers, data ownership, and sequences

docs/03-architecture-detail-design/
  modules, transport contracts, IAM, DDD, frontend, deployment, and OpenAPI

docs/04-sprints/
  sprint process and agent-sized vertical slices

docs/05-process/
  SDLC rules, templates, task format, review checklist

docs/06-recovery/
  recovery plan, legacy inventory, and backbone flow refactor

docs/07-problems/
  dated incident/problem notes used as recovery evidence
```

If a needed topic has no folder yet, add it only after updating
`docs/README.md` and linking it from the relevant SRS or sprint doc.

## First Recovery Flow

When the codebase feels too large to trust, start with one backbone flow:

```text
Create tenant or workspace
  -> login
  -> verify token
  -> resolve tenant/store context
  -> call one protected API
  -> read/write tenant-scoped data
```

For Podzone specifically, the current backbone is:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call protected business API
```

Do not expand feature breadth until this backbone is reliable.

## Strangler Refactor Strategy

Do not rewrite the whole system.

Use this sequence:

1. Inventory current services, APIs, packages, and tables.
2. Pick one vertical flow.
3. Stabilize that flow end to end.
4. Mark surrounding code as legacy/prototype when needed.
5. Delete or archive dead code only after a working slice replaces it.

Suggested inventory file:

```text
docs/06-recovery/legacy-inventory.md
```

Track:

- services and whether they run;
- packages and consumers;
- APIs and contract status;
- DB tables, owner service, and tenant awareness;
- generated code and regeneration commands.

## Agent Restriction Rules

Agents must not:

- create new service/package/table without spec;
- change public contracts without updating docs;
- add dependencies without explicit reason;
- refactor unrelated code;
- move code into `pkg` just to avoid imports;
- change architecture without an ADR;
- use frontend permission checks as security boundaries.

Agents may:

- update docs when code and docs disagree;
- add focused tests for touched behavior;
- update generated mocks when port contracts change;
- report blockers instead of guessing contracts.
