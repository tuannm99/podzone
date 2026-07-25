# PZEP-0012: Angular store collection screen — first ADR-0008 declarative-spec screen

## Status
Done — implemented, `ng build`/`format`/`format:check`/`ng test` all pass,
and verified serving cleanly from the docker dev `frontend-v2` container
(route resolves, no build errors). **Not click-through verified in a
browser** — no browser automation available this session; only build/serve/
route-resolution checks were run. Treat as code-complete, not UX-verified.

## Date
2026-07-25

## Related Commit
(none yet — implemented uncommitted in the working tree this session)

## Requirement Sources
- Business: exercises [ADR-0008](../08-adr/ADR-0008-frontend-v2-declarative-collection-field-spec.md)'s
  pattern on a real screen, per that ADR's own note that it "decides the
  pattern, not the screens themselves."
- Feature: `docs/03-architecture-detail-design/11-iam-platform.md`
  (store entity), reuses the existing `stores` GraphQL query documented in
  `internal/backoffice/controller/graphql/schema/store.graphqls`.
- Use Cases: an operator can search, filter, sort, and page through a
  tenant's stores from `frontend-v2`.
- Functional Requirements: `agent/ANGULAR_STYLE_GUIDE.md`'s Collections
  section (search/filter/sort/loading/error/empty/pagination states;
  collection state lives in URL query params, not a plain signal).
- Acceptance Criteria: see below.
- UI Specs: none new — reuses Material components already in
  `frontend-v2/shared/ui`.

## Summary

Builds the shared `CollectionFilters`/`CollectionToolbar` components ADR-0008
named, plus the `CollectionFieldSpec<T>`/`operatorsFor` helper, and applies
them to the first real Backoffice collection screen: a tenant-scoped store
list at `/t/:tenantId/backoffice/stores`. This both delivers a real screen
and proves the pattern before any other collection screen (catalog, routing)
reuses it.

## Problem

`frontend-v2/shared/ui` had `app-data-table` (table + pagination, already
generic) but nothing for the filter/sort/search half of a collection screen.
Every future Backoffice collection screen would otherwise re-invent this,
repeating the exact duplication problem RQ2 already closed on the backend.

## Goals

- `shared/services/collection-field-spec.ts`: `CollectionFieldKind`
  ('text'|'number'|'boolean'|'timestamp'|'enum'), `CollectionFieldSpec<T>`,
  and `operatorsFor(kind)` — operators are derived from `kind`, never
  hand-listed per field, so a filterable field can't be given an operator
  its kind doesn't support (closes the RQ1 defect class at the UI layer,
  before a request is ever sent).
- `shared/services/collection-types.ts` gains the GraphQL-facing service
  logic previously left unported (`normalizePageInfo`, `emptyPageInfo`,
  `toGraphQLCollectionInput`, `WirePageInfo`) — ported verbatim from
  `frontend/packages/shared/services/collection.ts`.
- `shared/ui/collection-toolbar/` (search + sort + direction — page size is
  `app-data-table`'s `mat-paginator`'s job, not duplicated here) and
  `shared/ui/collection-filters/` (filter builder + active-filter chips),
  both ports of the equivalent SolidJS components, adapted to read operators
  from `operatorsFor(field.kind)` instead of a per-field `operators` array.
- `features/backoffice/store-list-page/`: `StoreListService` (query state
  derived from `ActivatedRoute.queryParamMap` via `toSignal`, updates via
  `Router.navigate(..., { queryParamsHandling: 'merge' })` — **not** a plain
  `signal()`, per `agent/ANGULAR_STYLE_GUIDE.md`'s explicit rule) +
  `StoreListPage` (thin composition root) + `store-list.types.ts` declaring
  `storeFieldSpec`, matching `internal/backoffice/infrastructure/repository/
  store/collection_query.go`'s `storeSpec` field-for-field.
- `BackofficeGraphqlService.query()`'s `storeId` scope parameter becomes
  optional — the `stores` query is tenant-scoped only (no `X-Store-ID`),
  matching `frontend/`'s `includeStoreHeader: false` option on the
  equivalent SolidJS call.
- New route `/t/:tenantId/backoffice/stores`, gated by the existing
  `authGuard`.

## Non-Goals

- Store create/activate/deactivate mutations — this PZEP is read-only
  (list/search/filter/sort/page). The existing onboarding store-creation
  flow (`/t/:tenantId/stores`) is unaffected and unrelated.
- Any other Backoffice collection screen (catalog, routing/orders,
  partners) — each is separately scoped, building on this same foundation.
- A `CollectionFieldSpec`-driven screen wired into `Shell`'s nav — this
  route exists and is reachable by URL; adding a sidebar link is a small
  follow-up, not required for this PZEP's goal of proving the pattern.

## Proposed Solution

```text
frontend-v2/src/app/
  shared/services/
    collection-field-spec.ts          (new)
    collection-types.ts               (extended: GraphQL helpers)
  shared/ui/
    collection-toolbar/               (new)
    collection-filters/                (new)
  features/backoffice/
    backoffice-graphql.service.ts     (storeId now optional)
    store-list-page/
      store-list.types.ts              (new: storeFieldSpec)
      store-list.service.ts            (new)
      store-list-page.component.*      (new)
  app.routes.ts                        (add /t/:tenantId/backoffice/stores)
```

## Affected Components
- Frontend: `frontend-v2/src/app/{shared,features/backoffice,app.routes.ts}`
- Gateway/Backend/Worker/Database/External Integration: none — reuses the
  existing `stores` GraphQL query unchanged.

## API / DB / Event Contract Changes
- None.

## Permission Changes
- None new. Same GraphQL-level authorization the `stores` query already
  enforces; not verified end-to-end this session (no browser to log in
  with and confirm a real user's permission set renders the list
  correctly) — flagged as an open risk, see Test Plan.

## Security Considerations
- Unchanged from existing `stores` query authorization. Frontend
  operator-restriction (ADR-0008) is UX/correctness, not a security
  boundary — the backend's `sqlspec.Spec` remains the actual enforcement
  point, per `docs/00-governance/agent-working-rule.md`.

## Test Plan
- Unit: none added (matches existing policy — Angular unit tests are
  optional, same as `frontend/`).
- Build/format/test: `ng build`, `prettier --check`, `ng test` all pass.
- Manual QA: confirmed the docker dev `frontend-v2` container serves the
  new route (HTTP 200, no federation/TS build errors) after a container
  restart cleared a stale `.angular/cache` build-cache entry (the cache is
  a container-only volume, isolated from the host's own `.angular/cache`
  used by local `ng build` runs — see `deployments/docker/services.yml`'s
  comment on that volume for why). **Did not** log in and click through
  the screen in an actual browser — no browser automation tool available
  this session. This is a real gap: the screen's real-data behavior
  (whether `storeFieldSpec` actually matches live GraphQL response shapes,
  whether IAM permissions render an empty/error state correctly) is
  unverified beyond "it builds and the route resolves."

## Agent Implementation Plan
- TASK-0001: `collection-field-spec.ts` + `collection-types.ts` GraphQL
  helpers. Done.
- TASK-0002: `CollectionToolbar` + `CollectionFilters` components. Done.
- TASK-0003: `BackofficeGraphqlService` optional `storeId`. Done.
- TASK-0004: `store-list-page` (service + component) + route. Done.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: a filterable field only offers operators valid for its `kind` | TASK-0001/0002 | Code review (`operatorsFor`) |
| AC-2: collection query state (page/search/sort/filters) lives in the URL, not a plain signal | TASK-0004 | Code review (`StoreListService.query` derives from `queryParamMap`) |
| AC-3: `stores` list loads without `X-Store-ID` | TASK-0003 | Code review + docker dev route-resolution check |
| AC-4: screen renders real data correctly for an authenticated user | none | **Not verified** — needs a real browser click-through, flagged as follow-up |

## Open Questions
- Whether `/t/:tenantId/backoffice/stores` is reachable from `Shell`'s nav
  yet, or only by direct URL — deferred; add a nav link when the next
  collection screen lands and the Backoffice nav shape is worth settling.
- A real browser click-through against live Docker dev data should happen
  before this is trusted as more than "compiles and serves" — flagged, not
  done here.
