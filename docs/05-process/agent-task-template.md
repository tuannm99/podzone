# Agent Task Template

Use this prompt shape when assigning work to Claude, Codex, or another coding
agent.

```text
You are working in the Podzone monorepo.

Task:
<One small task only.>

Goal:
<What should work after this task.>

References:
- AGENTS.md
- docs/00-governance/agent-working-rule.md
- CLAUDE.md
- agent/SKILL.md
- <feature spec>
- <architecture doc>
- <existing code file>

Scope:
You may modify:
- <file or folder>
- <file or folder>

Out of scope:
- Do not change API contracts unless listed above.
- Do not create new packages unless listed above.
- Do not modify database schema unless listed above.
- Do not refactor unrelated code.
- Do not add dependencies unless explicitly required.

Architecture rules:
- Controllers are inbound adapters only.
- Interactors own usecase orchestration.
- Domain must not import infrastructure.
- Usecases depend on ports, not adapters.
- Repositories must be tenant-aware where applicable.
- Frontend components must not call services directly.
- Frontend must not call IAM permission-check endpoints as authorization probes.

Acceptance criteria:
- <Criterion>
- <Criterion>

Validation:
- Run: <command>
- Run: <command>

Handoff:
- Changed files
- Behavior changed
- Verification results
- Known gaps
- Suggested commit message
```

## Good Task Example

```text
Task:
Implement onboarding readiness query usecase only.

Goal:
Operators can ask whether one store request is blocked by missing placement,
missing route projection, or unresolved DB placement.

References:
- AGENTS.md
- docs/00-governance/agent-working-rule.md
- internal/onboarding/README.md
- docs/05-process/spec-first-vertical-slice.md
- internal/onboarding/domain/infrasmanager/placement_reconcile.go

Scope:
You may modify:
- internal/onboarding/domain/infrasmanager
- internal/onboarding/domain/infrasmanager/inputport
- internal/onboarding/domain/infrasmanager/outputport
- related mockery mocks and tests

Out of scope:
- Do not add HTTP routes.
- Do not change Mongo collection names.
- Do not change frontend UI.
- Do not change pdtenantdb.

Acceptance criteria:
- Usecase returns ready/blocked status with reason.
- Missing allocation and missing route are distinguishable.
- Tests cover ready, missing allocation, and missing route.

Validation:
- GOCACHE=/tmp/podzone-gocache go test ./internal/onboarding/domain/infrasmanager
- git diff --check
```

## Bad Task Example

```text
Build onboarding properly and clean the UI.
```

This is too broad. It has no contract, no file scope, no done criteria, and no
verification target.
