# PZEP-0010: Backoffice repository layer — declarative collection spec (closes RQ1)

## Status
Done

## Date
2026-07-25

## Related Commit
(none yet — implemented uncommitted in the working tree this session)

## Requirement Sources
- Business: recovery/maintainability track following the 2026-07-24
  repository-and-query-layer audit — see `docs/07-problems/
  repository-query-layer-2026-07-24.md` RQ1 and RQ2.
- Feature: none new — internal refactor plus one real bug fix.
- Use Cases: a `routing` collection query that sends `contains`/
  `starts_with` on a timestamp column gets a normal validation error
  instead of a Postgres 500.
- Functional Requirements: no SRS behavior change beyond the bug fix itself
  (RQ1) — filter/sort semantics for `store` and `routing` collections are
  otherwise unchanged, proven by the pre-existing `collection_query_test.go`
  files passing unmodified against the new implementation.
- Non-functional Requirements: `agent/SOLID_STYLE_GUIDE.md`/
  `agent/ANGULAR_STYLE_GUIDE.md`'s collection-state requirements are
  unaffected (backend-only change).
- Acceptance Criteria: see below.
- UI Specs: none.

## Summary

`internal/backoffice/infrastructure/repository/{store,routing}/collection_query.go`
were ~90% duplicated hand-written filter/sort builders (RQ2). The one place
they diverged was a real bug: `routing` had no text-column guard on
`contains`/`starts_with`, so filtering a TIMESTAMPTZ column like `createdAt`
built `LOWER(created_at) LIKE ...` and Postgres rejected it as a runtime 500
(RQ1). Both files are now thin declarations of a shared, generic
`pkg/collection/sqlspec.Spec` compiler, which applies the text-guard
uniformly - closing RQ1 by construction, not by a one-off patch.

## Problem

See RQ1/RQ2 in `docs/07-problems/repository-query-layer-2026-07-24.md` for
the full writeup, including the measured benchmark establishing that this
change is justified by maintainability, not performance (Squirrel query
building is ~0.5-2% of a request's round trip).

## Goals

- One reusable `sqlspec.Spec` compiler in `pkg/collection/sqlspec`,
  covering: plain column filters (all 9 operators), the text-only guard on
  `contains`/`starts_with`, free-text search across a declared column
  subset, default/explicit sort, and two escape hatches for logic that
  isn't a plain column comparison - `CustomFilters` (routing's `queueView`
  preset clauses) and `CustomSorts` (routing's `priority` CASE expression).
- `store`/`routing`'s `collection_query.go` files become field
  declarations (`var storeSpec = sqlspec.Spec{...}`) instead of hand-written
  switches - `buildStoreCollectionQuery`/`buildRoutedOrderCollectionQuery`
  keep their exact prior signatures, so `repository.go` call sites in both
  packages needed zero changes.
- RQ1 closes as a side effect: `routing`'s spec marks the same columns
  `KindText` that `store` already did, so `contains`/`starts_with` on a
  timestamp field now returns `collection.ErrInvalidQuery` before any SQL
  is built.

## Non-Goals

- Migrating `catalog`'s collection queries (catalog doesn't have a
  `collection_query.go` - out of scope, not audited).
- Migrating `auth`/`iam`/`onboarding`/`partner`'s own `collection_query.go`
  files. A repo-wide grep during this work found the same duplicated shape
  repeated in all four of those services - each is a legitimate future
  candidate for this same `sqlspec.Spec`, but none were audited as part of
  RQ1/RQ2 and migrating them is out of scope here. Flagged as a follow-up,
  not attempted.
- Deriving a spec from struct tags or backend-to-frontend codegen - RQ2
  explicitly rejected coupling the external filter key (e.g. `ownerId`) to
  the storage column name (`owner_id`); the mapping stays an explicit,
  hand-written `Fields` map per entity.

## Proposed Solution

`pkg/collection/sqlspec.Spec` (new package):

```go
type Spec struct {
	Entity        string
	DefaultSort   string
	SearchOrder   []string
	Fields        map[string]Field
	CustomFilters map[string]func(collection.Filter) (sq.Sqlizer, error)
	CustomSorts   map[string]string
}
func (s Spec) Compile(query collection.Query, base ...sq.Sqlizer) (collection.Query, []sq.Sqlizer, string, error)
```

`base` is a `Compile`-time argument, not a `Spec` field, because it's
per-call data (e.g. `routing`'s `sq.Eq{"store_id": storeID}`), not part of
the entity's reusable shape.

`store/collection_query.go` and `routing/collection_query.go` each declare
one `Spec` value and a one-line wrapper function preserving the original
signature. `routing`'s `routedOrderQueueViewClause` function is reused
verbatim as a `CustomFilters["queueView"]` entry; its `priority` CASE
expression becomes a `CustomSorts["priority"]` entry.

## Affected Components
- Backend: `pkg/collection/sqlspec` (new), `internal/backoffice/
  infrastructure/repository/{store,routing}/collection_query.go`
- Gateway/Frontend/Worker/Database/External Integration: none

## API Contract Changes

**Yes, one, narrow, and intentional:** a `routing` collection query using
`contains` or `starts_with` on a timestamp field (`createdAt`, `updatedAt`,
`shipmentSlaDueAt`, `issueSlaDueAt`) now returns a validation error
(`collection.ErrInvalidQuery`, surfaced as whatever 4xx-class response the
GraphQL/HTTP layer maps that to) instead of a Postgres-error-derived 500.
This is strictly a bug fix (the request never had a well-defined result),
but it is a response-shape change for a client that was, improbably, relying
on the 500. No other filter/sort/search behavior changes for `store` or
`routing` - proven by the pre-existing `collection_query_test.go` test
files in both packages passing unmodified against the new implementation.

**Frontend check required:** confirm no `routing`/orders collection filter
UI currently offers `contains`/`starts_with` on a date field. Not verified
as part of this PZEP (frontend-v2 has no `routing` collection screen yet -
see `docs/08-adr/ADR-0008-frontend-v2-declarative-collection-field-spec.md`,
which specifically designs the future Angular field spec to prevent this
class of mismatch at the UI layer).

## DB / Event Contract Changes
- None.

## Permission Changes
- None.

## Security Considerations
- None beyond the bug fix itself (a validation error is a smaller attack/
  error surface than a raw DB-error-derived 500).

## Observability
- No new requirement. The error is already routed through
  `collection.ErrInvalidQuery`, the same sentinel every other invalid-query
  case in both repositories already uses.

## Test Plan
- Unit: `pkg/collection/sqlspec` - new tests covering search+filter
  composition, default vs. explicit sort, the text-kind guard, unknown
  field/sort rejection, base-predicate prepending, and both escape hatches
  (`CustomFilters` nil-clause-means-no-predicate, `CustomSorts`).
- Regression: `store/collection_query_test.go` and
  `routing/collection_query_test.go` - both pre-existing files pass
  unmodified against the new implementation (predicate counts and
  `ORDER BY` strings unchanged).
- New: `routing/collection_query_test.go` gained four cases - RQ1's fix
  (`contains` on `createdAt` now rejected), `queueView=overdue` (predicate
  count), `queueView=all` (no predicate added), and `priority` sort
  (CASE expression present, direction applied).
- All of the above run green; `go build ./...`, `gofmt -l`, and
  `golangci-lint` are clean for every touched package.

## Agent Implementation Plan
- TASK-0001: `pkg/collection/sqlspec.Spec` + `Compile` + unit tests. Done.
- TASK-0002: `store/collection_query.go` migrated; existing tests verified
  unmodified-passing. Done.
- TASK-0003: `routing/collection_query.go` migrated (`CustomFilters`/
  `CustomSorts`), existing tests verified unmodified-passing, four new RQ1/
  custom-hook tests added. Done.

## Acceptance Criteria Mapping

| AC | Task | Test |
|---|---|---|
| AC-1: `store`'s existing filter/sort/search behavior is unchanged | TASK-0002 | `store/collection_query_test.go` (unmodified, passing) |
| AC-2: `routing`'s existing filter/sort/search/queueView/priority behavior is unchanged | TASK-0003 | `routing/collection_query_test.go` (unmodified, passing) |
| AC-3: RQ1 closes - `contains`/`starts_with` on a timestamp column is rejected before any SQL runs | TASK-0003 | New test `TestBuildRoutedOrderCollectionQueryRejectsContainsOnTimestampColumn` |
| AC-4: `sqlspec.Spec` is independently correct, not just "happens to work" for these two entities | TASK-0001 | `pkg/collection/sqlspec` unit tests |

## Open Questions
- Resolved 2026-07-25: audited `auth`/`iam`/`onboarding`/`partner`. Only
  `partner` had the same duplicated shape *and* the RQ1-class bug (`auth`,
  `iam`, and `onboarding`'s two Mongo-backed builders had already solved
  it independently). `partner` migrated onto `sqlspec.Spec`; the other
  three left as-is. See `docs/07-problems/repository-query-layer-2026-07-24.md`'s
  "RQ1-Class Audit Of Other Services" section for the full breakdown.
