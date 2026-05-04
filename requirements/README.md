# Podzone Requirements

## Purpose

This folder documents the product requirements view of Podzone as a future POD-first commerce platform, based on the codebase that currently exists.

The goal is not to restate implementation details. The goal is to clarify:

- who the product is for
- what business problems it solves
- how the main actors move through the system
- which business domains already exist in the repo
- which product gaps still separate the current system from a production-grade POD operations platform

## Document map

- [01-product-vision.md](./01-product-vision.md)
  Product vision, product scope, business goals, and core platform proposition.
- [02-actors-and-business-flows.md](./02-actors-and-business-flows.md)
  Primary actors, lifecycle journeys, and high-level business flows.
- [03-domain-map.md](./03-domain-map.md)
  Business capability map, bounded areas, current implementation status, and priority gaps.
- [04-dropship-operating-model.md](./04-dropship-operating-model.md)
  POD-first business context, value chain, business objects, and future capability needs.
- [05-epics-and-user-stories.md](./05-epics-and-user-stories.md)
  Product backlog starter for the POD-first context, grouped by epic.
- [06-ux-copy-guideline.md](./06-ux-copy-guideline.md)
  User-facing language guide for rewriting the product from technical admin wording to POD operations wording.
- [07-partner-mvp.md](./07-partner-mvp.md)
  Interpretation of the current partner implementation inside a POD-first partner model.
- [08-partner-refactor-plan.md](./08-partner-refactor-plan.md)
  Migration plan for evolving the current supplier implementation into a broader partner model.

## How to read these docs

These requirements are written from a Business Analyst perspective.

They intentionally distinguish between:

- `Current system`
  What the repository already supports in some form.
- `Target product`
  What a seller-facing POD operations platform should support.
- `Gap`
  What still needs to be designed or built.

## Working assumptions

The current repository suggests the following intended direction:

- `AuthService` handles authentication, session lifecycle, OAuth, refresh, logout, and tenant switching.
- `IAMService` handles tenant creation, memberships, role bindings, permissions, invites, platform roles, and audit-related admin actions.
- `Backoffice` is the seller/admin portal and is the natural place for operational commerce workflows.
- `Onboarding` manages infrastructure and tenant placement metadata.
- `Catalog`, `Cart`, `Order`, and `Payment` are planned or partially scaffolded platform domains.
- `pdtenantdb` is the runtime foundation for multi-tenant data access.

For the updated product context, these docs now assume Podzone is moving toward a `POD-first operating platform`, not only a generic admin console.

## BA conclusion

At a product level, Podzone should be framed as:

`A multi-tenant commerce operations platform for store owners and their teams, with POD-first workflows across product setup, partner coordination, order handling, fulfillment visibility, and settlement readiness.`

That framing is stronger than the current UI language, which still reads more like an internal IAM or infrastructure console.
