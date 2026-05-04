# Partner Domain MVP

This document explains how the current `partner` implementation should be interpreted in a `POD-first` product strategy.

## Product framing

At the requirements level, Podzone should speak about:

- `print partner`
- `production partner`
- `fulfillment partner`

The old `supplier` domain should be treated as historical implementation language for the broader `partner` model now used by the active runtime.

The current implementation already exposes `partner_type` in the active partner model.

## Objective

Introduce a store-scoped external business partner so Podzone can move from `store access platform` to `POD operations platform`.

This MVP does **not** attempt to solve:

- partner product feeds
- product setup workflow
- fulfillment routing
- shipment tracking
- settlement

It creates the minimum business object that those later flows can attach to.

## MVP scope

The first release should support:

1. create a partner record for a store
2. list partner records in a store
3. inspect one partner record
4. update business/contact details
5. activate or deactivate a partner

## BA note

The product language should remain `partner` or `print partner`.

Historical `supplier` wording may still appear in older migrations, backlog notes, or earlier requirement drafts, but it should not be treated as the preferred current terminology for the active runtime.

Current implementation note:

- `partner_type` defaults to `print_on_demand`
- additional values currently supported are `fulfillment` and `dropship_supplier`
