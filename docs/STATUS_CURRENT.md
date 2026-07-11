# Current Status

Living snapshot of where Podzone recovery stands. Update this at the end of
every sprint (see `docs/04-sprints/sprint-process.md`). This is a pointer and
snapshot, not a spec — link out to the owning doc for detail.

Last updated: 2026-07-11.

## Read First

- [Documentation Index](./README.md) — canonical doc chain and start-here links.
- [Recovery Plan](./06-recovery/recovery-plan.md) — why we're in recovery mode and the phase plan.
- [Backbone Flow Refactor](./06-recovery/backbone-flow-refactor.md) — the one flow being stabilized right now.
- [Legacy Inventory](./06-recovery/legacy-inventory.md) — what exists, what's verified, what's not.

## Recovery Phase

Currently in **R0 (Documentation Spine) / R1 (Legacy Inventory)**, per
`docs/06-recovery/recovery-plan.md`'s sprint mapping. R2 (backbone
requirement/design) and R3 (contracts) are partially done ahead of schedule
for the store-readiness slice (see below).

## Backbone Flow Status

`sign in -> choose workspace -> request/select store -> onboarding resolves
placement -> open store-scoped Backoffice -> call one protected API`

- Combined store-readiness endpoint (`GET /requests/:id/readiness`) is
  **implemented** in onboarding, but **not yet called by the frontend** and
  **not verified end to end in Docker**. This is the next slice — see
  `docs/06-recovery/backbone-flow-refactor.md` "First Agent-Ready Task
  Candidate".
- Every other backbone capability (session validation, workspace
  membership, tenant DB route resolution, protected Backoffice API) is
  ⚠️ exists-but-unverified-in-Docker. No known code gap; the open work is
  running the stack and confirming the flow, not writing new handlers.

## Known Doc Debt

- `docs/01-srs/traceability-matrix.md` is mostly `TBD` — Sprint 0 exit
  criteria requires it mapped for the backbone SRS IDs; not done yet.
- `docs/01-srs/podzone-srs.md` backbone requirement IDs lack acceptance
  criteria (single "shall" sentences only).

## Not Yet Started

- R4 (implementable slices beyond store readiness), R5 (stabilize runtime
  end to end), R6 (product surface expansion) — do not start R6 work until
  R5 exit criteria in `recovery-plan.md` are met.
