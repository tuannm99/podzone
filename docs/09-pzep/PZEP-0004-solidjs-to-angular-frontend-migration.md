# PZEP-0004: SolidJS → Angular frontend migration plan

## Status
Draft

## Date
2026-07-13

## Related Commit
(none yet — Phase 0 work below is uncommitted in the working tree)

## Requirement Sources
- Business: user directive 2026-07-13 — current SolidJS frontend is hard to
  maintain; migrate to Angular (latest, v22) incrementally, starting with a
  `frontend-v2` scaffold, keeping the existing microfrontend architecture.
- Feature: [ADR-0005](../08-adr/ADR-0005-frontend-framework-angular-replaces-solidjs.md)
  (the framework decision this PZEP sequences)
- Use Cases: (1) a developer/agent can build a new feature in Angular
  following `agent/ANGULAR_STYLE_GUIDE.md` with the same architectural
  guarantees the Solid app has today; (2) each existing remote is replaced
  behind the same host without a big-bang cutover; (3) the design system's
  visual language survives the migration unchanged.
- Functional Requirements: none new — this is a technical/process PZEP, not
  a product feature. It does not change any SRS requirement's behavior.
- Non-functional Requirements: [SRS-NFR-006](../01-srs/podzone-srs.md) AI-Agent
  Safety (each phase below requires its own follow-up PZEP/task before an
  agent starts coding it — this document sets sequence and boundaries, it is
  not itself a blanket implementation authorization for every phase)
- Acceptance Criteria: see below
- UI Specs: `docs/03-architecture-detail-design/15-design-system.md` (preserved
  as-is; re-implemented in Angular incrementally, not redesigned)

## Summary

Angular replaces SolidJS as the frontend framework (ADR-0005), migrated
**incrementally, remote by remote, behind the existing host**, not as a
big-bang rewrite. `frontend-v2/` (Angular) starts as a standalone scaffold
and becomes the real app over multiple phases, each gated by its own PZEP.
`frontend/` (SolidJS) keeps serving production traffic until each of its
pieces is individually cut over.

## Problem

`frontend/` is ~28,600 lines across a HOST + 3 Module-Federation remotes
(backoffice, iam, onboarding) + a shared design-system package. Rewriting
all of it at once is not viable without stopping all other frontend work
for an unknown, likely multi-month, duration, and offers no working
intermediate state if the rewrite stalls partway. A phased plan is needed
that:

- lets Angular code ship and be evaluated early (pilot remote) before the
  most business-critical remote (Backoffice) is touched;
- keeps the app fully working at every phase boundary — no phase should
  require both stacks to be broken simultaneously;
- does not require redesigning the visual language, re-deriving DTOs, or
  changing any backend contract — this is a frontend-only, framework-only
  change.

The one hard technical prerequisite, confirmed this session: the host's
current federation plugin (`@originjs/vite-plugin-federation`) cannot load
an Angular native-federation remote — different manifest/runtime protocol
(verified: the host's remotes use a Webpack-MF-v1-style `remoteEntry.js`;
`frontend-v2`'s live `remoteEntry.json` is native-federation's own
`$version: "v4"` format). No remote can be cut over to Angular until the
host itself migrates to a federation runtime both stacks can speak
(`@module-federation/vite`, MF2) — this is Phase 1, and it is the highest-risk
single step in the whole plan because it touches a currently-working system
before any Angular code is user-facing.

## Goals

- A working Angular remote is loadable by the real host (not just standalone).
- Each of the 3 existing remotes (onboarding, iam, backoffice) is migrated
  to Angular one at a time, smallest/lowest-risk first, each behind its own
  PZEP and each independently revertible (old remote stays deployable until
  its replacement is verified equivalent).
- The host shell itself (`frontend/src` — navbar, auth bootstrap, workspace
  switcher, top-level routing) migrates last, once all three remotes it
  hosts are already Angular — minimizes the time two different "kinds" of
  host logic must be reasoned about at once.
- `agent/ANGULAR_STYLE_GUIDE.md` (this PZEP's Phase 0 deliverable) is the
  binding reference for every line of Angular code written in any phase.
- The design system's Tailwind tokens/classes
  (`docs/03-architecture-detail-design/15-design-system.md`) are preserved;
  Angular components are new implementations of the same visual contract,
  not a redesign.

## Non-Goals

- No backend, API contract, database, or event schema change of any kind,
  in any phase.
- No new product features introduced "while we're in there" — each
  remote's Angular version must be behavior-equivalent to its SolidJS
  predecessor at cutover time. Feature work continues to go through the
  normal SRS/sprint-task process independently of this migration.
- No redesign of the visual language — colors, spacing, component shapes
  from `15-design-system.md` carry over unchanged.
- Full task-level breakdown for Phases 1-5 — each is its own follow-up PZEP
  written when that phase actually starts (per `docs/00-governance/agent-working-rule.md`:
  no PZEP, no cross-component implementation). This document fixes the
  **sequence and boundaries**, not the line-by-line implementation of every
  phase.
- Deleting `frontend/` now. It is not frozen either — bug fixes and small
  feature work continue there per normal process until its last remote is
  cut over.

## Proposed Solution — Phase Sequence

### Phase 0 — Foundations (this session, mostly done)
- `ADR-0005` (framework decision).
- `agent/ANGULAR_STYLE_GUIDE.md` (this PZEP's companion deliverable).
- `frontend-v2/` scaffolded: Angular 22, native-federation configured as a
  remote on port 3004, verified live (`ng serve`, `ng build`, `ng test` all
  pass). Standalone only — not wired into the real host yet.
- Root `README.md` documents the new dev-tool requirement (`@angular/cli`).
- Status: done except the style guide, which this PZEP ships alongside.

### Phase 1 — Host federation migration (MF2)
**Status: Done (uncommitted), 2026-07-13 — see [PZEP-0005](./PZEP-0005-host-federation-migration-to-mf2.md).**
The spike found the original assumption below incomplete: MF2 and Angular's
native-federation are separate ecosystems requiring a bridge
(`native-to-mf-bridge`) — see PZEP-0005 for the real findings, including a
genuine remoteEntry-path bug found and fixed along the way.

Migrate `frontend/vite.config.ts`'s federation plugin from
`@originjs/vite-plugin-federation` to `@module-federation/vite`. This must
prove **two** things before it's considered done, not one:
1. All 3 existing SolidJS remotes (backoffice, iam, onboarding) still load
   correctly into the migrated host — a regression here breaks the entire
   live app, not just the migration.
2. The Angular `frontend-v2` remote loads into the same migrated host for
   the first time — real cross-framework federation, not a standalone
   `ng serve`.

Highest risk phase: it changes the one component every other remote already
depends on. Needs its own PZEP with a rollback plan before any code is
written, and should be spiked/verified in a branch before touching the
default dev environment.

### Phase 2 — Pilot remote: onboarding (smallest, ~31 files)
Rewrite the onboarding remote in Angular inside `frontend-v2`, following
`agent/ANGULAR_STYLE_GUIDE.md`. Port only the design-system components
onboarding actually needs — not the whole `packages/shared` inventory
upfront. Cut over onboarding's routes from the SolidJS remote to the
Angular remote once verified behavior-equivalent (same acceptance criteria
onboarding already has in `docs/01-srs/onboarding/`). Chosen as the pilot
specifically because it is the smallest remote and lets the whole
pattern — host wiring, design-system porting, style-guide fit — be proven
once, cheaply, before the two larger, more business-critical remotes.

### Phase 3 — IAM remote (~63 files)
Same process as Phase 2, applied to IAM. Larger and more structurally
complex (organizations, policies, groups, principals, trust/simulation) —
scheduled after the pilot proves the pattern, not before.

### Phase 4 — Backoffice remote (~58 files)
Same process, applied last among the three remotes because it is the most
business-critical (the store-scoped operational surface — orders, catalog,
settlement) and benefits most from the pattern already being proven twice.

### Phase 5 — Host shell (`frontend/src`)
Once all three remotes it hosts are Angular, migrate the host shell itself
(navbar, auth bootstrap, workspace switcher, top-level routing) to Angular.
Deliberately last: the host is the one component every remote depends on
throughout Phases 2-4, so it stays SolidJS (stable, already migrated to
MF2 in Phase 1) until nothing else needs it to be.

### Phase 6 — Decommission
Delete `frontend/` (SolidJS) entirely. Remove `agent/SOLID_STYLE_GUIDE.md`'s
mandatory status from `CLAUDE.md` (the file itself can stay as historical
record or be deleted — decide at the time). Rename `frontend-v2/` to
`frontend/` or keep the `-v2` suffix — decide at the time based on whatever
is least disruptive to in-flight tooling/CI references at that point.

## Affected Components
- Frontend: all of it, eventually (`frontend/`, new `frontend-v2/`)
- Gateway: none
- Backend: none
- Worker: none
- Database: none
- External Integration: none

## API Contract Changes
- None, in any phase. Angular remotes call the exact same REST/GraphQL/gRPC-gateway
  endpoints the SolidJS remotes call today.

## DB Contract Changes
- None.

## Event Contract Changes
- None.

## Permission Changes
- None — authorization stays entirely backend-enforced per
  `docs/00-governance/agent-working-rule.md` ("Do not use frontend
  permission checks as security boundaries"), regardless of which frontend
  framework renders the UI.

## Data Ownership
- Unchanged — this PZEP does not move data ownership between services.

## Security Considerations
- Authentication/authorization: unchanged, backend-enforced, framework-agnostic.
- Tenant/workspace/store isolation: unchanged.
- Sensitive data: none added.
- New risk specific to this migration: **during Phases 2-4, two
  implementations of the same remote's business logic exist
  simultaneously** (old SolidJS route, new Angular route, cut over
  atomically once verified). Each phase's own PZEP must specify how the
  cutover is verified equivalent before the old code path is removed, to
  avoid a silent behavior regression reaching users.

## Observability
- No new requirement beyond what each phase's own PZEP specifies. Recommend
  (not required by this PZEP) a temporary feature flag or route-level
  toggle per remote during Phases 2-4, so a cutover can be reverted without
  a redeploy if a regression is found — decide the mechanism in that
  phase's own PZEP.

## Alternatives Considered

See ADR-0005's Alternatives Considered for the framework choice itself.
This PZEP's own alternative is about **sequencing**:

### Alternative: Migrate the host first, all three remotes in parallel
Pros: shorter total wall-clock time if enough parallel capacity exists.
Cons: no pilot to catch host-wiring or style-guide problems before they're
repeated three times at once; three simultaneous large migrations are much
harder to review and roll back independently than one pilot followed by
two informed repeats. Rejected — smallest-remote-first pilot is cheaper to
get wrong and cheaper to learn from.

### Alternative: Big-bang rewrite (no `frontend-v2`, no host)
Pros: no dual-stack period, no federation-protocol migration needed.
Cons: no working intermediate state if the rewrite stalls; blocks all other
frontend feature work for the rewrite's entire duration; directly
contradicts the user's explicit instruction ("k cần sửa toàn bộ now").
Rejected.

## Test Plan
- Unit/Integration/E2E: defined per-phase in that phase's own PZEP — Phase
  0 (this document) has no new runtime behavior to test beyond what was
  already verified live this session (`ng serve`/`ng build`/`ng test`,
  `remoteEntry.json` served correctly).
- Manual QA: none for Phase 0. Phase 1 onward requires manual QA of the
  full existing app (all 3 SolidJS remotes) after any host federation
  change, before Angular-specific testing even starts — a host regression
  is a production incident regardless of this migration.

## Agent Implementation Plan
- TASK-0001: Write `agent/ANGULAR_STYLE_GUIDE.md` (this PZEP's Phase 0
  deliverable, shipped alongside this document).
- TASK-0002: Update `CLAUDE.md`'s Frontend section to reference both style
  guides, scoped by path (`frontend/` → Solid guide, `frontend-v2/` →
  Angular guide), and note the migration's existence/status.
- Phases 1 through 6: **no tasks are authorized by this PZEP.** Each phase
  requires its own PZEP (with its own Agent Implementation Plan, Acceptance
  Criteria, and rollback plan) written and accepted before an agent starts
  implementation — per `docs/00-governance/agent-working-rule.md`. This
  document's job is to fix the phase order and the boundaries between
  phases, not to pre-authorize their implementation.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: `agent/ANGULAR_STYLE_GUIDE.md` exists and covers the same category list as `SOLID_STYLE_GUIDE.md` (boundaries, reactivity, async/mutations, forms, collections, routing/a11y, MFE singleton rules) | TASK-0001 | Doc review |
| AC-2: `CLAUDE.md` correctly scopes which style guide governs which path | TASK-0002 | Doc review |
| AC-3: No phase beyond Phase 0 has any code merged without its own accepted PZEP | all future phases | Governance review at each phase boundary |
| AC-4: `frontend/` continues to build/lint/test clean throughout every phase (verified by that phase's own PZEP test plan, not just Phase 0) | Phase 1+ | `cd frontend && npm run build && npm run lint` |

## Open Questions
- Exact target date/owner for Phase 1 (host MF2 migration) — not set by
  this PZEP; depends on when the user wants to resume this work.
- Whether `@module-federation/vite` (MF2) requires any change to the two
  other currently-Solid remotes' own `vite.config.ts` beyond the host's —
  needs a spike at the start of Phase 1, not assumed here.
- Whether Phase 6's decommission renames `frontend-v2/` back to `frontend/`
  or keeps the suffix permanently — deliberately deferred; low-impact,
  decide when Phase 5 is actually complete.
