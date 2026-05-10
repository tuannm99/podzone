# POD Demo Mode

## Purpose

This repository is currently being used as an experimental POD workspace rather than a production-grade commerce platform.

The goal of demo mode is to let the team validate:

- store-scoped workflow language
- POD operations flow
- partner, product, and order handoff concepts
- UI direction before deeper cloud and integration decisions are locked

## What Is Real Backend Today

These parts are backed by real services and persistence:

- authentication and sessions
- IAM, tenant membership, and permission checks
- store access and active tenant switching
- print partner records via the partner API served by `partner-service`
- product setup drafts and catalog candidates via `backoffice`
- published mock product state via `backoffice`
- routed POD orders and issue state via `backoffice`
- manual shipment control via `backoffice`
- settlement readiness and realized margin fields via `backoffice`
- issue cost handling for reprint and delivery exceptions via `backoffice`

In practice, the core POD-facing workspace now runs through real backend paths for partners, product setup, and routed orders.

## What Is Local-First Today

These parts are still intentionally sandboxed even though they now persist server-side:

- publish state remains a mock catalog operation, not a channel integration
- routed orders remain POD sandbox orders, not marketplace or carrier orders
- shipment state is manually managed by store operators, not synced from third-party fulfillment
- settlement status is manually managed by store operators, not synced from finance or fulfillment partners
- issue handling and issue cost updates remain operator workflow only, without downstream automations
- store dashboard analytics are lightweight summaries derived from sandbox data

The browser no longer stores the product setup or order state as the source of truth.

## Demo Flow

Recommended demo sequence for one store:

1. Open the store workspace.
2. Review `Print partners` if you need real partner records.
3. Open `Product setup`.
4. Create a draft and promote it into a catalog candidate.
5. Mock publish at least one candidate.
6. Open `Orders`.
7. Create a routed order.
8. Advance routing and test issue handling.
9. Update shipment, issue cost, and settlement state on the orders board.
10. Return to store home to review the dashboard.

## Why This Split Exists

The current project goal is experimentation.

That means:

- minimize early external integrations
- keep enough real backend surface to validate authz and store scoping
- keep product and ops workflow flexible while the business model is still being shaped
- avoid committing too early to cloud/event architecture or fulfillment-specific eventing

## Next Suggested Direction

After the current sandbox workflow is comfortable, the next layer should move beyond CRUD persistence into execution integrations, for example:

- add real analytics or finance summaries
- connect to real partner or channel integrations
- introduce fulfillment state sync instead of manual route advancement
- add external finance reconciliation and delivery outcome automations
