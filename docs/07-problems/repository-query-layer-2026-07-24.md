# Repository And Query Layer — Known Problems

Identified from code review and measurement on 2026-07-24.
Scope: `internal/backoffice/infrastructure/repository/{store,catalog,routing}/`,
`pkg/collection/`.

Continues the 2026-07-23 audit series:
[Multi-tenancy issues](./multitenancy-2026-07-23.md) (tenant isolation) and
[Platform scaling and runtime topology](./platform-scaling-2026-07-23.md)
(clusters, capacity, k8s). This note covers the repository/query layer:
maintainability of the SQL construction code, and one real defect found while
reviewing it.

Status as of 2026-07-25: **RQ1, RQ2, RQ3, and RQ4 step 1 and step 2 are all
resolved** — see [PZEP-0010](../09-pzep/PZEP-0010-backoffice-declarative-collection-spec.md)
(RQ1/RQ2: `pkg/collection/sqlspec.Spec`, applied to `store`/`routing`) and
the inline notes on each item below (RQ3/RQ4: `internal/backoffice/
infrastructure/repository/sqlrow`). `catalog`'s three duplicated `Select`
column lists and `store`'s two were migrated to `sqlrow.SelectFrom`;
`routing`'s more complex joined/aggregate `Select` shapes were deliberately
left as-is (not a clean fit for the same helper — see RQ3's note). Unlike
the other two 07-problems notes, parts of this one are **measured rather
than reasoned** — benchmarks and equivalence tests were actually run, and
are marked as such per item.

---

## Performance Of The Query Builder — Measured, Not A Problem

Recorded here so the question does not get re-litigated: **Squirrel is not a
bottleneck and should not be replaced for performance reasons.**

Benchmarked on the actual query shapes in this repo (AMD Ryzen 7 PRO 5850U,
Go 1.26.5, `-benchtime=200000x`):

| Operation | ns/op | allocs/op |
|---|---|---|
| Squirrel build, `ListCandidates` shape (15 cols, 1 WHERE, 1 ORDER BY) | 10 711 | 102 |
| Equivalent raw `const` string + arg slice | 32 | 1 |
| Squirrel build, 4 dynamic filters + limit (`collection_query` shape) | 14 422 | 127 |

The ratio (≈330×) is misleading; the absolute number is what matters.
**10.7 µs** against a Postgres round trip of roughly 0.5–2 ms means query
building is **0.5–2 % of a single round trip**.

For comparison, on the same request path:

- MT8 (multi-tenancy note) — every repository call adds a
  `SELECT current_database(), current_schema()` round trip, roughly **100×
  more expensive** than all query building in that call.
- MT14 — one gRPC `CheckPermission` per resolved field, ~1–5 ms each.
- MT2 — the whole service shares a 10-connection pool.
- PS9 — CPU limit of `50m` guarantees CFS throttling.

Caveat worth recording: 102 allocs/op does add GC pressure, and at a `50m` CPU
limit everything is relatively worse. But it would take **>10k query builds per
second** for this to register, and the 10-connection pool caps throughput far
below that. Revisit only if MT2, MT8 and PS9 are all fixed and profiling then
points here.

**Conclusion:** any change to the query-construction code in this note is
justified by maintainability alone. Do not present it as a performance fix.

---

## RQ1 — `routing` permits `contains` / `starts_with` on `TIMESTAMPTZ` columns, producing a runtime SQL error

**Resolved as of 2026-07-25** by RQ2's `sqlspec.Spec` migration — `routing`
now applies the same `KindText` guard `store` always had. Confirmed with a
new test, `TestBuildRoutedOrderCollectionQueryRejectsContainsOnTimestampColumn`.

**File:** `internal/backoffice/infrastructure/repository/routing/collection_query.go:139-173,11-27`,
`internal/backoffice/migrations/sql/0003_create_routed_orders.sql:13-14`
**Severity:** Medium — user-triggerable 500 on a valid-looking filter
**Not reproduced against a live database** — derived from the code and the
migration's column types.

`store` guards text-only operators with a `text` flag on the column
(`store/collection_query.go:85-93`):

```go
case collection.FilterContains:
    if !column.text {
        return nil, unsupportedStoreFilter(filter)
    }
```

`routing` has **no equivalent check** — it applies the `LIKE` wrapper to any
mapped column:

```go
case collection.FilterContains:
    return likeRoutedOrderColumn(column, "%"+escapeRoutedOrderLike(filter.Values[0])+"%"), nil
```

But `routedOrderCollectionColumns` maps four timestamp fields — `createdAt`,
`updatedAt`, `shipmentSlaDueAt`, `issueSlaDueAt` — and migration `0003`
declares `created_at TIMESTAMPTZ NOT NULL` / `updated_at TIMESTAMPTZ NOT NULL`.
So a filter of `createdAt contains "2026"` generates:

```sql
LOWER(created_at) LIKE LOWER($1) ESCAPE '\'
```

which Postgres rejects with `function lower(timestamp with time zone) does not
exist`. The client sees a 500 rather than a validation error.

**Fix:** apply the same kind/type check `store` already has. RQ2's shared spec
does this structurally — see the note there about the behaviour change.

---

## RQ2 — `collection_query.go` is duplicated between `store` and `routing`

**Resolved as of 2026-07-25** — see [PZEP-0010](../09-pzep/PZEP-0010-backoffice-declarative-collection-spec.md).
Both files are now ~15-30 line `sqlspec.Spec` declarations; the pre-existing
`collection_query_test.go` in each package passes unmodified against the
new implementation.

**File:** `internal/backoffice/infrastructure/repository/store/collection_query.go` (124 lines),
`internal/backoffice/infrastructure/repository/routing/collection_query.go` (181 lines)
**Severity:** Medium — every new collection costs ~124 lines of copy-paste, and
fixes land in one copy only

The two files are near-identical with different identifier prefixes:

| `store` | `routing` | Relationship |
|---|---|---|
| `storeFilter` | `routedOrderFilterClause` | same 9-case operator switch |
| `escapeStoreLike` | `escapeRoutedOrderLike` | byte-identical |
| `storeLike` | `likeRoutedOrderColumn` | byte-identical |
| `buildStoreCollectionQuery` | `buildRoutedOrderCollectionQuery` | same shape: search → filter loop → sort → direction |

RQ1 is a direct consequence: the divergence between the two copies *is* the
bug. Only one copy got the type guard.

**Fix — a declarative spec compiled once in `pkg/collection`.** Each model then
declares fields instead of writing the switch:

```go
var storeSpec = sqlspec.Spec{
	Entity:      "store",
	DefaultSort: "createdAt",
	SearchOrder: []string{"id", "name", "description", "status"},
	Fields: map[string]sqlspec.Field{
		"id":          {Column: "id", Kind: sqlspec.KindText, Searchable: true},
		"name":        {Column: "name", Kind: sqlspec.KindText, Searchable: true},
		"description": {Column: "description", Kind: sqlspec.KindText, Searchable: true},
		"ownerId":     {Column: "owner_id", Kind: sqlspec.KindText},
		"status":      {Column: "status", Kind: sqlspec.KindText, Searchable: true},
		"createdAt":   {Column: "created_at"},
		"updatedAt":   {Column: "updated_at"},
	},
}
```

124 lines → ~15. Call site becomes `compiled, err := storeSpec.Compile(query)`,
returning predicates and `ORDER BY`.

Non-column cases stay out of the generic path via explicit hooks, so
`routing`'s specifics survive unchanged:

- `CustomFilters["queueView"]` — reuses the existing
  `routedOrderQueueViewClause` verbatim; the convention `(nil, nil)` means
  "match everything, add no predicate", preserving today's `all` / `my_queue`
  behaviour.
- `CustomSorts["priority"]` — holds the existing `CASE … END` expression.
- `Base` — always-applied predicates such as `sq.Eq{"store_id": storeID}`.

This removes the `if filter.Field == "queueView" { … continue }` branch from
the filter loop while keeping the logic in its own function.

**Evidence this is behaviour-preserving:** a prototype was built and tested
against the current `store` implementation copied verbatim — **14 query cases
(empty, search with `%`/`_` escaping, all 9 operators, both sort directions,
combined search+filter+sort) produce byte-identical SQL and arguments**, plus
error parity on 4 rejection cases (unknown filter field, missing value,
text-operator-on-non-text, unknown sort field). Prototype is in the session
scratchpad, not committed.

**Behaviour change to declare:** migrating `routing` onto the shared spec turns
RQ1 from a Postgres 500 into an `ErrInvalidQuery` rejection. That is the
correct behaviour but it is a visible API change — it belongs in a changelog
entry, and the frontend should be checked for any filter UI that currently
offers `contains` on a date field.

**Explicitly not proposed:** deriving the spec from the row struct's `db` tags.
The external filter key (`ownerId`) and the storage column (`owner_id`) are
deliberately separate concerns — coupling them means a column rename silently
breaks the client-facing filter contract. Keep the mapping explicit.

---

## RQ3 — Column lists are hand-written and duplicated per query, and must be kept in sync with `db` tags manually

**Partially resolved as of 2026-07-25.** `internal/backoffice/
infrastructure/repository/sqlrow.ColumnsOf[T]`/`SelectFrom[Row]` implemented
and applied to `catalog` (5 sites: 2× `productSetupDraftRow`, 3×
`productSetupCandidateRow`) and `store` (2 sites). `routing`'s `Select`
call sites were left as hand-written — they mix joins, `FromSelect`, and
`COUNT(*)` aggregates that don't reduce to "select every column of one
struct," and forcing them into this helper would need per-call column
subsetting the helper doesn't (and shouldn't) support.

**File:** `internal/backoffice/infrastructure/repository/catalog/repository.go:63-77,101-115,165-180,205-220`
(and equivalents in `store` / `routing`)
**Severity:** Medium — silent drift between struct and query

The 15-column list for `product_setup_candidates` is written out **four times
in one file**, and must independently match the `db` tags on
`productSetupCandidateRow`. Adding a column means editing five places; missing
one produces either a scan error or a silently absent field. Approximately
**205 lines** across the three repository files sit inside `Select(...)` /
`Columns(...)` argument lists (measured with an `awk` line-range heuristic —
treat as an order of magnitude, not exact).

**Fix:** derive the list from the row struct once.

```go
func columnsOf[T any]() []string                        // reads db tags, cached by reflect.Type
func selectFrom[Row any](table string) sq.SelectBuilder // psql.Select(columnsOf[Row]()...).From(table)
```

**Measured:** the derived list is identical to the hand-written one for
`productSetupCandidateRow` (15 columns), and costs **26.6 ns/op with 0
allocs** once cached — irrelevant next to the 10.7 µs build and the millisecond
round trip.

Unlike RQ2 this is purely internal; there is no contract change.

---

## RQ4 — Every repository method repeats the same `ToSql` → error → `withTenantTx` → `ensureXTables` sandwich

**Step 1 resolved as of 2026-07-25**, bundled with MT8 as planned: all 20
`ensureXTables(ctx, tx)` call sites and the 3 identical wrapper functions
are gone from `store`/`catalog`/`routing`'s `repository.go`. Migrations now
run once per tenant scope via `pdtenantdb.Config.OnSchemaBootstrap` (see
MT8's resolution note in `multitenancy-2026-07-23.md`), not on every
repository call.
**Step 2 partially resolved**: `sqlrow.QueryMany[Row]`/`QueryOne[Row]`
implemented and applied to `catalog`'s plain read methods (`ListDrafts`,
`ListCandidates`, `GetCandidateByID`, `GetCandidateByDraftID` — the ones
whose not-found semantics are a plain `(nil, nil)`). `GetDraftByID` was left
untouched — it returns a custom "not found" error instead, a different
contract `QueryOne` doesn't (and shouldn't) special-case. `store` and
`routing`'s remaining call sites were not converted — lower priority once
step 1 already removed most of the per-call overhead this item targeted.

**File:** `internal/backoffice/infrastructure/repository/{store,catalog,routing}/repository.go`
**Severity:** Low (mechanical) — but it is most of the bulk of these files

Measured across the three repository files (~1 600 lines total):

| Pattern | Occurrences | ≈ Lines |
|---|---|---|
| `ToSql()` + `if err != nil` | 26 | 104 |
| `ensure*Tables(ctx, tx)` guard | 20 | 60 |
| `withTenantTx(ctx, func(tx *sqlx.Tx) error {…})` wrapper | 20 | 80 |

**Fix, in dependency order:**

1. **Delete the `ensure*Tables` calls.** This is already required by MT8
   (migrations must come off the request path). It removes ~60 lines *and* a
   database round trip per repository call — the best ratio available here, and
   it is not optional work.
2. **Generic transaction helpers** collapsing the remaining sandwich:
   ```go
   func queryMany[Row any](ctx, r txRunner, tenantID string, b sq.SelectBuilder) ([]Row, error)
   func queryOne[Row any](ctx, r txRunner, tenantID string, b sq.SelectBuilder) (*Row, error)
   ```

Combined with RQ3, `catalog.ListDrafts` goes from **38 lines to ~14** (the
row → domain mapping loop stays). A prototype of the helpers plus two rewritten
methods **compiles against the real `sqlx` / `squirrel` types**; it was not
wired into the repo.

Estimated reduction across the three files is 30–40 %. That figure is an
extrapolation from two converted methods, not a completed migration.

---

## Suggested Order Of Work

1. **RQ4 step 1** — remove `ensure*Tables`. Do this as part of MT8, not
   separately: splitting them means editing the same 20 call sites twice.
2. **RQ3** — `columnsOf[T]` / `selectFrom[Row]`. Internal only, no contract
   change, lands independently.
3. **RQ4 step 2** — generic `queryMany` / `queryOne` helpers.
4. **RQ2 on `store` first** — port the equivalence test alongside it as the
   safety net; `store` has no custom filters or sorts, so it exercises the
   generic path cleanly.
5. **RQ2 on `routing`** — carries the custom hooks and closes **RQ1**. Needs a
   changelog entry and a frontend check for the behaviour change.

RQ2 changes an outward-facing error contract, so it needs its own sprint slice
or PZEP before implementation. RQ3 and RQ4 are internal refactors and can be
scoped as ordinary slices — but RQ4 is entangled with MT8 and should be
sequenced with it.

---

## RQ1-Class Audit Of Other Services (2026-07-25)

PZEP-0010's Open Question asked whether `auth`/`iam`/`onboarding`/`partner`
have the same duplicated-collection-query shape and are worth migrating onto
`sqlspec.Spec`. Audited all five `collection_query.go` files that exist
outside `backoffice`. Result: **only `partner` had the actual bug; the other
four had already solved it independently, before this audit.**

| Service | Storage | Text/kind guard on `contains`/`starts_with`? | Duplication already collapsed? |
|---|---|---|---|
| `auth` | Postgres/Squirrel | Yes — explicit per-field **per-operator allowlist** (`map[FilterOperator]struct{}`), stricter than `sqlspec.Spec`'s binary text/non-text `Kind` | Yes — one generic `buildCollectionQuery(query, spec)` already |
| `iam` | Postgres/Squirrel | Yes — same `{sql string; text bool}` shape `sqlspec.Field` uses | Yes — one generic `buildIAMCollectionQuery` feeds 13 different column maps |
| `onboarding/store` | MongoDB | Yes — `requestFieldKind` (string/objectID/time), guards `contains`/`starts_with` to string only | Yes — one generic path per collection |
| `onboarding/infrasmanager` | MongoDB | Yes — `collectionFieldKind` (string/int/bool/time), same guard | Yes — one generic `buildInfrastructureCollection` feeds 5 column maps |
| `partner` | Postgres/Squirrel | **No** — `partnerFilterClause` applied `LIKE` to any column, no kind check | No — single hand-written switch, not generic |

**`partner` had a real RQ1-class bug**, confirmed against its own migrations
(`internal/partner/migrations/sql/0004_add_partner_capabilities.sql`):
`slaDays`/`routingPriority` are `INTEGER`, `createdAt`/`updatedAt` are
`TIMESTAMPTZ`. A `contains` or `starts_with` filter on any of these built
`LOWER(sla_days) LIKE ...` etc., which Postgres rejects — the exact RQ1
shape, just in a different service.

**Resolved as of 2026-07-25**: `internal/partner/infrastructure/repository/collection_query.go`
migrated onto `sqlspec.Spec` (same pattern as `store`/`routing`), closing the
bug and collapsing the single-service duplication. The pre-existing
`buildPartnerCollectionQuery` tests pass unmodified against the new
implementation; two new tests confirm `contains`/`starts_with` on
`slaDays`/`createdAt` are now rejected as `collection.ErrInvalidQuery`, and
one confirms the `ListPartnersQuery.Status`/`PartnerType` fixed-filter
behavior (`sqlspec.Spec.Compile`'s `base` predicates) is unchanged.

**`auth` and `iam` are explicitly not migrated onto `sqlspec.Spec`** — they
already have equivalent (or, for `auth`, strictly better) protection, and
migrating them would be pure churn with no bug fixed and no duplication
reduced net of the churn itself. `onboarding`'s two files are MongoDB-based;
`sqlspec.Spec` is Squirrel/Postgres-specific and does not apply there at all
— a Mongo-flavored equivalent would be a separate, new piece of work, not
something to retrofit speculatively without a concrete duplication problem
to point at (`onboarding`'s two generic builders already avoid the
duplication `sqlspec.Spec` targets).
