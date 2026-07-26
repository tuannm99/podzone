# PZEP-0014: Angular routed-orders collection screen (second ADR-0008 screen)

## Status
Done — implemented, `ng build`/`format`/`format:check`/`ng test` pass, and
verified serving from docker dev. Not click-through verified in a browser
(standing gap, same as PZEP-0011/0012/0013).

## Date
2026-07-25

## Related Commit
(none yet — implemented uncommitted in the working tree this session)

## Requirement Sources
- Business: PZEP-0012 built the pattern (`CollectionFieldSpec<T>`,
  `CollectionFilters`/`CollectionToolbar`) and one screen (stores); this
  applies it to a second, store-scoped screen to prove it generalizes.
- Feature: reuses the existing `routedOrders(collection: CollectionInput):
  RoutedOrderPage!` GraphQL query
  (`internal/backoffice/controller/graphql/schema/routing.graphqls`),
  backed by the `routing` repository's `sqlspec.Spec` (RQ2, PZEP-0010).
- Use Cases: an operator can search/filter/sort/page through a store's
  routed orders from `frontend-v2`.
- Acceptance Criteria: see below.

## Summary

Second real screen on the ADR-0008 pattern:
`/t/:tenantId/backoffice/orders`, store-scoped (unlike `stores`, which is
tenant-only), reusing `routedOrders`.

## Problem / Scope Note — `catalog` was considered and rejected for this slot

Before picking `routing`, `catalog`'s `productSetupSnapshot` query was
evaluated as the "next" collection screen (natural sequence from
PZEP-0012). It does not qualify: `productSetupSnapshot: ProductSetupSnapshot!`
takes no `collection: CollectionInput` argument at all — it returns the
full unpaginated drafts+candidates arrays, with no server-side search/
filter/sort. Building a `CollectionFieldSpec`-driven UI against it would
mean either (a) faking pagination/filter/sort client-side over a
fully-fetched array — a bug per `agent/ANGULAR_STYLE_GUIDE.md`'s explicit
"client-side `.slice()`/`.filter()` on operational data is a bug" rule, or
(b) inventing GraphQL arguments the backend doesn't have — forbidden by
CLAUDE.md's "do not invent API DTOs... outside the documented contract."

**Flagging this as a real backend gap, not fixing it here**: catalog needs
its own PZEP to add `CollectionInput` support to `productSetupSnapshot` (or
split it into two paginated queries, `productSetupDrafts(collection)`/
`productSetupCandidates(collection)`) before a real collection screen can
be built for it. `routing`'s `routedOrders` already has this support end to
end (Go `routedOrderSpec` + GraphQL `CollectionInput` argument), so it was
picked instead — no backend change needed.

## Goals

- `features/backoffice/routed-order-list-page/`: `routed-order-list.types.ts`
  (`routedOrderFieldSpec`, matching the backend's `routedOrderCollectionColumns`
  keys exactly — see `internal/backoffice/infrastructure/repository/routing/
  collection_query.go`), `routed-order-list.service.ts` (URL-query-param-driven,
  same shape as `StoreListService`), `routed-order-list-page.component.*`.
- New route `/t/:tenantId/backoffice/orders`, nested under `Shell` (alongside
  `backoffice`/`backoffice/stores` from PZEP-0013), gated by `authGuard`.
- `Shell.platformNavItems` gains an "Orders" entry.
- `BackofficeGraphqlService.query()` called with both `tenantId` and
  `storeId` (routing is store-scoped, unlike `stores`) — `storeId` resolved
  the same way `TenantHomePage` already does: query param first, falling
  back to `StoreStorageService`.

## Non-Goals

- `queueView` preset filter and `priority` custom sort — these are
  backend-specific hooks (`sqlspec.Spec.CustomFilters`/`CustomSorts`) with
  no equivalent concept in `CollectionFieldSpec<T>` yet. Only the 15 plain
  columns `routedOrderCollectionColumns` declares are exposed here. Adding
  queue-view/custom-sort support to the frontend pattern is a real,
  separate follow-up if a screen needs it — not invented speculatively.
- Order detail/edit/mutation actions (shipment, settlement, issue-handling
  updates) — this screen is read-only, matching PZEP-0012's store screen.
- The `catalog` backend gap noted above — flagged, not fixed.

## Proposed Solution

```text
frontend-v2/src/app/
  features/backoffice/routed-order-list-page/
    routed-order-list.types.ts     (new: routedOrderFieldSpec)
    routed-order-list.service.ts   (new)
    routed-order-list-page.component.*  (new)
  app.routes.ts                    (add backoffice/orders under Shell)
  shell/shell.component.ts         (add Orders nav item)
```

Field spec (matches backend exactly):

| Field | Kind | Filterable | Sortable |
|---|---|---|---|
| id, productTitle, partner, customerName, status, exceptionType, exceptionStatus, shipmentStatus, operatorAssignee, settlementStatus, realizedMargin | text | yes | (productTitle, status, updatedAt sortable in the UI's sort dropdown) |
| createdAt, updatedAt, shipmentSlaDueAt, issueSlaDueAt | timestamp | yes | yes |

## Affected Components
- Frontend: `frontend-v2/src/app/features/backoffice/routed-order-list-page/`,
  `app.routes.ts`, `shell/shell.component.ts`
- Backend/Gateway/Worker/Database/External Integration: none — reuses the
  existing `routedOrders` query unchanged.

## API / DB / Event / Permission Contract Changes
- None.

## Security Considerations
- Same `X-Tenant-ID`/`X-Store-ID`-scoped authorization the backend already
  enforces; frontend adds no check of its own.

## Test Plan
- Build/format/test: `ng build`, `prettier --check`, `ng test`.
- Manual QA: docker dev route resolution only (HTTP 200) — no browser
  click-through this session, same standing gap.

## Agent Implementation Plan
- TASK-0001: `routed-order-list.types.ts` (field spec).
- TASK-0002: `routed-order-list.service.ts`.
- TASK-0003: `routed-order-list-page.component.*` + route + nav item.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: field spec matches backend's `routedOrderCollectionColumns` exactly, no invented fields | TASK-0001 | Code review against `collection_query.go` |
| AC-2: `contains`/`starts_with` unavailable on the four timestamp fields in the filter UI | TASK-0001 | Code review (`kind: 'timestamp'`, `operatorsFor`) |
| AC-3: screen is store-scoped (sends `X-Store-ID`) | TASK-0002 | Code review |
| AC-4: `/t/:tenantId/backoffice/orders` resolves under `Shell` | TASK-0003 | docker dev route check |

## Open Questions
- `catalog`'s missing `CollectionInput` support — needs its own future
  PZEP (backend + frontend) before a catalog collection screen exists.
- `queueView`/`priority` support in `CollectionFieldSpec` — worth adding
  once a second backend entity needs a custom filter/sort hook (routing
  already has one that this screen doesn't expose yet).
