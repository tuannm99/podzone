# PZEP-0006: Angular base MFE scaffold — shell, onboarding feature area, design system port

## Status
Done — implemented and verified 2026-07-13/14 (build/serve/test level). Not yet committed.

## Date
2026-07-13

## Related Commit
(uncommitted — working tree only, pending user review)

## Requirement Sources
- Business: user directive 2026-07-13 — before writing onboarding business
  logic, get the Angular base MFE running: a shell layout and an onboarding
  feature area that actually builds and serves, plus confirm whether
  Flowbite can be used in Angular (it currently backs part of `frontend/`'s
  Tailwind setup).
- Feature: [PZEP-0004](./PZEP-0004-solidjs-to-angular-frontend-migration.md)
  Phase 2 (pilot remote: onboarding) — this PZEP is Phase 2's first slice
  (infrastructure/scaffold), not the full pilot.
- Use Cases: an agent can open `frontend-v2/src/app/features/onboarding/`
  and start writing real onboarding pages against a working shell + routing
  + design tokens, without first having to invent app structure.
- Functional Requirements: none new — no business behavior implemented.
- Non-functional Requirements: none new.
- Acceptance Criteria: see below.
- UI Specs: `docs/03-architecture-detail-design/15-design-system.md` (tokens
  ported verbatim, not redesigned).

## Summary

Three things, scoped tightly per the user's explicit sequencing
("shell + onboarding chạy được để code trước, nghiệp vụ sau"):

1. **Flowbite compatibility check** — `flowbite-angular` (the Angular
   component wrapper) does not support Angular 22 yet (latest release
   `21.0.0`, peer-capped `@angular/core >=21.0.0 <22.0.0`). Not usable.
   Vanilla `flowbite` (the Tailwind CSS plugin + tokens, no framework
   dependency) works fine and was wired in the same way `frontend/` uses
   it — CSS/utility classes only, no Flowbite JS behavior (interactive
   components are hand-rolled per `agent/ANGULAR_STYLE_GUIDE.md`'s existing
   CDK-a11y guidance, unchanged by this finding).
2. **Design tokens ported verbatim** from `frontend/src/solid/global.css`
   into `frontend-v2/src/styles.css` — same `@theme` color tokens, same
   dark-mode overrides, same base layer. The Toast UI Editor-specific CSS
   in the Solid file was not ported (no editor component exists in
   `frontend-v2` yet; port it when that component is actually built).
3. **Base app structure**: `src/app/shell/` and
   `src/app/features/onboarding/` (one stub page, `AdminHomePage`), wired
   through `app.routes.ts`. No business logic, no API calls yet — those are
   explicitly deferred to the next PZEP that actually specs onboarding's
   redesigned flow.
4. **Shell layout matches the real product chrome exactly**, not a rough
   approximation. Ported verbatim (same Tailwind classes, same structure)
   from `frontend/src/modules/shell/PodzoneNavbar.tsx` + `Root.tsx`'s
   `<main class="pb-8 lg:pl-64">` composition — fixed `w-64` left sidebar
   (brand mark, workspace-switch row, Platform nav section, account row),
   sticky header with `lg:pl-64` offset, `max-w-[96rem]` content container.
   Cross-checked against the earlier `root-vs-org-nav-mockup` Artifact
   (`https://claude.ai/code/artifact/49ddd60f-97b3-404d-b1fd-28e6b7b1ed3b`),
   which itself was built to mirror `PodzoneNavbar.tsx` — used the real
   component as the source of truth since it's the actual current
   implementation, not the standalone mockup's own hand-rolled CSS classes.
   Workspace/account data is static placeholder ("Choose a store" / "Not
   signed in") since no auth/workspace context exists in `frontend-v2` yet
   — wiring that is business logic, deferred per Non-Goals.

## Problem

Phase 0's scaffold (`frontend-v2`) had no app structure beyond Angular
CLI's default boilerplate (placeholder marketing page, no routing, no
design tokens, no feature boundaries) — not enough to start writing real
onboarding code against `agent/ANGULAR_STYLE_GUIDE.md`'s stated boundaries
(`src/app/features/<feature>/`, `src/app/core/`, `src/app/shared/`).

## Goals
- A working shell layout and one routed feature page, verified live
  (served, not just built).
- Design tokens match `frontend/`'s exactly — no visual drift between the
  two stacks during the migration.
- A clear, evidence-based answer on Flowbite for Angular, so no one
  spends time later trying to install an incompatible package.

## Non-Goals
- Any onboarding business logic, API integration, or real page content —
  next PZEP.
- Redesigning the onboarding/multitenant flow itself (workspace UX,
  approval queue, lifecycle states) — that redesign needs its own SRS/UI
  spec update and PZEP before code, per `docs/00-governance/agent-working-rule.md`.
  This PZEP is infrastructure only.
- Wiring `frontend-v2` into the real host — still standalone, per PZEP-0004
  Phase 2's own sequencing (cutover happens once the pilot remote's real
  pages exist and are verified).
- Backend contract changes — none needed yet; this PZEP touches
  `frontend-v2` only.
- Porting the Toast UI Editor CSS or any other component-specific styling
  not yet used by any real `frontend-v2` component.

## Proposed Solution
See Summary above — already implemented in full; this section intentionally
does not repeat it.

## Affected Components
- Frontend: `frontend-v2/src/styles.css` (new), `frontend-v2/.postcssrc.json`
  (new), `frontend-v2/src/app/shell/*` (new), `frontend-v2/src/app/features/onboarding/*`
  (new), `frontend-v2/src/app/app.ts`/`app.routes.ts`/`app.spec.ts` (rewired),
  `frontend-v2/package.json` (new deps: `tailwindcss`, `@tailwindcss/postcss`, `flowbite`)
- Everything else: none.

## API / DB / Event / Permission Contract Changes
- None.

## Data Ownership / Security / Observability
- Unchanged — no runtime behavior beyond static UI.

## Alternatives Considered

### Alternative: Use `flowbite-angular`
Rejected — genuinely not possible right now, not a preference call: its
peer dependency range excludes Angular 22 entirely (`<22.0.0`). Revisit
when/if a `22.x` release ships.

### Alternative: Hand-roll new design tokens instead of porting
Rejected — the whole point of a design system is one visual source of
truth; forking values here would let the two stacks drift silently.

## Test Plan
- Build: `ng build` — clean (2 benign CSS "empty sub-selector" warnings,
  not errors; not investigated further, cosmetic).
- Serve: `ng serve --port 3004` — confirmed via curl that the served
  bootstrap chunk contains the real sidebar markup ("Podzone", "Seller
  Center", "Choose a store") and the onboarding stub page's "Admin Home"
  text — i.e., the real routed component tree is what's being served, not
  stale/cached boilerplate. Re-verified after the layout was rebuilt to
  match `PodzoneNavbar.tsx` exactly.
- Test: `ng test --watch=false` — 1/1 passing (adjusted `app.spec.ts` to
  drop the removed placeholder-title assertion, added no new tests since
  there's no logic yet to test).
- E2E: none — no browser tool available, and no interactive behavior exists
  yet to test.

## Agent Implementation Plan
- TASK-0001: Research Flowbite Angular compatibility; document finding.
- TASK-0002: Install `tailwindcss`/`@tailwindcss/postcss`/`flowbite`, add
  `.postcssrc.json`, port design tokens into `styles.css`.
- TASK-0003: Build `Shell` layout component + `AdminHomePage` stub +
  routing wiring.
- TASK-0004: Verify per Test Plan.

All 4 tasks done.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: Flowbite-for-Angular question answered with evidence, not a guess | TASK-0001 | npm registry inspection (peer deps, dist-tags) |
| AC-2: `frontend-v2` builds clean with Tailwind+Flowbite tokens active | TASK-0002 | `ng build`, output CSS size confirms tokens compiled (70+ kB vs near-zero before) |
| AC-3: Shell + onboarding stub actually serve, not just compile | TASK-0003 | `ng serve` + curl content check on the real served chunk |
| AC-4: Existing test suite still passes | TASK-0003 | `ng test --watch=false` |

## Follow-up (2026-07-13, same session): shared UI primitives

User feedback: the shell template's inline Tailwind utility strings were
too heavy to write/read directly (same utility-first style `frontend/`
already uses, but `frontend-v2` had no component layer hiding it yet).
Fix: ported `frontend/packages/shared/ui/components/common/Primitives.tsx`'s
`Button`/`Card`/`Badge`/`Spinner` to `frontend-v2/src/app/shared/ui/` —
same class strings, same color/size maps, verbatim — plus a new `NavLink`
component (no Solid equivalent; extracted from `PodzoneNavbar.tsx`'s inline
`navLinkClass()`/`navTagClass()` helpers since the shell needs it directly).
`classes()`/`isExternalUrl()` also ported verbatim into `shared/utils.ts`.

`Shell` and `AdminHomePage` were rewritten to consume `app-nav-link` and
`app-card` instead of inline classes. Rebuilt, re-served, and re-confirmed
via curl against the live bootstrap chunk that the component-rendered
output (not just a clean compile) still contains the same real content.
`ng test --watch=false` still 1/1 passing.

Note: `Button`'s internal-vs-external link branching required an
Angular-specific choice `Primitives.tsx` doesn't have to make — Solid's
`Link` component picks `<a routerLink>` vs `<a href target=_blank>` at
runtime inside one component; Angular's `[routerLink]` and a plain `[href]`
are different directives on different template branches, so `Button`'s
template has two `@if` branches (internal / external) instead of Solid's
one. Behavior is equivalent, structure isn't identical — worth knowing if
someone diffs the two files expecting a 1:1 line match.

## Open Questions
- Whether the 2 "empty sub-selector" CSS warnings need investigation before
  more components are added — deferred, cosmetic, not blocking.
- Real onboarding business logic and the workspace/multitenant flow
  redesign are explicitly out of scope here — next PZEP must define that
  scope (which lifecycle states, approval queue UI, workspace UX) before
  any of it is coded, per the user's own request to plan before building.
