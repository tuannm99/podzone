# Podzone Recovery And Refactor Plan

Status: draft recovery plan.

## Problem

Podzone has grown faster than its product and architecture contracts. The
current codebase contains useful foundations, but the runtime flow is not stable
enough to keep expanding features.

The immediate goal is not to add more product surface. The immediate goal is to
make one backbone flow reliable and use it as the standard for future agent
tasks.

## Recovery Principle

Treat the current codebase as a prototype with valuable parts.

Do not rewrite everything. Do not let agents keep broad-refactoring. Recover
with strangler-style vertical slices:

1. document the intended flow;
2. inventory current code;
3. choose the smallest working slice;
4. stabilize it end to end;
5. mark or remove dead code only after replacement is proven.

## Recovery Backbone

The first flow to stabilize is:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

This maps to:

- SRS-AUTH-001
- SRS-IAM-001
- SRS-ONB-001
- SRS-ONB-002
- SRS-ONB-003
- SRS-BO-001

## Refactor Constraints

Agents must not:

- add new services;
- add new tables/collections;
- rewrite IAM;
- rewrite Backoffice;
- redesign UI broadly;
- change generated contracts;
- change DB placement source of truth;
- move code into `pkg`;
- patch symptoms without updating docs and tests.

Agents may:

- add missing requirement, architecture, or contract notes;
- add focused tests;
- thin a handler as part of one slice;
- add one usecase/repository/adapter path for a documented contract;
- remove dead code only after a verified replacement exists.

## Recovery Phases

### Phase R0: Documentation Spine

Goal:

- establish SRS, C4, sprint, process, and traceability docs.

Exit:

- `docs/README.md` links the delivery chain;
- `docs/01-srs/podzone-srs.md` exists;
- `docs/02-architecture-overall/c4.md` exists;
- `docs/04-sprints/sprint-00-foundation.md` exists;
- `docs/05-process` contains templates.

### Phase R1: Legacy Inventory

Goal:

- identify what exists, what runs, what is used, and what is unsafe.

Exit:

- `legacy-inventory.md` records services, APIs, DB ownership, packages, and
  broken flows;
- every item has a status: keep, stabilize, replace, archive, unknown.

### Phase R2: Backbone Requirement And Design

Goal:

- define the sign-in to store-scoped Backoffice flow as a product and UI spec.

Exit:

- requirement exists;
- UI state spec exists;
- permission and error states are documented;
- traceability matrix links SRS -> design -> sprint.

### Phase R3: Contracts

Goal:

- lock the minimum API/DB contracts needed for the backbone flow.

Exit:

- auth/session contract verified or documented;
- workspace/store selection contract documented;
- onboarding readiness contract documented;
- protected Backoffice API contract documented.

### Phase R4: Implementable Slices

Goal:

- break the recovery flow into agent-ready tasks.

Exit:

- each task has allowed files, out-of-scope, acceptance criteria, and
  verification;
- no task requires extra product decisions.

### Phase R5: Stabilize Runtime

Goal:

- make the backbone flow run in Docker dev without manual DB edits.

Exit:

- user can sign in;
- user can choose workspace;
- user can see pending/failed/ready store state;
- ready store resolves placement;
- Backoffice opens only ready store;
- one protected API works;
- errors remain in current screen.

### Phase R6: Expand Product Surface

Goal:

- only after R5, expand to product setup, partner, order routing, fulfillment,
  and settlement.

Exit:

- each feature follows the same SRS -> design -> contract -> sprint -> task
  chain.

## Required Agent Flow During Recovery

For every agent task:

1. Read `docs/STATUS_CURRENT.md`.
2. Read `docs/06-recovery/recovery-plan.md`.
3. Read linked SRS and sprint docs.
4. Confirm task scope and allowed files.
5. Implement only one slice.
6. Run scoped verification.
7. Handoff changed files, behavior, verification, gaps, and suggested commit.

## Recovery Sprint Mapping

| Phase | Sprint | Status |
| --- | --- | --- |
| R0 Documentation spine | Sprint 0 | In progress |
| R1 Legacy inventory | Sprint 0 | Next |
| R2 Backbone requirement/design | Sprint 0 | Next |
| R3 Contracts | Sprint 0/1 | Next |
| R4 Implementable slices | Sprint 1 | Planned |
| R5 Stabilize runtime | Sprint 1 | Planned |
| R6 Product expansion | Sprint 2+ | Later |
