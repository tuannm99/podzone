# Epics And User Stories

## Purpose

This document translates the `POD-first` requirements into a backlog starter.

It is intentionally written at product level, not implementation level.

## Prioritization model

The recommended delivery sequence is:

1. store foundation
2. team access
3. partner onboarding
4. product setup and publishing
5. order and fulfillment operations
6. margin and settlement visibility

## Epic 1. Store Creation And Store Access

### Outcome

A merchant can create a store, enter it safely, and understand which store is currently active.

### User stories

- As a merchant, I want to create my first store so I can start operating my business.
- As a merchant, I want to belong to multiple stores so I can manage more than one brand or operation.
- As a merchant, I want to switch between stores safely so I do not accidentally operate in the wrong store.
- As a merchant, I want the system to clearly show my current store context so I can trust where my actions apply.

## Epic 2. Team Access And Store Roles

### Outcome

A store owner can manage access for store staff without platform support.

### User stories

- As a store owner, I want to invite teammates by email so I can onboard my team quickly.
- As a store owner, I want to assign store roles so each teammate gets the right level of access.
- As a store owner, I want to revoke store access so I can remove users who should no longer operate in my store.
- As a store admin, I want to see who has access to the store so I can review operational risk.

## Epic 3. Partner Onboarding

### Outcome

A merchant or platform operator can register print or fulfillment partners as real business entities in the system.

### User stories

- As a merchant, I want to add a print partner so I can fulfill orders reliably.
- As a platform operator, I want to onboard partners in a controlled way so the network remains trustworthy.
- As a merchant, I want to store partner contact and operating details so order execution can be automated later.
- As a merchant, I want to activate or deactivate partners per store so I can control which execution paths my store uses.

## Epic 4. Product Setup And Publishing

### Outcome

A merchant can prepare products for sale and publish them into store operations.

### User stories

- As a merchant, I want to prepare products or templates so I can launch catalog quickly.
- As a store operator, I want to validate product details before they go live unchecked.
- As a merchant, I want to edit titles, descriptions, and media so listings fit my brand.
- As a merchant, I want to set retail prices so I can manage margin and positioning.

## Epic 5. Order Intake And Routing

### Outcome

Customer orders become partner-executable work with clear routing logic.

### User stories

- As a merchant, I want orders to appear in backoffice quickly so my team can monitor operations.
- As a store operator, I want each order routed to the right partner so fulfillment can start without manual lookup.
- As a merchant, I want fulfillment orders created from customer orders so partner-side execution becomes traceable.
- As a merchant, I want routing logic to be visible so I understand why a partner was selected.

## Epic 6. Fulfillment Tracking And Exceptions

### Outcome

The merchant team can monitor execution and respond to failures.

### User stories

- As a store operator, I want to see fulfillment status per order so I can track what is late or blocked.
- As a store operator, I want to be alerted when a partner cannot fulfill an order so I can take action fast.
- As a merchant, I want an exception queue so my team can focus on orders needing intervention.
- As a merchant, I want to know which partner is causing repeated delays so I can protect store performance.

## Epic 7. Margin And Settlement Visibility

### Outcome

The merchant can understand whether each order and store is commercially healthy.

### User stories

- As a merchant, I want to see estimated margin per product so I can choose what to sell.
- As a merchant, I want to see realized margin per order so I can detect pricing or fulfillment problems.
- As a finance operator, I want revenue, partner cost, and platform fee data in one place so I can review settlement readiness.

## Epic 8. Store Operations Dashboard

### Outcome

A merchant sees the health of a store immediately after opening backoffice.

### User stories

- As a merchant, I want a home dashboard so I understand store status at a glance.
- As a store operator, I want to see product count, order count, exception count, and shipment risk so I know where to act first.
- As a merchant, I want partner alerts surfaced on the dashboard so I understand execution risk early.

## Epic 9. Platform Governance

### Outcome

Platform operators can safely manage the multi-tenant network.

### User stories

- As a platform owner, I want to manage platform admins so governance is delegated safely.
- As a platform operator, I want audit visibility into sensitive actions so the platform remains trustworthy.
- As a platform operator, I want controlled store creation and partner onboarding so the ecosystem remains clean.

## Recommended release framing

## Release 1. Store Foundation

Focus:

- store creation
- store switching
- team invites
- store roles
- basic store shell

## Release 2. POD Operations Foundation

Focus:

- partner onboarding
- product setup
- publish-ready store catalog

## Release 3. Order And Fulfillment Operations

Focus:

- order intake
- partner routing
- fulfillment state
- exception handling

## Release 4. Commercial Intelligence

Focus:

- margin visibility
- settlement visibility
- operational dashboards

## BA recommendation

If the team wants the fastest route from current platform foundation to believable POD product, the next backlog should prioritize:

1. `Epic 8` Store Operations Dashboard
2. `Epic 3` Partner Onboarding
3. `Epic 4` Product Setup And Publishing
4. `Epic 5` Order Intake And Routing
5. `Epic 6` Fulfillment Tracking And Exceptions
