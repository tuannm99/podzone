# Backoffice DDD Discovery

## Purpose

This document captures the extra business analysis required before migrating `internal/backoffice` to DDD.

It extends the event-storming table with migration decisions, current-code evidence, and unresolved business questions.

The goal is to avoid splitting code by file shape alone.

## Current Business Evidence

Backoffice already behaves like a store-scoped operations console.

Current evidence:

- users enter through tenant/session context
- users must select a store before store-scoped work
- product setup is store-scoped in the current implementation
- order routing, shipment, exception, queue control, settlement, and activity are exposed through GraphQL
- `RoutedOrder` combines several business concerns into one object
- onboarding readiness controls whether a store can be operated

The current `RoutedOrder` should be treated as a composed read model during migration, not as the target aggregate root.

## Business Capabilities To Preserve

| Capability             | Current surface                                     | Target DDD interpretation                                             |
| ---------------------- | --------------------------------------------------- | --------------------------------------------------------------------- |
| Store entry            | Store GraphQL + tenant middleware                   | `Workspace` controls active store scope and readiness                 |
| Product setup          | Catalog GraphQL + `ProductSetup` models             | `Catalog Setup` owns merchant product setup and publish readiness     |
| Order creation         | `createRoutedOrder`                                 | `Order Operations` records customer order facts                       |
| Partner recommendation | `routedOrderRecommendation`                         | `Routing` evaluates eligible partners and block reasons               |
| Forced reroute         | `forceRerouteBlockedOrder`                          | `Routing` owns overrideable routing decision changes                  |
| Order progression      | `advanceRoutedOrder`                                | Needs decision: order lifecycle command or fulfillment-derived status |
| Shipment update        | `updateOrderShipment`                               | `Fulfillment` owns shipment state and tracking                        |
| Exception handling     | `openOrderException`, `updateOrderExceptionStatus`  | `Exception Handling` owns issue lifecycle and SLA                     |
| Queue control          | `updateOrderQueueControl`, bulk update              | `Order Operations` owns operational queue assignment and SLA view     |
| Settlement update      | `updateOrderSettlement`, `updateOrderIssueHandling` | `Settlement` owns costs, margin, reconciliation, payout state         |
| Activity feed          | `routedOrderActivities`                             | `Activity` is a projection from domain events                         |

## Candidate Aggregate Decisions

| Aggregate          | Can start now?     | Reason                                                                          |
| ------------------ | ------------------ | ------------------------------------------------------------------------------- |
| `StoreWorkspace`   | Yes                | Store scope/readiness requirements are already explicit                         |
| `ProductSetup`     | Yes                | Current product setup data already has a clear lifecycle                        |
| `CustomerOrder`    | Yes                | Order intake facts are distinct from routing and fulfillment                    |
| `RoutingDecision`  | Yes                | Partner recommendation and block reason are already separate business decisions |
| `FulfillmentOrder` | Yes                | Shipment state and tracking rules already exist                                 |
| `OrderException`   | Partially          | Need to decide one exception per order or multiple issues per order             |
| `SettlementRecord` | Partially          | Need to decide merchant-facing vs platform-facing settlement visibility         |
| `ActivityEntry`    | Yes, as projection | It should not own source-of-truth business rules                                |

## Decisions Required Before Code Migration

### 1. Partner Assignment Mode

Decision needed:

- auto-assign the recommended partner
- always require operator confirmation
- auto-assign only when policy allows it

Recommended default:

Use policy-based auto-assignment.

Reason:

- current code already selects a partner when recommendation succeeds
- blocked orders still need manual reroute
- IAM/policy can later control which stores are allowed to auto-route

Architecture impact:

- `RoutingDecision` owns recommendation and selection
- a process policy can turn `PartnerRecommended` into `AssignFulfillmentPartner`
- manual override remains a command

### 2. Product Setup Ownership

Decision needed:

- store-owned setup only
- tenant-shared product pool with store-specific usage

Recommended default:

Use tenant-shared catalog pool plus store-specific product setup/listing state.

Reason:

- multitenancy requirements already define catalog as an exception to pure store ownership
- current product setup has `StoreID`, so migration can start with store-scoped state
- future tenant-shared catalog can be introduced without rewriting store operations

Architecture impact:

- `Catalog Setup` aggregate stays store-scoped for now
- shared catalog source remains an external dependency/read source

### 3. Order Lifecycle Source

Decision needed:

- order status is manually advanced
- order status is derived from fulfillment/routing events
- hybrid model

Recommended default:

Use a hybrid model during migration.

Reason:

- current UI has `advanceRoutedOrder`
- shipment updates already influence order status
- removing manual advance immediately would break existing behavior

Architecture impact:

- `CustomerOrder` owns high-level customer order status
- fulfillment events may update an order projection or trigger an order status policy
- `advanceRoutedOrder` can become a temporary command until the target lifecycle is clearer

### 4. Exception Cardinality

Decision needed:

- one active exception per order
- many issue records per order

Recommended default:

Start with one active exception per order, but model `OrderException` as a separate aggregate.

Reason:

- current code stores one `ExceptionType` and one `ExceptionStatus`
- separating the aggregate now avoids keeping exception behavior inside `RoutedOrder`
- multiple issue records can be added later as a storage/API migration

Architecture impact:

- `OrderException` has `OrderID`, status, type, SLA, resolution, notes
- GraphQL can still compose it into `RoutedOrder`

### 5. Settlement Visibility

Decision needed:

- merchant-facing settlement only
- platform-facing settlement only
- both

Recommended default:

Support both views but keep one source-of-truth settlement record for the first migration.

Reason:

- current fields mix merchant-visible values and operational reconciliation status
- finance/platform operators will need stronger controls later

Architecture impact:

- `SettlementRecord` owns cost, margin, reconciliation, and payout state
- query model can expose merchant-safe and platform-safe views separately

### 6. Activity Source Of Truth

Decision needed:

- keep manually appended activity logs
- derive activity from domain events

Recommended default:

Introduce domain events first, then migrate activity feed into a projection.

Reason:

- current code manually appends timeline/activity entries in many command paths
- replacing that directly risks behavior regression
- projection can be introduced while GraphQL remains stable

Architecture impact:

- aggregates record domain events
- interactor persists state and appends/publishes events
- activity query reads from projection/read model

### 7. Durable Events

Decision needed:

- which events require outbox/CDC
- which events can be in-process or normal pub/sub

Recommended default:

Use durable event publication only for cross-service or rebuild-critical events.

Examples:

- Durable: `StoreBecameReady`, `CustomerOrderReceived`, `FulfillmentPartnerAssigned`, `ShipmentDelivered`, `SettlementReconciled`
- Non-durable or in-process first: UI activity entries, local queue assignment updates, local readiness banners

Architecture impact:

- Backoffice should not create outbox tables for every local event
- local domain events can feed in-process policies and projections
- cross-service events go through the platform messaging/outbox/CDC pattern

## Migration Readiness Matrix

| Area               | Ready to migrate? | Required pre-migration work                                         |
| ------------------ | ----------------- | ------------------------------------------------------------------- |
| Workspace          | Yes               | Align store readiness naming with onboarding                        |
| Catalog Setup      | Yes               | Define store-owned setup vs tenant-shared catalog source explicitly |
| Order Operations   | Yes               | Keep GraphQL read model stable while introducing `CustomerOrder`    |
| Routing            | Yes               | Decide auto-route policy default                                    |
| Fulfillment        | Yes               | Decide whether shipment status can drive order status               |
| Exception Handling | Partial           | Confirm one active exception per order for MVP                      |
| Settlement         | Partial           | Split merchant/platform views at query level                        |
| Activity           | Yes               | Introduce event-backed projection gradually                         |

## Proposed Migration Gates

Code migration should start only after these gates are accepted:

1. `RoutedOrder` is agreed to be a GraphQL/read-model object, not the aggregate.
2. Current GraphQL API remains stable in the first migration phase.
3. The first persistence phase keeps existing tables and uses repository mapping.
4. Domain events are introduced in-process first.
5. Durable outbox/CDC is used only where cross-service reliability is required.
6. `OrderException` starts as one active exception per order unless product explicitly changes it.
7. Settlement keeps one source-of-truth record but exposes separate query views later.

## First Migration Slice

The first safe implementation slice should be:

1. Add DDD package boundaries under `internal/backoffice/domain`.
2. Introduce aggregate-specific input/output ports.
3. Move routing business mutations into aggregate behavior methods.
4. Keep existing GraphQL schema unchanged.
5. Keep existing SQL tables unchanged.
6. Add snapshot/rehydration mapping inside infrastructure repositories.
7. Add domain-event collection without turning every event into outbox.

Recommended order:

1. `CustomerOrder`
2. `RoutingDecision`
3. `FulfillmentOrder`
4. `OrderException`
5. `SettlementRecord`
6. `Activity` projection

## Current Migration Status

The first in-process DDD slice is implemented for the current Backoffice codebase.

Implemented aggregate and domain-event coverage:

| Context            | Aggregate / concept        | Current status                                                                 |
| ------------------ | -------------------------- | ------------------------------------------------------------------------------ |
| Store Workspace    | `Store`                    | Validated creation, activate/deactivate behavior, in-process domain events      |
| Catalog Setup      | `ProductSetupDraft`        | Validated draft creation and candidate promotion events                         |
| Order Operations   | `CustomerOrder`            | Create, advance, queue-control, manual reroute behavior and domain events       |
| Routing            | `RoutingDecision`          | Recommendation decision, selected-partner event, routing-blocked event          |
| Fulfillment        | `FulfillmentOrder`         | Shipment transition invariants and shipment status domain events                |
| Exception Handling | `OrderException`           | Open/status lifecycle invariants and exception domain events                    |
| Settlement         | `SettlementRecord`         | Money-based settlement/issue handling behavior and domain events                |

The current `RoutedOrder` GraphQL type remains a composed read model during this migration slice.
`CustomerOrder` now has a dedicated `customer_orders` aggregate store with optimistic versioning. Order command workflows
load that aggregate store instead of reconstructing the aggregate from `RoutedOrder`, then persist the aggregate state and
the existing projection in one tenant transaction.

The other workflow aggregates still map snapshots into the existing `routed_orders` projection while their dedicated
command persistence is introduced incrementally. The GraphQL API remains stable throughout this transition.

Not done in this slice:

- durable domain-event publisher/outbox for Backoffice
- activity feed rebuilt from domain-event projection
- separate command persistence tables for routing, fulfillment, exception, and settlement aggregates
- separate merchant/platform settlement query views

## Open Questions To Confirm

| Question                                               | Recommended temporary answer                     |
| ------------------------------------------------------ | ------------------------------------------------ |
| Is partner assignment automatic or operator-confirmed? | Policy-based auto-assignment                     |
| Is product setup store-owned or tenant-shared?         | Store-owned setup, tenant-shared catalog source  |
| Is order status manually advanced or event-derived?    | Hybrid during migration                          |
| Can one order have multiple active exceptions?         | No, one active exception for MVP                 |
| Is settlement merchant-facing or platform-facing?      | Both views, one source-of-truth record initially |
| Should every event use outbox?                         | No, only cross-service/rebuild-critical events   |
