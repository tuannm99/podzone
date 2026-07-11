# Sprint 0: Delivery Foundation And Backbone Flow

Status: draft.

## Goal

Create a reliable SDLC spine and stabilize the first operational backbone flow:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

## SRS Coverage

- SRS-AUTH-001
- SRS-IAM-001
- SRS-ONB-001
- SRS-ONB-002
- SRS-ONB-003
- SRS-BO-001
- SRS-NFR-006

## Source Docs

- SRS: `../01-srs/podzone-srs.md`
- Traceability: `../01-srs/traceability-matrix.md`
- C4: `../02-architecture-overall/c4.md`
- Recovery: `../06-recovery/recovery-plan.md`
- Backbone: `../06-recovery/backbone-flow-refactor.md`
- Onboarding: `../../internal/onboarding/README.md`
- Process: `../05-process/sdlc-operating-model.md`

## Scope

In:

- SDLC docs and templates;
- SRS baseline;
- C4 index;
- onboarding/backoffice backbone design;
- sprint task breakdown.

Out:

- broad product feature expansion;
- IAM open-source extraction;
- Terraform/cloud automation;
- complete UI redesign.

## Vertical Slices

### Slice 0.1: Documentation Spine

SRS:

- SRS-NFR-006

Goal:

- Add SRS, C4, process, and sprint docs that future agent tasks can trace to.

Agent role:

- Docs

Allowed files/boundaries:

- `docs/01-srs`
- `docs/05-process`
- `docs/04-sprints`
- `docs/02-architecture-overall/c4.md`

Acceptance criteria:

- docs link to each other;
- traceability matrix exists;
- sprint template exists;
- `git diff --check` passes.

Verification:

```bash
git diff --check -- docs
```

### Slice 0.2: Backbone Flow Requirement

SRS:

- SRS-AUTH-001
- SRS-IAM-001
- SRS-ONB-001
- SRS-BO-001

Goal:

- Write one requirement/design spec for the sign-in to store-scoped Backoffice
  backbone.

Agent role:

- Architect

Allowed files/boundaries:

- `docs/00-project-vision`
- `docs/01-srs/traceability-matrix.md`

Acceptance criteria:

- user flow has loading, empty, permission denied, failed placement, and ready
  states;
- linked SRS IDs are recorded.

Verification:

```bash
git diff --check -- docs/00-project-vision docs/01-srs
```

### Slice 0.2b: Legacy Inventory

SRS:

- SRS-NFR-006

Goal:

- Record which services, packages, APIs, and data stores are keep/stabilize/
  replace/archive before assigning code refactors.

Agent role:

- Docs

Allowed files/boundaries:

- `docs/06-recovery/legacy-inventory.md`
- `docs/07-problems`

Acceptance criteria:

- each service has a status;
- first broken backbone flow symptoms are listed;
- next action is documented for each broken area.

Verification:

```bash
git diff --check -- docs/06-recovery docs/07-problems
```

### Slice 0.3: Onboarding Readiness Contract

SRS:

- SRS-ONB-002
- SRS-ONB-003

Goal:

- Define contract for checking whether a store request is ready, blocked, or
  failed due to placement/route/readiness issues.

Agent role:

- Architect

Allowed files/boundaries:

- `docs/03-architecture-detail-design/transport-contracts.md`
- `docs/03-architecture-detail-design/collection-api-contract.md`
- `docs/05-process`
- `docs/01-srs/traceability-matrix.md`

Acceptance criteria:

- contract lists request, response, errors, permission, and UI behavior;
- no code implementation yet.

Verification:

```bash
git diff --check -- docs
```

### Slice 0.4: First Implementable Agent Task

SRS:

- SRS-ONB-002
- SRS-ONB-003

Goal:

- Break the readiness contract into one backend task for an agent.

Agent role:

- Architect

Allowed files/boundaries:

- `docs/04-sprints`
- `docs/05-process`

Acceptance criteria:

- task has allowed files, out-of-scope, acceptance criteria, and verification;
- task does not require extra product decisions.

Verification:

```bash
git diff --check -- docs/04-sprints docs/05-process
```

## Sprint Exit Criteria

- New agent can read SRS -> C4 -> sprint -> task without asking what to build.
- First backbone vertical slice is documented enough to implement.
- `docs/STATUS_CURRENT.md` links to SRS/process/sprint docs.
- Traceability matrix maps Sprint 0 slices to SRS IDs.
