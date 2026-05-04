# Domain Map

## Overview

The repository already contains several important product and platform domains.

They should be viewed as business capability areas, not only as code modules.

## Domain 1. Identity And Access

### Purpose

Manage authentication, sessions, store access, roles, permissions, and auditability.

### Current modules

- `internal/auth`
- `internal/iam`

### Current maturity

- technically strong for MVP
- product copy still too technical

## Domain 2. Store Workspace / Backoffice

### Purpose

Provide the operational portal for store owners and their teams.

### Current modules

- `internal/backoffice`
- `internal/ui-podzone`

### Current capabilities

- authenticated store-scoped access
- store-related GraphQL shell
- admin/settings pages for IAM-backed operations

### Gap

This should become the main product surface for POD operations, but today it still reads too much like an admin console.

## Domain 3. Tenant Data Isolation

### Purpose

Ensure each store's operational data is isolated and routed correctly.

### Current modules

- `pkg/pdtenantdb`
- `internal/onboarding`

### Product interpretation

This is critical infrastructure, but should remain mostly invisible in the merchant-facing product.

## Domain 4. Store Infrastructure Onboarding

### Purpose

Manage the internal provisioning and placement metadata of store infrastructure.

### Current modules

- `internal/onboarding`

## Domain 5. Catalog

### Purpose

Manage products and product data.

### Current modules

- `cmd/catalog`
- related API contracts

### Future POD relevance

This domain will be central for:

- product setup
- listing management
- merchandising
- publishing

## Domain 6. Orders

### Purpose

Track and manage customer orders.

### Future POD relevance

This domain becomes the center of store operations once catalog and storefront workflows are alive.

It must eventually include fulfillment routing and execution visibility.

## Domain 7. Payments And Settlement

### Purpose

Track payment state, payout state, and eventual settlement readiness.

## Domain 8. Partner Management

### Purpose

Represent print, production, or fulfillment partners used by the store.

### Current status

- an early but active implementation slice now exists in code through the `partner` abstraction
- product language should continue to prefer `partner` or `print partner`

### Future requirements

- partner profile
- contact and operating details
- service capability
- fulfillment contract
- SLA visibility
- future extensibility toward dropship suppliers

## Domain 9. Fulfillment Operations

### Purpose

Manage the execution side of customer orders.

### Future requirements

- order routing
- shipment tracking
- partial fulfillment handling
- fulfillment exceptions
- SLA breach visibility

## Domain 10. Margin And Settlement

### Purpose

Help merchants and platform operators understand commercial health per order and per store.

### Future requirements

- cost aggregation
- fee model
- margin calculation
- payout readiness
- reconciliation views

## Domain relationship view

The platform should be understood as:

1. `Identity and Access`
   The user can safely enter and operate.
2. `Store Workspace`
   The user sees the operational surface of one store.
3. `Catalog`
   The store manages what it intends to sell.
4. `Orders`
   The store manages what customers bought.
5. `Partners`
   The store coordinates how orders get produced or fulfilled.
6. `Fulfillment`
   Orders are routed and tracked through execution.
7. `Payments and Settlement`
   The store understands money flow and platform economics.

## Current maturity assessment

### Strongest areas today

- authn/authz foundation
- session lifecycle
- multi-tenant routing
- admin governance model

### Weakest business areas today

- seller-facing narrative
- partner model
- product and order journeys
- POD-specific workflows

## Priority requirement gaps

### Priority 1. Reframe the UI around store operations

### Priority 2. Add a visible commerce home/dashboard

### Priority 3. Introduce partner and product onboarding flows

Needed because:

- POD value is not visible without product setup and execution readiness

### Priority 4. Design order lifecycle for merchant operations

### Priority 5. Add fulfillment and margin visibility

Needed because:

- POD operations break down quickly when the merchant cannot see execution quality and commercial viability

## BA recommendation

The current codebase should be treated as:

- `platform foundation complete enough for a serious MVP`
- `product story not yet expressed clearly enough for a POD audience`

The best next work should align user-facing language and flows around:

- store creation
- team onboarding
- partner connection
- product setup
- first order operations
- fulfillment visibility
