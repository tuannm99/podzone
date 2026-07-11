# Review Workflow

Review agent output in this order:

1. Traceability
2. Scope control
3. Architecture boundary
4. Contract compatibility
5. Security and tenant isolation
6. Backend correctness
7. Frontend state/UX correctness
8. DB and migration safety
9. Tests and validation
10. Documentation updates

## Checklist

Traceability:

- Does the task link to requirement/usecase/acceptance criteria?
- Is PZEP/ADR present for cross-component or boundary changes?
- Did the agent stay within allowed files?

Architecture:

- Is business logic in the right layer?
- Did the change preserve component boundaries?
- Did it avoid unnecessary abstraction?
- Did it avoid unrelated refactor?

Backend:

- Are errors standardized?
- Is tenant/workspace/store context handled?
- Are permission checks enforced by backend?
- Are repository queries scope-safe?
- Are transactions used where needed?

Frontend:

- Does UI follow screen spec?
- Are loading/error/empty/success states handled?
- Does ViewModel own page state?
- Does API client follow contract?

Database:

- Does migration match DB spec?
- Are indexes and unique constraints scoped correctly?
- Is soft delete considered?

Testing:

- Are success and failure paths covered?
- Were relevant validation commands run?
- Are skipped checks explained?

Documentation:

- Are docs updated when behavior changed?
- Is traceability updated?
