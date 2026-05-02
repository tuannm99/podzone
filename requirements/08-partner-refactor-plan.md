# Partner Refactor Plan

## Purpose

This document describes the completed move from the old `supplier` implementation to the current `partner` model that fits the `POD-first` product direction.

## Why this refactor matters

The current code uses `supplier` because it was introduced while the product narrative still leaned toward dropship.

That creates two problems:

1. the product language now says `print partner` or `partner`
2. the implementation language still says `supplier`

That gap has now been removed at the runtime and product API layers.

## Target model

The long-term business concept should be:

- `partner`

With future partner types such as:

- `print_on_demand`
- `fulfillment`
- `dropship_supplier`

## Recommended migration sequence

### Phase 1. Keep runtime stable, change product language first

At the start of the migration, the safe move was:

- say `partner`
- prefer `print partner` in POD flows

Status:

- completed

### Phase 2. Introduce type-aware domain model

Add to the code model:

- `partner_type`

Suggested values:

- `print_on_demand`
- `fulfillment`
- `dropship_supplier`

Outcome:

- current `supplier` implementation becomes flexible enough for POD-first language and future dropship extension

### Phase 3. Rename API surface

Move to:

- `PartnerService`
- `/partner/v1/partners`

Recommended migration style:

- add `PartnerService`
- migrate callers to the new route
- delete the old route once the workspace is clean

Status:

- completed
- UI now calls `/partner/v1/partners`
- runtime now serves `PartnerService` as the primary product-facing API

### Phase 4. Rename storage and module boundaries

Status:

- runtime source has already moved to `internal/partner`
- dev entrypoint has moved to `cmd/partner`
- table rename migration is now in place
- `partners` is now the target storage name for the active runtime
- legacy supplier runtime files and docker wiring have been removed

## Implementation recommendation

The next safe engineering step is:

1. keep `partner_type` aligned with POD defaults such as `print_on_demand`
2. build product workflows on top of `partner` rather than reopening `supplier`
3. update docs and runtime naming where it improves clarity
4. only keep backward-compatibility layers if a real caller still needs them

## BA conclusion

The correct move was:

- `POD-first language first`
- `type-aware partner model next`
- `full supplier -> partner rename before deeper product workflows`

That rename is now complete enough for current experimental use.
