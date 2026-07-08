# 06 Refactor And Recovery

This folder defines how to recover Podzone from the current unstable,
agent-expanded codebase.

Parent index: [Podzone Documentation Index](../README.md).

Start here:

1. `recovery-plan.md`
2. `legacy-inventory.md`
3. `backbone-flow-refactor.md`

Rule:

- Do not start broad refactors.
- Stabilize one vertical backbone flow first.
- Every refactor task must link to SRS, C4, sprint, and verification evidence.

## Current Backbone

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

Linked docs:

- [SRS baseline](../01-srs/podzone-srs.md)
- [Traceability matrix](../01-srs/traceability-matrix.md)
- [C4 architecture](../02-architecture-overall/c4.md)
- [Sprint 0](../04-sprints/sprint-00-foundation.md)
