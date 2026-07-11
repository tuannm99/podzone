# Agent Working Rule

Podzone is developed with **Spec-first Vertical Slice Delivery**. Agents are
allowed to implement only after the target behavior is traceable to
requirements, architecture, contracts, and a scoped task.

## Delivery Chain

```text
Business need
  -> Requirement
  -> Use case
  -> Functional / non-functional requirement
  -> UI / mockup spec when frontend is involved
  -> Acceptance criteria
  -> PZEP when cross-component
  -> Architecture / component spec
  -> API / DB / event / permission contract
  -> Agent task
  -> Code
  -> Test
  -> Review
```

## Document Layers

| Layer | Current Podzone Location | Purpose |
|---|---|---|
| Governance | `docs/00-governance` | Rules, gates, naming, review, traceability |
| Product / BA | `docs/00-project-vision` | Why the product or feature exists |
| SRS | `docs/01-srs` | What the system must do and traceability |
| Architecture overall | `docs/02-architecture-overall` | C4, system context, containers, ownership |
| Detail design | `docs/03-architecture-detail-design` | Component details, contracts, DDD, IAM, frontend |
| Sprints | `docs/04-sprints` | Approved vertical slices |
| Process | `docs/05-process` | SDLC templates and task format |
| Recovery | `docs/06-recovery` | Current stabilization plan |
| Problems | `docs/07-problems` | Evidence and dated issue notes |

## Implementation Gates

An agent must not code if any required item is missing:

- requirement or recovery source;
- use case or workflow when user behavior is involved;
- acceptance criteria;
- PZEP or architecture decision for cross-component work;
- API/DB/event/permission contract for integration work;
- component or module boundary;
- allowed files or allowed folders;
- forbidden changes;
- validation commands.

If missing, stop and respond:

```text
Cannot implement safely because these required documents are missing:
- ...
```

Review-only work is allowed without a ready implementation task, but it must not
modify source code. A review-only response must report findings, missing
documents, validation results, and the next documentation or implementation gate
needed before fixes begin.

## PZEP Rule

PZEP means **Podzone Enhancement Proposal**. It is required for:

- FE + BE + DB features;
- public API or GraphQL/proto changes;
- permission/security changes;
- background jobs, Kafka, outbox, CDC, or worker changes;
- service boundary or data ownership changes;
- external integration changes.

A PZEP explains the feature-level solution. It is not a business requirement and
does not replace ADRs for architecture decisions.

Minimum PZEP sections:

- status;
- requirement sources;
- goals and non-goals;
- proposed solution;
- affected components;
- runtime flow;
- API/DB/event/permission changes;
- test plan;
- agent implementation plan.

## Contract Rule

Contracts must exist before FE/BE/DB/event integration:

- HTTP: OpenAPI or Markdown contract;
- gRPC: proto;
- GraphQL: schema and resolver contract;
- Events: AsyncAPI or event schema Markdown;
- Database: table/collection, indexes, ownership, migration notes;
- Permission: permission matrix or IAM contract.

Agents must not invent DTOs, response fields, error codes, database columns, or
permissions outside the documented contract.

## Backend Rules

Use this dependency direction:

```text
domain -> usecase/interactor -> port -> adapter -> handler
```

- Domain contains business behavior and must not import infrastructure.
- Interactor/usecase orchestrates behavior and depends on ports.
- Output ports are owned by the consuming context.
- Adapter implements ports and owns infrastructure details.
- Handler/controller maps transport only.
- Repository persists and rehydrates; it does not decide business rules.
- Tenant/workspace/store scoped queries must always filter the required scope.
- Backend services enforce authorization at inbound boundaries and delegate IAM
  evaluation through backend channels, not frontend probes.

## Frontend Rules

Use this dependency direction:

```text
Page -> ViewModel -> API Client -> Contract Types
```

- Page composes layout.
- Components render state and invoke callbacks only.
- ViewModel owns state, remote reads, mutations, loading, errors, and actions.
- API client only calls documented contracts.
- No direct fetch in UI components.
- Collections need loading, empty, error, search, filter, sort, pagination, and
  stable current page behavior.
- Permission UX must display missing permission details when backend returns
  them.

## Database Rules

- Schema changes must use migrations.
- Do not edit applied migrations unless explicitly approved for local-only reset.
- Every business table/collection must have an owner.
- Tenant-scoped business data must include tenant/workspace/store scope.
- Unique indexes must include scope unless globally unique by design.
- Soft delete must be reflected in unique constraints.
- IDs should be application-generated when domain identity is required; do not
  depend on database identity for aggregate identity unless documented.

## Agent Task Rule

Every coding task must state:

- requirement source;
- acceptance criteria;
- PZEP/ADR when applicable;
- component/module;
- contract links;
- allowed files;
- forbidden changes;
- implementation steps;
- validation commands.

Agents must keep changes scoped. Unrelated refactors, dependencies, modules,
tables, and public contract changes are rejected by default.

## Required Handoff

After implementation, agents must report:

- summary of behavior changed;
- changed files;
- tests and checks run;
- skipped checks with reason;
- risks and follow-up;
- suggested commit message.

## Final Priority

```text
Correctness > Traceability > Simplicity > Consistency > Speed
```
