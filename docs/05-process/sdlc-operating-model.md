# SDLC Operating Model For AI Agents

Podzone must be developed with a document-driven SDLC. AI agents may implement
only after product intent, design behavior, contracts, and task scope are clear.

## Why This Exists

Podzone is currently too large for agent-first coding. Without a delivery model,
agents tend to:

- add code faster than requirements are clarified;
- invent FE, BE, DB, and permission contracts independently;
- refactor across boundaries without a decision record;
- create screens that do not map to working workflows;
- create services and tables that are not owned by a product capability.

The fix is not "more agent work". The fix is a controlled SDLC where agents
consume small, approved implementation tasks.

## SDLC Phases

```text
0. Intake
1. Product requirement
2. UX and Figma-to-spec
3. Domain and architecture design
4. Contract design
5. DB and migration design
6. Vertical slice breakdown
7. Agent implementation
8. Human review and verification
9. Release/status update
```

Each phase has a gate. Do not move to agent implementation while upstream gates
are still open.

## Phase 0: Intake

Goal: decide whether the work is worth specifying.

Inputs:

- problem statement;
- user/operator pain;
- affected product area;
- urgency and risk.

Outputs:

- short entry in `docs/06-recovery/recovery-plan.md`, a problem note, or a
  requirement draft;
- initial owner domain;
- decision: reject, park, research, or specify.

Gate:

- the work has a named product flow or platform capability.

## Phase 1: Product Requirement

Goal: describe the behavior before designing screens or code.

Artifact:

- `docs/00-project-vision/<feature>.md`

Required sections:

- goal;
- actors;
- user flow;
- business rules;
- permissions;
- acceptance criteria;
- out of scope.

Gate:

- a reviewer can explain what must work without reading code.

## Phase 2: UX And Figma-To-Spec

Goal: turn design into implementable UI behavior.

Artifacts:

- `docs/00-project-vision/<feature>.md` with UI state sections, or a linked feature
  spec following `docs/05-process/feature-spec-template.md`.

Required sections:

- route or screen list;
- component inventory;
- fields and validation;
- actions;
- loading, empty, validation error, permission denied, API error, and success
  states;
- responsive/layout notes;
- accessibility notes;
- API calls required by each action.

Gate:

- a frontend agent can implement the screen without guessing states or copying
  design pixels blindly.

## Phase 3: Domain And Architecture Design

Goal: decide where the behavior belongs.

Artifacts:

- update existing architecture docs when possible;
- write an ADR in `docs/08-adr/` (see
  `docs/05-process/adr-template.md`) for a meaningful architecture
  decision — service boundary, data ownership, cross-service
  communication, dependency direction, or a technology/pattern choice —
  and link it from every architecture doc it affects.

Required questions:

- Which bounded context owns this behavior?
- Which service owns the data?
- Which commands and queries are required?
- Which permissions guard reads and writes?
- Does this need sync gRPC, async event, projection, or no cross-service call?
- Does this need an outbox, CDC, worker, or simple best-effort pub/sub?

Gate:

- service boundary and dependency direction are clear.

## Phase 4: Contract Design

Goal: lock the external and internal contract before implementation.

Artifacts:

- OpenAPI, proto, GraphQL schema, or REST contract notes;
- common error and pagination shape;
- DTO examples.

Required sections:

- request;
- response;
- error responses;
- auth and permission;
- pagination/search/filter/sort if list operation;
- backward compatibility and migration notes.

Gate:

- FE and BE agents can work from the same contract.

## Phase 5: DB And Migration Design

Goal: avoid agent-created tables that do not match ownership or tenant rules.

Artifacts:

- migration plan;
- ERD or table notes;
- index plan;
- tenant isolation notes.

Required questions:

- Which service owns the table/collection?
- Is this tenant/workspace/store scoped?
- What is the primary key and identity source?
- Which unique constraints include tenant scope?
- What indexes are needed for list/search/filter/sort?
- Is soft delete involved?
- Are secrets stored only as references?

Gate:

- migration and repository behavior are clear.

## Phase 6: Vertical Slice Breakdown

Goal: split the approved work into agent-sized tasks.

Artifact:

- `docs/05-process/vertical-slice-breakdown-template.md` filled for the feature, or
  a task list inside the requirement file.

Rules:

- one task should have one primary behavior;
- prefer 1 to 3 main files, rarely more than 10;
- each task has explicit allowed files or module boundaries;
- generated files are listed separately;
- verification command is scoped.

Gate:

- a coding agent can start with no extra product decisions.

## Phase 7: Agent Implementation

Goal: implement exactly one task slice.

Agent must:

- read source-of-truth docs;
- inspect existing patterns;
- state planned files before editing when the slice is substantial;
- keep changes scoped;
- update tests and generated mocks if contracts changed;
- run verification;
- provide handoff.

Gate:

- handoff includes changed files, behavior, verification, gaps, and suggested
  commit message.

## Phase 8: Human Review And Verification

Goal: protect architecture, contracts, and user behavior.

Use:

- `docs/05-process/review-checklist.md`

Review order:

1. requirement fit;
2. contract compatibility;
3. architecture boundary;
4. security and tenant scope;
5. FE states and UX;
6. DB and migration safety;
7. tests and verification.

Gate:

- reviewer accepts the diff or sends it back with specific findings.

## Phase 9: Release And Status Update

Goal: keep future agents aligned.

Update when relevant:

- `docs/06-recovery/recovery-plan.md` when recovery scope changes;
- requirement status;
- architecture docs;
- ADRs;
- known gaps.

Gate:

- the next agent can tell what is done and what remains.

## Documentation Map

Canonical structure:

```text
docs/README.md
  canonical index and reading order

docs/00-governance/
  mandatory agent rules, ready/done gates, traceability, naming, and review workflow

docs/00-project-vision/
  BA/product/domain requirements and UI state specs when needed

docs/01-srs/
  podzone-srs.md
  traceability-matrix.md

docs/02-architecture-overall/
  c4.md
  data-ownership.md
  sequences.md

docs/03-architecture-detail-design/
  transport-contracts.md
  collection-api-contract.md
  modules.md

docs/04-sprints/
  sprint-process.md
  sprint-template.md
  sprint-00-foundation.md

docs/05-process/
  sdlc-operating-model.md
  spec-first-vertical-slice.md
  feature-spec-template.md
  vertical-slice-breakdown-template.md
  agent-task-template.md
  review-checklist.md

docs/06-recovery/
  recovery plan, legacy inventory, and backbone refactor plan

docs/07-problems/
  dated issue notes and recovery evidence

docs/08-adr/
  Architecture Decision Records (ADR-NNNN)

docs/09-pzep/
  Podzone Enhancement Proposals (PZEP-NNNN)

docs/10-knowledge-base/
  concise runtime/infra incident notes from live debugging
```

Do not create parallel draft folders. Add a new top-level docs folder only when
`docs/README.md` names it and links it into the delivery chain.

## Minimum Viable SDLC For The Next 14 Days

Do not attempt to document everything at once.

Recommended sequence:

1. Create or update MVP scope.
2. Create legacy inventory.
3. Pick one backbone flow.
4. Write one requirement file for that flow.
5. Write UI state spec for that flow.
6. Lock API and DB contracts for that flow.
7. Break it into vertical slices.
8. Give only the first slice to an agent.
9. Review with the checklist.
10. Update status before the next slice.

For Podzone, the first backbone flow should be:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

Do not add broad product features until this flow is reliable.
