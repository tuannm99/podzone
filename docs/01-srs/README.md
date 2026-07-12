# 01 Software Requirements Specification

This folder is the product and system requirement source for implementation
planning.

Parent index: [Podzone Documentation Index](../README.md).

Start here:

1. `podzone-srs.md` — purpose, scope, actors, NFRs, and an index table into
   the per-requirement files below.
2. `traceability-matrix.md`

Individual functional requirements live one file per requirement, grouped by
domain folder (`auth/`, `iam/`, `onboarding/`, `backoffice/`, `partner/`) —
`podzone-srs.md` section 5 is the index into these, not the content itself.
`backoffice/` also holds `SRS-CAT-*`, `SRS-ORDER-*`, and `SRS-SETTLEMENT-*`
since those are DDD subdomains inside the one `internal/backoffice` service,
not separate services.

Related documents:

- `../00-project-vision/README.md`
- `../02-architecture-overall/01-c4.md`
- `../05-process/sdlc-operating-model.md`
- `../04-sprints/README.md`

## Rule

Agent tasks must trace back to an SRS requirement or a linked requirement
document. If a task cannot name its requirement ID, it is not ready for coding.

## Next Step

After selecting SRS IDs, read:

- [Recovery plan](../06-recovery/recovery-plan.md) for current stabilization scope.
- [Architecture index](../02-architecture-overall/README.md) for service boundaries and runtime design.
- [Sprint 0](../04-sprints/sprint-00-foundation.md) for agent-sized slices.
