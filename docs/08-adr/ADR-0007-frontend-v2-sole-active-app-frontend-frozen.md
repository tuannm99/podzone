# ADR-0007: `frontend-v2` becomes the sole actively developed app; `frontend/` is frozen, kept for reference only

## Status
Accepted

## Date
2026-07-25

## Related Commit
(none yet — decision only)

## Context

ADR-0005 chose Angular over SolidJS but sequenced the rollout as
**incremental, remote-by-remote, both stacks coexisting behind a
federation-bridged host** (PZEP-0004): `frontend/` (SolidJS) would keep
serving every route until each of its three remotes was individually
migrated and cut over, with the host itself migrating last.

Two things have since made that rollout strategy moot:

1. Actual delivery already diverged from it. PZEP-0008 (Approved) built the
   onboarding backbone directly in `frontend-v2` as a standalone Angular
   app talking to the real onboarding/IAM backend contracts — not through
   the planned federation-bridge cutover from a SolidJS remote. The
   "pilot remote proves the pattern, then cut over" mechanism in
   PZEP-0004's Phase 2 was never actually used.
2. Explicit user direction (2026-07-25): stop developing `frontend/`
   entirely; keep its code only as a reference for porting existing
   behavior and visual language; treat `frontend-v2` as the only frontend
   under active development going forward.

## Decision

`frontend-v2` (Angular) is the sole frontend under active development from
this date forward. `frontend/` (SolidJS) is **frozen**: no new features, no
routine maintenance — it stays in the repository only as a reference for
porting existing behavior/visual language into `frontend-v2`, not as a
live application target. Deletion of `frontend/` is not decided by this
ADR (deferred to a later cleanup decision, once nothing still depends on
reading it as reference).

`agent/ANGULAR_STYLE_GUIDE.md` is the sole binding frontend guide for any
new work from this date forward. `agent/SOLID_STYLE_GUIDE.md` remains
useful only for understanding `frontend/`'s existing code while porting
from it — it no longer governs any active development target.

## Alternatives Considered

### Option A: Continue ADR-0005/PZEP-0004's incremental coexistence plan
Pros:
- Lower risk per step; each remote's cutover independently verifiable
  against its SolidJS predecessor before removal.
- Matches the original plan's explicit rationale for avoiding a big-bang
  rewrite (no working-intermediate-state risk).

Cons:
- Actual delivery had already stopped following it (PZEP-0008 built
  standalone, not through a federation cutover) — continuing to plan
  against a sequence that isn't being executed adds confusion, not safety.
- Requires the host MF2 federation bridge (PZEP-0005) to stay a load-bearing
  dependency for every future remote's cutover, when `frontend-v2` has
  already shown it can ship real, working, backend-integrated features
  standalone without it.
- Doubles ongoing maintenance burden (both stacks must keep building/
  linting/testing clean) for a coexistence period with no fixed end date.

### Option B: This decision — consolidate on `frontend-v2`, freeze `frontend/`
Pros:
- Matches the delivery pattern already proven to work (PZEP-0008).
- Removes the host MF2 federation bridge as a *blocking* prerequisite for
  `frontend-v2` feature work (it may still be useful for other reasons,
  but is no longer on the critical path).
- One frontend to build/lint/test/deploy going forward, not two.

Cons:
- No remaining side-by-side verification against a working SolidJS
  equivalent before a `frontend-v2` feature ships — acceptance criteria
  must come from the SRS/existing behavior directly, not a live comparison.
- Some already-completed work (PZEP-0005's MF2 host bridge) was built for
  a cutover mechanism that will now see less use than originally planned.
  Not wasted — it is still real, working infrastructure — just no longer
  the load-bearing path.

Chosen: **B**, per explicit user direction (2026-07-25).

## Consequences

- PZEP-0004's Phase 2 through Phase 6 sequencing (pilot remote → IAM →
  Backoffice → host shell → decommission, each via federation cutover) is
  superseded — see the amendment note on that document. Phase 0 and
  Phase 1 already executed and remain valid, unaffected by this ADR.
- Future `frontend-v2` feature work is scoped by its own PZEP per feature
  area (the same pattern PZEP-0008 already used), not by the old phase
  list.
- `CLAUDE.md`'s Frontend section is updated to point to `frontend-v2` as
  the active app and `frontend/` as frozen/reference-only.
- No backend, contract, database, or event change of any kind — this
  decision is frontend-only, same scope boundary as ADR-0005.
- `docs/03-architecture-detail-design/15-design-system.md` remains the
  source of truth for visual language; `frontend-v2` continues
  re-implementing it in Angular/Material terms (per ADR-0006), now as the
  only active implementation target rather than one of two.

## Rule Of Thumb

New frontend work happens in `frontend-v2` only. `frontend/` is read-only
reference for porting existing behavior — do not add features, fix
non-critical bugs, or extend it.
