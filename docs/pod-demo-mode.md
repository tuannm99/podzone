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

In practice, `Print partners` is the main POD-facing area that already uses a real backend path.

## What Is Local-First Today

These parts are intentionally local to the browser for fast experimentation:

- product setup drafts
- catalog candidates
- mock publish state
- mock routed orders
- mock exception handling
- store dashboard analytics derived from those mock states

The local-first data is stored per store in `localStorage`.

Keys currently used:

- `podzone:product-setup:<tenantId>`
- `podzone:mock-orders:<tenantId>`

## Demo Flow

Recommended demo sequence for one store:

1. Open the store workspace.
2. Use `Seed demo store` from store home.
3. Review `Print partners` if you need real partner records.
4. Open `Product setup`.
5. Promote a draft into a catalog candidate.
6. Mock publish at least one candidate.
7. Open `Orders`.
8. Create a routed order.
9. Advance routing and test issue handling.
10. Return to store home to review the dashboard.

## Seed And Reset Behavior

`Seed demo store` does:

- creates demo product drafts
- creates demo catalog candidates
- marks one candidate as `published_mock`
- creates mock routed orders
- creates at least one open operational issue

`Reset demo store` does:

- clears local product setup state
- clears local mock order state

`Reset demo store` does not:

- delete real print partner records
- alter IAM or session state
- remove backend data from real services

## Why This Split Exists

The current project goal is experimentation.

That means:

- minimize early external integrations
- avoid committing too early to cloud/event architecture
- keep enough real backend surface to validate authz and store scoping
- keep product and ops workflow flexible while the business model is still being shaped

## Next Suggested Direction

Only after the team is comfortable with the current demo flow should the project move to deeper backendization, for example:

- backendize product setup on top of the now-renamed `partner` domain
- persist product setup server-side
- persist routed mock orders server-side
- add real analytics or finance summaries
- connect to real partner or channel integrations
