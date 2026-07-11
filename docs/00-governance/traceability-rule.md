# Traceability Rule

Every implementation task must trace back to the smallest useful source of
truth.

## Required Links

| Work Type | Required Trace |
|---|---|
| UI behavior | Requirement/recovery source, screen spec, acceptance criteria |
| API integration | Requirement, contract, error shape, acceptance criteria |
| DB schema | Requirement, DB contract, owner, migration plan |
| Permission/security | Requirement, permission matrix/IAM contract, backend guard |
| Cross-component feature | Requirement, PZEP, component specs, contracts |
| Architecture boundary | ADR or architecture decision section |
| Recovery/refactor | Recovery doc, problem evidence, validation target |

## Matrix Format

Use this shape in SRS, sprint, or task docs:

| Task | Requirement | Use Case / Flow | AC | PZEP / ADR | Component | Contract | Test | Status |
|---|---|---|---|---|---|---|---|---|
| TASK-0001 | ... | ... | ... | ... | ... | ... | ... | Draft |

## Hard Rules

- No traceability, no code.
- No contract, no FE/BE/DB/event integration.
- No PZEP, no cross-component implementation.
- No ADR or architecture note, no boundary change.
- No acceptance criteria, no task assignment.
