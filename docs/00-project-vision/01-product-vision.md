# Product Vision

## Product statement

Podzone should be positioned as a `multi-tenant commerce operations platform` for merchants who run one or more online stores, with a `POD-first` operating model.

The platform should let a merchant:

- create and manage one or more stores
- invite team members into each store
- connect print or fulfillment partners
- manage products, orders, and store settings from one backoffice
- grow later into broader sourcing models such as dropship

More specifically, Podzone should become:

`a seller backoffice for multi-store POD businesses`

## Business problem

Small and mid-sized POD teams often operate with fragmented tooling:

- one tool for identity and access
- another for storefront setup
- another for product setup
- another for order operations
- spreadsheets or chat for print partner coordination

Podzone should unify these concerns into a single platform with clear workspace boundaries, while also connecting the merchant side of commerce to the production and fulfillment side of the business.

## Product objective

The product should support `many stores`, each with:

- isolated operational data
- isolated access control
- shared platform governance
- a common backoffice experience

This is why the current technical design uses:

- tenants/workspaces
- IAM memberships and permissions
- active tenant switching
- multi-tenant database routing

From a product perspective, these should be translated into:

- `store`
- `team access`
- `current store`
- `store data isolation`

## Core business outcomes

### 1. Merchant self-service

Merchants should be able to:

- create a store
- access that store immediately
- invite their team
- begin basic operations without engineering support

### 2. Safe team collaboration

Each store should support:

- role-based access
- invite and acceptance lifecycle
- auditability for sensitive admin actions

### 3. Scalable multi-store operations

The same user may belong to multiple stores.

The platform should support:

- switching between stores
- keeping permissions scoped to the selected store
- preventing cross-store data leaks

### 4. POD-first operational expansion

The product should be extensible toward:

- print partner onboarding
- product template and production setup
- order routing to production partners
- shipment tracking
- payment and settlement

### 5. Fulfillment-safe commerce operations

POD merchants need more than product CRUD.

They need the platform to reduce operational risk around:

- delayed production
- failed fulfillment handoff
- weak shipment visibility
- catalog inconsistency across stores
- margin erosion

## Product scope

### Current MVP scope implied by the repo

- user authentication
- session and refresh token management
- Google OAuth login
- store/workspace creation
- team membership management
- invite acceptance
- permission checks
- platform administration
- audit logs
- seller portal shell

### Near-term product scope

- store management
- team access management
- store settings
- partner onboarding
- product setup and publishing
- order operations dashboard
- fulfillment visibility

### Out-of-scope for the current MVP

- customer-facing storefront maturity
- advanced partner SLA workflows
- settlement engine
- dispute management
- logistics network orchestration
- advanced reporting and BI

## Product positioning

### External positioning

Podzone should not be described to end users as:

- IAM
- tenant management
- platform membership console

Podzone should be described as:

- a commerce backoffice
- a multi-store seller operations hub
- a POD operations platform
- a foundation for fulfillment-enabled commerce

## Language model for the UI

The product language should move from technical terms to business terms.

Preferred language:

- `Tenant` -> `Store` or `Workspace`
- `Membership` -> `Team access`
- `Platform roles` -> `Platform administration`
- `Switch active tenant` -> `Open store`
- `Tenant invite` -> `Store team invite`
- `Tenant owner` -> `Store owner`
- `Print partner` -> `Production partner` when the conversation is operational rather than commercial
- `Store product` -> `Product listing` when the context is sales-side

This change matters because product language defines how users understand the system.

## Product risks

### 1. Technical language leaking into user-facing UX

The current UI reads like a developer/admin console.

Risk:

- merchants will not recognize the product as a commerce platform

### 2. Strong access model, weak commerce narrative

The current system is becoming strong in authn/authz, but product workflows still feel infrastructure-first.

Risk:

- a technically correct platform that is not yet commercially intuitive

### 3. Missing core POD journeys

There is not yet a visible end-to-end merchant journey like:

- create store
- connect print partner
- set up products
- publish catalog
- process order
- track fulfillment issues

Risk:

- the product promise remains abstract

## BA recommendation

The next product framing should be:

1. Treat `Backoffice` as the main product surface.
2. Rename user-facing concepts from infrastructure language to commerce language.
3. Build around the merchant and print/fulfillment partner journey, not around system capabilities.
4. Keep IAM, session, and tenant routing as internal enabling layers.
5. Keep the domain extensible toward dropship, but do not let dropship define the primary story.

In short:

`The codebase already has the right foundation for a multi-store platform. The next product step is to make the POD business model visible and understandable to merchants and operators.`
