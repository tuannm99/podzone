# ADR-0005: Frontend framework — Angular replaces SolidJS as the target, migrated incrementally

## Status
Proposed

## Date
2026-07-13

## Related Commit
(none yet — proposal only; `frontend-v2/` exists as an unwired spike, see Context)

## Context

The current frontend (`frontend/`) is SolidJS + Vite, split into a HOST
(`frontend/src`) and three Module-Federation remotes (`apps/backoffice`,
`apps/iam`, `apps/onboarding`) plus a shared design-system package
(`packages/shared`) — roughly 300 files, ~28,600 lines of TypeScript/TSX.

Two independent signals point at the same problem:

1. `docs/03-architecture-detail-design/13-frontend-solid-audit.md` records
   real, still-partially-open findings: effect-driven remote fetching
   without stale-response protection (P0), inconsistent
   `try/finally` around mutation submitting state, large owners mixing
   unrelated responsibilities, and permanent editor/list mixing across IAM,
   Partner, and Product Setup workflows. `agent/SOLID_STYLE_GUIDE.md`'s
   dedicated "MFE / Vite-Plugin-Federation — SolidJS Reactive System Split"
   section documents a subtle, Solid-fine-grained-reactivity-specific bug
   class (a remote's `createSignal` and `createRenderEffect` can resolve to
   two different module instances across a federation boundary, silently
   breaking reactivity) that has already required one documented fix and
   remains a footgun for every future remote.
2. User-stated direct pain point (2026-07-13): the SolidJS codebase is hard
   to maintain.

A spike this session (`frontend-v2/`) confirmed Angular 22 is viable
end-to-end: `ng new` scaffold, `@angular-architects/native-federation`
configured as a remote, `ng serve`/`ng build`/`ng test` all verified live.
It is **not** wired into the running host — the host's
`@originjs/vite-plugin-federation` speaks a different federation
protocol/manifest format than native-federation's `remoteEntry.json`
(confirmed by inspecting a live-served manifest, `$version: "v4"` vs the
host's Webpack-MF-v1-style format). Real interop requires the host to
migrate to `@module-federation/vite` (MF2) first — see PZEP-0004 for the
concrete sequencing.

## Decision

Angular becomes the target frontend framework, adopted **incrementally**,
not as a big-bang rewrite:

1. `frontend-v2/` (Angular) becomes the real app over time. `frontend/`
   (SolidJS) is not deleted now — it keeps serving every current route
   until each remote is migrated and cut over, per PZEP-0004's phased plan.
2. `agent/ANGULAR_STYLE_GUIDE.md` is the mandatory reference for any file
   under `frontend-v2/`. `agent/SOLID_STYLE_GUIDE.md` remains mandatory for
   `frontend/` until that tree is fully decommissioned — neither guide
   replaces the other while both trees are live; `CLAUDE.md`'s Frontend
   pointer is updated to reference both, scoped by path.
3. The design system (`packages/shared`'s Tailwind classes/tokens,
   `docs/03-architecture-detail-design/15-design-system.md`) is preserved
   as source of truth for visual language — it is framework-agnostic CSS,
   not Solid-specific — and gets re-implemented as Angular components
   incrementally, not thrown away and redesigned.

## Alternatives Considered

### Option A: Stay on SolidJS, fix the audit findings instead of migrating
Pros:
- No framework-migration risk, no dual-stack maintenance window.
- Audit's P0/P1/P2 findings are independently addressable without a rewrite.
- Preserves existing SolidJS-specific knowledge already encoded in
  `agent/SOLID_STYLE_GUIDE.md` and `packages/shared`.

Cons:
- Does not address the stated pain point if the difficulty is Solid's
  fine-grained-reactivity mental model itself, not just accumulated debt —
  fixing the audit's findings doesn't remove `createEffect`/reactive-prop/
  federation-singleton footguns from the language, only from today's code.
- Smaller ecosystem and hiring pool than Angular.
- The MFE reactive-system-split bug class is structural to
  Solid-JSX-compiled-static-imports + federation, and reappears with every
  new remote unless each one is configured correctly by hand (see
  `SOLID_STYLE_GUIDE.md`'s Rule Of Thumb) — there is no framework-level fix.

### Option B: Migrate to Angular (this decision)
Pros:
- Batteries-included framework (Router, Reactive Forms, DI, CLI, HttpClient,
  CDK a11y primitives) reduces the "assemble your own architecture"
  surface that produced several audit findings in the first place
  (ad hoc effect-driven fetching, inconsistent submitting-state handling).
- Angular 22 is zoneless by default (confirmed: `frontend-v2`'s scaffold has
  no `zone.js` dependency or `provideZoneChangeDetection` call) and uses
  Signals (`signal`/`computed`/`effect`/`resource`) as its reactivity
  primitive — directly analogous to Solid's `createSignal`/`createMemo`/
  `createEffect`/`createResource`, so this is more an idiom/tooling change
  than a full paradigm change for anyone who has worked in the current code.
- `@angular-architects/native-federation` ships in lockstep with Angular
  major versions (confirmed: both at 22.0.x today) — lower drift risk than
  a community-maintained single-bundler plugin.
- Largest ecosystem/hiring pool among realistic options.

Cons:
- Full rewrite of ~28,600 lines across host + 3 remotes + shared design
  system — large, multi-phase effort (see PZEP-0004).
- Dual-stack period: both `frontend/` and `frontend-v2/` must build, lint,
  and deploy correctly for the duration of the migration.
- The host must migrate its federation plugin
  (`@originjs/vite-plugin-federation` → `@module-federation/vite`, MF2)
  before any cross-framework remote loading works at all — a change to a
  currently-working system, and the single highest-risk step in the plan
  (see PZEP-0004 Phase 1).
- Team must apply Angular idioms (DI, Angular Router, Reactive Forms,
  RxJS where signals/`resource()` aren't enough) in place of existing
  Solid-specific experience.

### Option C: Rewrite in React instead of Angular
Pros:
- Largest ecosystem/hiring pool of any single option.

Cons:
- Not what was requested.
- Loses Angular's more opinionated batteries-included structure that
  directly targets the audit's root causes (ad hoc state management,
  inconsistent async/mutation handling patterns assembled per-feature) —
  React would require re-assembling that structure by convention, the same
  failure mode that produced the current SolidJS findings in the first
  place.

Chosen: **B**, per explicit user direction (2026-07-13).

## Consequences

- Two frontend stacks coexist for the duration of the migration. CI,
  `make` targets, and `CLAUDE.md`'s Frontend Commands section must be able
  to build/lint/test both independently — `frontend-v2/` must not be
  allowed to silently break `frontend/`'s pipeline or vice versa.
- The host's federation-plugin migration (MF2) is a prerequisite for the
  *first* Angular remote cutover, not just a nice-to-have — until it lands,
  `frontend-v2/` can only run standalone. This is called out explicitly so
  it is not rediscovered as a surprise mid-migration.
- `agent/SOLID_STYLE_GUIDE.md`'s MFE reactive-split section is Solid-specific
  and does not apply to Angular remotes. `agent/ANGULAR_STYLE_GUIDE.md`
  documents Angular/native-federation's own singleton rules (e.g.
  `@angular/core` must be a true singleton — already the scaffold's
  default via `shareAll` + an explicit override, confirmed in
  `frontend-v2/federation.config.mjs`) as a separate, non-transferable
  rule of thumb.
- No backend, contract, database, or event change of any kind — this
  decision and its follow-up PZEP are frontend-only.
- `frontend/` is not frozen during the migration — bug fixes and small
  feature work continue there per the normal sprint/task process until its
  corresponding remote is actually cut over; this ADR does not grant a
  blanket exemption from fixing production issues in SolidJS code while it
  is still live.

## Rule Of Thumb

A remote (or the host) migrates from SolidJS to Angular only when a
PZEP for that specific remote exists and is accepted — this ADR licenses
the overall direction and the framework choice, it does not itself
authorize touching any specific remote's code. See PZEP-0004 for the
phase sequence and which PZEP gates which remote.
