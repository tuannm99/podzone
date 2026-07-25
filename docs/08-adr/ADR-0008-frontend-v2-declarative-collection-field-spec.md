# ADR-0008: `frontend-v2` collection screens declare a typed per-entity field spec, not per-page filter/sort literals

## Status
Accepted

## Date
2026-07-25

## Related Commit
(none yet — decision only; no collection screens exist in `frontend-v2` yet)

## Context

`agent/ANGULAR_STYLE_GUIDE.md` (same rule as the SolidJS style guide)
requires every collection page to provide search/filter/sort/loading/
error/empty/pagination states. `frontend-v2/src/app/shared/ui` already has
a generic `DataTable` component (columns declared once, rows passed in,
loading/error/empty built in) — but no equivalent generic components for
the filter/sort/toolbar part of a collection page yet, because no
Backoffice collection screen (store, catalog, routing management) has been
built in `frontend-v2` yet.

A pattern review of an unrelated, locally kept legacy multi-module
reference codebase (not a podzone dependency, used only to mine
architecture patterns before this codebase is discarded) surfaced a
recurring shape worth naming: at the scale of dozens of CRUD modules,
hand-writing a filter-fields array and a sort-options array per collection
page produces two costs:

1. The same field metadata (name, label, allowed filter operators, is it
   sortable) is duplicated per page instead of declared once per entity.
2. Nothing structurally prevents offering a filter operator that doesn't
   make sense for a field's type — e.g. a text operator like "contains" on
   a timestamp column. This is not hypothetical: it is the exact defect
   already logged as RQ1 in
   `docs/07-problems/repository-query-layer-2026-07-24.md` (podzone's own
   `routing` repository accepts `contains`/`starts_with` on a `TIMESTAMPTZ`
   column and only discovers the mismatch at the database, as a runtime
   500). The same duplicated-literal shape independently exists today in
   podzone's Go repository layer (RQ2/RQ3 in the same note) and in the
   SolidJS frontend's per-page filter arrays (e.g.
   `frontend/apps/iam/src/pages/admin-iam/groups/GroupsCollection.tsx`
   hand-writes its filter-field/operator list inline).

Since no `frontend-v2` collection screen exists yet, this is a choice made
at the point of creation, not a retrofit.

## Decision

Every collection screen in `frontend-v2` declares its queryable fields
once, as a typed field spec:

```ts
type CollectionFieldKind = 'text' | 'number' | 'boolean' | 'timestamp' | 'enum'

type CollectionFieldSpec<T> = {
  field: keyof T & string
  label: string
  kind: CollectionFieldKind
  filterable?: boolean          // default false
  sortable?: boolean            // default false
  // allowed operators are derived from `kind` by the shared component,
  // not repeated per field — e.g. `timestamp` never offers `contains`.
}
```

This spec is a plain exported constant next to the page's feature service
(not a new architectural layer), consumed by shared `CollectionFilters` /
`CollectionToolbar` components — parallel to how `DataTable` already
consumes a `DataTableColumn<T>[]` — instead of each page hand-writing an
inline filter/sort options array.

## Alternatives Considered

### Option A: Keep hand-writing filter/sort arrays per page
Pros:
- No new shared component to build; matches the one filter UI that exists
  today in the SolidJS app.

Cons:
- Duplicated literals per page, growing linearly with every new collection
  screen.
- No structural guard against offering an invalid operator for a field's
  type — reproduces the RQ1 defect class in a new framework instead of
  closing it.

### Option B (chosen): Typed per-entity field spec + generic filter/toolbar components
Pros:
- Field metadata declared once; operator validity follows from `kind`
  automatically.
- Closes off the RQ1 defect class on the frontend side before it can be
  introduced, complementing (not replacing) the backend's own validation.
- Matches the same "declare once, derive many" direction already proposed
  for the backend repository layer (RQ2's `sqlspec.Spec` idea) — same
  instinct, applied at the UI layer.

Cons:
- One more shared abstraction to design and maintain (`CollectionFilters`,
  `CollectionToolbar`) before the first real collection screen ships.

### Option C: Fully generic, config-driven collection page renderer (the whole page — including the table — rendered from one spec, no page-specific markup)
Pros:
- Least code per page of any option; this is the most aggressive version
  of the pattern the reference codebase actually used.

Cons:
- Becomes a black-box interpreter: rendering decisions live inside the
  shared engine instead of the page, which conflicts directly with
  `agent/ANGULAR_STYLE_GUIDE.md`'s rule that components render their own
  `input()`/`output()`s and invoke explicit actions. The reference
  codebase's own version of this (a ~1800-line generic render engine with
  implicit parent/child context binding) was flagged in review as the part
  *not* worth reproducing — only the narrower "declare field metadata
  once" instinct is. Rejected as too much magic for too little benefit
  over Option B.

Chosen: **B**.

## Consequences

- `frontend-v2/src/app/shared/ui` needs new `CollectionFilters` and
  `CollectionToolbar` components before the first Backoffice collection
  screen ships; `DataTable` stays as-is (already generic, already correct
  shape).
- Every future collection-screen PZEP declares a `CollectionFieldSpec<T>`
  next to its feature service instead of inlining a filter-fields array.
- This ADR decides the *pattern*, not the screens themselves — building
  any specific Backoffice collection screen (store, catalog, routing
  management) still requires its own PZEP per the usual rule (CLAUDE.md:
  do not expand feature breadth before the recovery backbone works; no
  PZEP, no cross-component implementation).
- Frontend-side operator restriction is a UX/correctness improvement, not
  a security or validation boundary — the backend remains the authority
  and must still reject invalid operator/field combinations itself
  (unaffected by this ADR; see RQ1/RQ2 in the repository-query-layer note
  for the backend-side fix).

## Rule Of Thumb

A new `frontend-v2` collection page declares a `CollectionFieldSpec<T>`
next to its feature service and passes it to the shared filter/toolbar/
table components — it does not inline a filter-fields array or a
sort-options array by hand.
