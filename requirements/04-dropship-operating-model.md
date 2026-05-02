# POD Operating Model

## Purpose

This document reframes Podzone specifically as a `POD-first operating platform`, not only as a generic multi-tenant commerce platform.

The key distinction is:

- a normal commerce backoffice mainly manages products and orders
- a POD backoffice must coordinate `merchant`, `print partner`, `product setup`, `order routing`, and `fulfillment visibility`

That difference should shape the product model, UI language, and roadmap.

## Business model assumptions

Podzone is assumed to support merchants who:

- sell products under their own storefront brand
- rely on external production or fulfillment partners
- need operational visibility across products, orders, shipments, and margin
- may later expand into broader sourced-product models such as dropship

This means the platform must help merchants manage `brand control without owning production capacity`.

## Primary POD actors

## 1. Merchant / Store Owner

This actor owns the storefront brand and customer relationship.

Responsibilities:

- choose which products to sell
- set retail pricing
- define merchandising strategy
- manage order exceptions
- monitor profit and fulfillment performance

## 2. Store Operations Team

This actor keeps the store running day to day.

Responsibilities:

- review product setup quality
- publish or unpublish products
- monitor incoming orders
- respond to delayed or failed fulfillment

## 3. Print Or Fulfillment Partner

This actor provides production and/or fulfillment capacity.

Responsibilities:

- receive product or order specifications
- confirm production or routing readiness
- accept routed orders
- fulfill shipments
- return fulfillment updates

In a mature product, the partner should not be just a note on a store. It should become a managed business party in the platform.

## 4. Platform Operator

This actor manages the network and platform governance.

Responsibilities:

- onboard merchants
- onboard partners
- define platform policies
- monitor system health and disputes
- control platform-level fees or access models

## 5. Finance / Settlement Operator

This actor handles revenue-quality questions.

Responsibilities:

- inspect merchant payout readiness
- inspect partner obligations
- review platform fee calculations
- handle chargeback or reconciliation issues

## 6. End Customer

This actor buys from the merchant storefront.

The end customer is not a backoffice user, but their lifecycle drives the operational workload.

## Core POD value chain

The platform should support this chain:

1. `Set up`
   Merchant creates a store, team, and operational foundation.
2. `Connect`
   Merchant connects a print or fulfillment partner.
3. `Prepare`
   Merchant prepares or selects products ready for production.
4. `Publish`
   Merchant publishes products to selling channels.
5. `Sell`
   Customer places order through merchant storefront.
6. `Route`
   Platform determines which partner should fulfill the order.
7. `Fulfill`
   Partner ships order and updates fulfillment state.
8. `Settle`
   Merchant, partner, and platform economics are reconciled.

If Podzone does not eventually support most of this chain, it is not yet a credible POD operations platform.

## Primary business objects

### Merchant Store

The commercial workspace where a merchant team operates.

### Print Partner

A production-side or fulfillment-side entity that helps fulfill customer demand.

### Product Template

The merchant-facing setup for something the store intends to sell.

### Product Listing

The published selling representation of a template/product.

### Customer Order

The commercial order received from the buyer.

### Fulfillment Order

The partner-facing order request derived from the customer order.

### Settlement Record

The financial record that tracks:

- retail revenue
- partner cost
- shipping cost
- platform fee
- merchant margin

## Critical POD business flows

## Flow 1. Partner onboarding

### Goal

Make a print or fulfillment partner available as a trusted execution partner.

### Steps

1. Platform or merchant creates partner record.
2. Contact and operating details are configured.
3. Partner becomes active for selected stores.

### BA note

This flow is missing today as a polished product feature and should become a first-class domain.

## Flow 2. Product setup and publishing

### Goal

Let the merchant decide what to sell and prepare it for fulfillment.

### Steps

1. Merchant reviews candidate products or templates.
2. Merchant selects items to activate.
3. Merchant edits title, media, and merchandising.
4. Merchant sets retail pricing and margin targets.
5. Merchant publishes items to storefront channels.

## Flow 3. Order intake and routing

### Goal

Turn a customer order into partner-fulfillable work.

### Steps

1. Customer places order.
2. Platform records store order.
3. Platform resolves partner mapping for each line.
4. Fulfillment order is created and sent to partner.
5. Merchant can track fulfillment state.

### BA note

This is the operational heart of the product.

## Flow 4. Fulfillment exception management

### Goal

Help merchant teams manage the messier parts of POD operations.

### Common exceptions

- delayed production
- delayed shipment
- partner rejects order
- partial fulfillment
- lost tracking visibility

## Flow 5. Settlement and margin visibility

### Goal

Let the merchant and platform understand whether each order is economically healthy.

### Steps

1. Order revenue is recorded.
2. Partner cost is associated.
3. Shipping cost and fees are applied.
4. Margin is calculated.
5. Settlement status is shown.

## Product wording implications

If the product is truly POD-oriented, UI and docs should use language like:

- `Print partners`
- `Production`
- `Product setup`
- `Publish to store`
- `Fulfillment status`
- `Shipment tracking`
- `Margin`
- `Payout`

## What still matters about dropship

Dropship should remain a future-compatible expansion path.

That means the platform should still be designed so that:

- partner models can later include `dropship supplier`
- product sourcing can later include external catalog import
- order routing can later branch between POD partners and dropship suppliers

## What is still missing for a credible POD context

- partner as first-class domain in the product surface
- product setup workflow that reflects production reality
- order routing to execution partner
- fulfillment tracking
- operational dashboard for store teams
- margin visibility

## BA conclusion

`A seller operations platform for multi-store POD businesses, where merchants manage team access, product setup, order routing, fulfillment visibility, and settlement from one backoffice.`
