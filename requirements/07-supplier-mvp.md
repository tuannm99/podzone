# Supplier Domain MVP

This document explains how the current `supplier` implementation should be interpreted in a `POD-first` product strategy.

## Product framing

At the requirements level, Podzone should speak about:

- `print partner`
- `production partner`
- `fulfillment partner`

The current `supplier` domain in code should be treated as a technical placeholder for a broader future `partner` model.

The implementation has now taken the first step in that direction by introducing `partner_type` in the current supplier model.

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

The code currently uses `supplier` naming. That is acceptable temporarily for implementation momentum, but the product language should remain `partner` or `print partner` until the strategic direction broadens again.

Current implementation note:

- `partner_type` defaults to `print_on_demand`
- additional values currently supported are `fulfillment` and `dropship_supplier`
