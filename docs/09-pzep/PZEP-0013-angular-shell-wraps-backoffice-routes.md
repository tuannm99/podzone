# PZEP-0013: Angular `Shell` wraps Backoffice routes; Stores nav link

## Status
Done — implemented, `ng build`/`format`/`format:check`/`ng test` pass.
Not click-through verified in a browser (same standing gap as PZEP-0011/
0012).

## Date
2026-07-25

## Related Commit
(none yet — implemented uncommitted in the working tree this session)

## Requirement Sources
- Business: PZEP-0012 shipped a second Backoffice screen
  (`/t/:tenantId/backoffice/stores`) reachable only by typing the URL —
  no navigation links it to the Backoffice home page.
- Feature: `Shell` (`frontend-v2/src/app/shell/`, from PZEP-0006) was
  scaffolded with exactly this intent — its own code comment says
  "Shell wraps real app content once frontend-v2 hosts screens past the
  handoff" — but was never actually mounted into any route.
- Use Cases: a user on the Backoffice home page can navigate to Stores (and
  back) via a sidebar link, not just a typed URL.
- Non-functional Requirements: `agent/ANGULAR_STYLE_GUIDE.md` (internal
  navigation via `routerLink`).
- Acceptance Criteria: see below.
- UI Specs: none new — `Shell`'s existing layout/nav-list components,
  unchanged visually.

## Summary

Mounts `Shell` as the parent route for `/t/:tenantId/backoffice` and
`/t/:tenantId/backoffice/stores` (Angular child routes), and makes its nav
items tenant-scoped instead of the static PZEP-0006 placeholder. This is
the point in the app the Shell's own design comment already named as the
right time to do this — Backoffice content now exists past the handoff.

## Problem

Two real screens exist under `/t/:tenantId/backoffice*` with no shared
layout or way to move between them except a typed URL. `Shell` exists,
`RouterOutlet`-ready, but sits completely unused.

## Goals

- New parent route: `{ path: 't/:tenantId', component: Shell, canActivate: [authGuard], children: [...] }`
  with `TenantHomePage` at `backoffice` and `StoreListPage` at
  `backoffice/stores`. The existing flat `t/:tenantId/stores` (store
  *chooser*, a different, pre-handoff screen — deliberately bare, no Shell)
  is untouched and stays a sibling top-level route, not a Shell child.
- `Shell.platformNavItems` becomes a `computed()` derived from the route's
  `tenantId` param instead of a static array: "Home" → `/t/:tenantId/backoffice`,
  "Stores" → `/t/:tenantId/backoffice/stores`.

## Non-Goals

- Fixing `Shell`'s workspace-switcher chip/header title (still the static
  "Choose a store" / "PZ" placeholder from PZEP-0006) — flagged there
  already as its own follow-up, not this PZEP's job.
- Any other Backoffice route joining `Shell` — only the two that exist
  today.
- Route guards beyond the existing `authGuard` (no store-scope-specific
  check — pre-existing gap, not introduced or fixed here).

## Proposed Solution

```ts
// app.routes.ts
{
  path: 't/:tenantId',
  component: Shell,
  canActivate: [authGuard],
  children: [
    { path: 'backoffice', component: TenantHomePage },
    { path: 'backoffice/stores', component: StoreListPage },
  ],
},
```

`Shell` injects `ActivatedRoute`, derives `tenantId` via
`toSignal(route.paramMap)`, and computes nav items from it.

## Affected Components
- Frontend: `frontend-v2/src/app/app.routes.ts`, `frontend-v2/src/app/shell/shell.component.ts`
- Backend/Gateway/Worker/Database/External Integration: none.

## API / DB / Event / Permission Contract Changes
- None.

## Test Plan
- Build/format/test: `ng build`, `prettier --check`, `ng test`.
- Manual QA: not done in a real browser this session (no browser tool
  available) — same gap as PZEP-0011/0012, not resolved here either.

## Agent Implementation Plan
- TASK-0001: nest `backoffice`/`backoffice/stores` under a new `Shell`
  parent route.
- TASK-0002: `Shell.platformNavItems` becomes tenant-scoped.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: `/t/:tenantId/backoffice` and `/t/:tenantId/backoffice/stores` render inside `Shell`'s layout | TASK-0001 | Code review + route resolution check |
| AC-2: `Shell`'s nav links point at the correct tenant-scoped URLs | TASK-0002 | Code review |
| AC-3: the pre-handoff `t/:tenantId/stores` (store chooser) is unaffected | TASK-0001 | Code review (route left untouched) |

## Open Questions
- None blocking. Real UX verification (does clicking the nav link actually
  work, does the layout look right) needs a browser session — flagged, not
  done.
