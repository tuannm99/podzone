# Backoffice Event Storming

## Purpose

This document captures the Backoffice business workflow as an event-storming table.

The goal is to guide the DDD migration of `internal/backoffice` without forcing an immediate database or API rewrite.

## Scope

Backoffice is the store-scoped operating surface for POD merchants and operators.

Current implementation already contains:

- `domain/store`
- `domain/catalog`
- `domain/routing`

Target DDD migration should split the current broad `routing` model into smaller business contexts while keeping GraphQL stable during migration.

## Candidate Bounded Contexts

| Context            | Purpose                                                  | Current source                    | Target direction                                              |
| ------------------ | -------------------------------------------------------- | --------------------------------- | ------------------------------------------------------------- |
| Workspace          | Tenant/store entry, active store scope, readiness gate   | `domain/store`, tenant middleware | Own store-scoped workspace rules and readiness checks         |
| Catalog Setup      | Product setup, activation readiness, publishing intent   | `domain/catalog`                  | Own merchant-facing product setup state                       |
| Order Operations   | Store customer order intake and order-level lifecycle    | `domain/routing`                  | Split customer order state from routing and fulfillment state |
| Routing            | Partner eligibility, recommendation, assignment, reroute | `domain/routing`                  | Own routing decisions and routing block reasons               |
| Fulfillment        | Partner-facing fulfillment work, shipment, tracking      | `domain/routing`                  | Own fulfillment order and shipment lifecycle                  |
| Exception Handling | Operational issue workflow, escalation, resolution, SLA  | `domain/routing`                  | Can start inside fulfillment, but should be explicit          |
| Settlement         | Fulfillment cost, shipping cost, margin, reconciliation  | `domain/routing`                  | Own financial state and settlement readiness                  |
| Activity           | Store activity feed and operator timeline                | `domain/routing` activity records | Projection/read model, not source of truth                    |

## Event Storming Table

| Step | Command / Intent                       | Domain Event                   | Aggregate / Entity | Bounded Context    | Policy / Rule                                                                     | Read Model / View                | External Dependency            |
| ---- | -------------------------------------- | ------------------------------ | ------------------ | ------------------ | --------------------------------------------------------------------------------- | -------------------------------- | ------------------------------ |
| 1    | User logs in and opens Backoffice      | BackofficeSessionStarted       | StoreWorkspace     | Workspace          | User must have tenant membership                                                  | Workspace shell                  | Auth / IAM                     |
| 2    | User selects a store                   | StoreSelected                  | StoreWorkspace     | Workspace          | Store must belong to active tenant and user must have store access                | Current store scope              | IAM                            |
| 3    | System checks store readiness          | StoreReadinessConfirmed        | StoreWorkspace     | Workspace          | Store cannot open until onboarding placement is ready                             | Store switcher, readiness banner | Onboarding / `pdtenantdb`      |
| 4    | Merchant requests a new store          | StoreRequestSubmitted          | StoreRequest       | Workspace          | Creation is asynchronous and may require approval                                 | Store request status             | Onboarding                     |
| 5    | Store provisioning completes           | StoreBecameReady               | StoreWorkspace     | Workspace          | Placement must be resolvable before operation starts                              | Store list                       | Onboarding / Consul projection |
| 6    | Merchant reviews a product candidate   | ProductCandidateReviewed       | ProductSetup       | Catalog Setup      | Candidate must be visible in tenant/store catalog scope                           | Product setup list               | Catalog source                 |
| 7    | Merchant creates product setup         | ProductSetupCreated            | ProductSetup       | Catalog Setup      | Product setup belongs to store usage, while catalog pool may be tenant-shared     | Product setup detail             | Catalog source                 |
| 8    | Merchant updates merchandising/pricing | ProductSetupUpdated            | ProductSetup       | Catalog Setup      | Pricing and margin inputs must be valid before publish readiness                  | Product setup form               | None                           |
| 9    | Product setup passes readiness         | ProductSetupApproved           | ProductSetup       | Catalog Setup      | Required production and listing fields must be complete                           | Setup readiness status           | Partner capability data        |
| 10   | Merchant publishes product             | ProductPublished               | ProductListing     | Catalog Setup      | Only approved setup can be published                                              | Store catalog page               | Storefront / channel API       |
| 11   | Customer order arrives                 | CustomerOrderReceived          | CustomerOrder      | Order Operations   | Order belongs to one store and must be immutable enough for audit                 | Orders list                      | Storefront / order source      |
| 12   | System checks routability              | OrderRoutabilityChecked        | CustomerOrder      | Order Operations   | Order must have product type, ship region, and quantity                           | Routing status                   | Catalog Setup                  |
| 13   | System requests partner recommendation | PartnerRecommendationRequested | RoutingDecision    | Routing            | Recommendation needs product type, region, preferred partner, and active partners | Recommendation panel             | Partner service                |
| 14   | Partner is recommended                 | PartnerRecommended             | RoutingDecision    | Routing            | Eligible partner must support product type and ship region                        | Routing recommendation           | Partner service                |
| 15   | No partner is eligible                 | OrderRoutingBlocked            | RoutingDecision    | Routing            | Block reason must be explicit and operator-visible                                | Blocked order queue              | Partner service                |
| 16   | Operator forces reroute                | BlockedOrderRerouted           | RoutingDecision    | Routing            | Preferred partner must be eligible or explicitly overrideable by policy           | Routing timeline                 | IAM / Partner service          |
| 17   | Partner is assigned                    | FulfillmentPartnerAssigned     | RoutingDecision    | Routing            | Assignment is store-scoped and must be auditable                                  | Order detail                     | Partner service                |
| 18   | Fulfillment work is created            | FulfillmentOrderCreated        | FulfillmentOrder   | Fulfillment        | A routed order creates partner-facing fulfillment work                            | Fulfillment queue                | Partner service                |
| 19   | Shipment label is requested            | ShipmentLabelRequested         | FulfillmentOrder   | Fulfillment        | Shipment cannot progress before partner assignment                                | Shipment status                  | Partner service / carrier      |
| 20   | Label is ready                         | ShipmentLabelReady             | FulfillmentOrder   | Fulfillment        | Tracking metadata should be captured when available                               | Shipment panel                   | Partner service / carrier      |
| 21   | Shipment moves in transit              | ShipmentInTransit              | FulfillmentOrder   | Fulfillment        | Carrier and tracking number should be stored                                      | Order shipment timeline          | Carrier                        |
| 22   | Shipment is delivered                  | ShipmentDelivered              | FulfillmentOrder   | Fulfillment        | Delivered state can trigger settlement readiness                                  | Fulfillment completed queue      | Carrier / partner              |
| 23   | Shipment issue is detected             | ShipmentIssueDetected          | FulfillmentOrder   | Fulfillment        | Delivery issue should open or link an exception                                   | Issue queue                      | Carrier / partner              |
| 24   | Operator opens exception               | OrderExceptionOpened           | OrderException     | Exception Handling | Exception type is required and SLA should start                                   | Exception detail                 | None                           |
| 25   | Operator escalates exception           | OrderExceptionEscalated        | OrderException     | Exception Handling | Escalation requires open exception                                                | Escalation queue                 | IAM                            |
| 26   | Operator resolves exception            | OrderExceptionResolved         | OrderException     | Exception Handling | Resolution should capture notes and cost impact when relevant                     | Resolved issue view              | None                           |
| 27   | Operator updates queue owner           | OperatorAssigned               | WorkQueueItem      | Order Operations   | Assignee must be store operator or allowed actor                                  | Operations queue                 | IAM                            |
| 28   | Operator updates SLA                   | OrderSlaUpdated                | WorkQueueItem      | Order Operations   | SLA date should be visible in queue and feed                                      | SLA dashboard                    | None                           |
| 29   | Operator updates fulfillment cost      | FulfillmentCostUpdated         | SettlementRecord   | Settlement         | Cost must be numeric and currency-safe                                            | Settlement panel                 | Partner invoice/source         |
| 30   | Operator updates shipping cost         | ShippingCostUpdated            | SettlementRecord   | Settlement         | Shipping cost affects margin                                                      | Settlement panel                 | Carrier / partner              |
| 31   | System recalculates margin             | MarginCalculated               | SettlementRecord   | Settlement         | Margin uses retail revenue minus fulfillment, shipping, issue, and platform fees  | Margin view                      | Pricing/order data             |
| 32   | Settlement is reconciled               | SettlementReconciled           | SettlementRecord   | Settlement         | Reconciliation requires enough cost and delivery evidence                         | Reconciliation queue             | Finance system                 |
| 33   | Settlement is paid                     | SettlementPaid                 | SettlementRecord   | Settlement         | Paid state should be final or tightly controlled                                  | Payout history                   | Finance system                 |
| 34   | Any operational change happens         | ActivityRecorded               | ActivityEntry      | Activity           | Activity is append-only and store-scoped                                          | Activity feed                    | Domain events                  |

## Cross-Context Process Policies

| Trigger Event                                | Policy                                           | Resulting Command / Event      |
| -------------------------------------------- | ------------------------------------------------ | ------------------------------ |
| StoreBecameReady                             | Enable store-scoped operations                   | StoreAvailableForOperations    |
| ProductSetupApproved                         | Product can be published                         | PublishProductListing          |
| CustomerOrderReceived                        | Evaluate whether order can be routed             | OrderRoutabilityChecked        |
| OrderRoutabilityChecked                      | Ask routing to recommend partner                 | PartnerRecommendationRequested |
| PartnerRecommended                           | Assign partner if auto-routing is allowed        | FulfillmentPartnerAssigned     |
| OrderRoutingBlocked                          | Put order into blocked queue                     | ActivityRecorded               |
| FulfillmentPartnerAssigned                   | Create fulfillment work                          | FulfillmentOrderCreated        |
| ShipmentDelivered                            | Prepare settlement record                        | SettlementEstimated            |
| ShipmentIssueDetected                        | Open operational exception                       | OrderExceptionOpened           |
| OrderExceptionResolved                       | Recalculate margin if issue cost changed         | MarginCalculated               |
| FulfillmentCostUpdated / ShippingCostUpdated | Recalculate margin                               | MarginCalculated               |
| Any domain event                             | Append activity feed entry when operator-visible | ActivityRecorded               |

## Aggregate Boundaries

| Aggregate        | Owns                                              | Should not own                    |
| ---------------- | ------------------------------------------------- | --------------------------------- |
| StoreWorkspace   | active store scope, readiness, display metadata   | product setup, orders, settlement |
| ProductSetup     | setup state, readiness, publish intent            | customer order lifecycle          |
| CustomerOrder    | customer-facing order facts and high-level status | partner recommendation internals  |
| RoutingDecision  | eligible partners, selected partner, block reason | shipment tracking                 |
| FulfillmentOrder | partner-facing work, shipment, delivery state     | commercial settlement             |
| OrderException   | issue type, status, SLA, resolution               | base shipment state               |
| SettlementRecord | cost, margin, reconciliation, payout state        | routing decision                  |
| ActivityEntry    | timeline projection                               | source-of-truth business rules    |

## Migration Notes

1. Keep the current GraphQL contract stable while splitting the domain internals.
2. Treat the current `RoutedOrder` GraphQL type as a composed read model, not the target aggregate.
3. Start by splitting current `domain/routing` into feature-owned files or subpackages:
   - `domain/order`
   - `domain/routing`
   - `domain/fulfillment`
   - `domain/settlement`
   - `domain/activity`
4. Keep existing SQL tables at first; introduce new repositories/output ports per context before changing storage.
5. Add domain events in-process first. Add outbox only for events that cross service boundaries or must survive process failure.
6. Avoid GraphQL resolver business rules. Resolvers should map request DTOs to commands/queries and call usecases.
7. Activity feed should be a projection fed by domain events, not manually assembled from every aggregate forever.

## Open Questions

| Question                                                                          | Why it matters                                                        |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| Is partner assignment automatic or always operator-confirmed?                     | Determines whether routing is a policy or a manual command by default |
| Is settlement merchant-facing, platform-facing, or both?                          | Determines visibility, permissions, and read models                   |
| Are product setup records store-owned or tenant-shared with store overrides?      | Determines catalog aggregate boundary                                 |
| Should exceptions be one aggregate per order or multiple issue records per order? | Determines SLA and activity feed model                                |
| Which events must be persisted in outbox?                                         | Determines consistency requirements and infra cost                    |
