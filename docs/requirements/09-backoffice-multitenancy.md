# Backoffice Multitenancy Requirements

## Purpose

This document captures the business requirements for evolving `Backoffice` from a tenant-aware seller console into a true multi-tenant, multi-store operating surface.

This is a requirements and architecture-shaping document.

It is not an implementation plan for feature delivery yet.

## Business framing

### Tenant

A `tenant` is a business account in Podzone.

A tenant:

- represents one business entity
- owns one or more stores
- has its own branch structure and storefront domain setup
- owns users, memberships, and permissions through IAM

### Store

A `store` is an operating unit that belongs to exactly one tenant.

A store:

- cannot be transferred to another tenant
- has its own operational lifecycle
- is the primary isolation unit for backoffice workflows
- may share selected assets or defaults from the tenant level

### Core hierarchy

The primary business hierarchy is:

`organization -> tenant -> store -> order`

For current refactor scope, the minimum required hierarchy is:

`tenant -> store -> order`

## Required tenancy model

### Ownership rules

- One tenant can own one or many stores.
- One store belongs to exactly one tenant.
- Store reassignment between tenants is not supported.
- Users belong to the tenant context and can be granted permissions per store.

### Access model

Access control must apply at both levels:

- `tenant-level`
- `store-level`

That means:

- a user can belong to a tenant
- a user may have permissions only for selected stores in that tenant
- cross-store access is not automatically granted just because the user belongs to the tenant

## Scope and isolation requirements

The following backoffice domains must be isolated by `store`:

- routed orders
- product setup
- shipment
- settlement
- activity and audit views
- partner configuration
- store configuration

### Catalog exception

Catalog is not fully store-owned.

Requirements:

- tenant has a shared POD catalog pool
- stores consume from that shared tenant catalog pool
- store-specific state can still exist around setup, readiness, publishing, or configuration

So the target model is:

- `tenant-shared catalog`
- `store-specific operational usage`

## User journey requirements

### Login and navigation

Backoffice login enters a user into a `tenant`.

After login:

- user lands in tenant context first
- user selects a store explicitly
- opening a store should open a separate tab/workspace context

This means the runtime model must support:

- current tenant
- current store

The current tenant alone is not enough for target backoffice behavior.

### URL and navigation shape

User-facing routes should not be forced to expose both raw `tenantId` and `storeId`.

Backoffice should follow a workspace-shell model closer to seller platforms such as Shopify, Amazon Seller, or TikTok Shop Seller Center:

- user enters the current tenant/workspace through session context
- user switches store through an explicit store switcher
- store context can be reflected by shell state, lightweight slug routing, query state, or a dedicated store route when needed

Examples of acceptable UX shapes:

- `/orders`
- `/catalog`
- `/finance`
- `/stores/:storeSlug/orders`
- `/orders?store=:storeSlug`

So the requirement is:

- `tenant + store` must always be explicit in runtime/application scope
- public UX routes do not need to expose `/t/:tenantId/stores/:storeId/...` as the canonical pattern

## Store lifecycle requirements

Store has its own lifecycle.

The lifecycle should follow the onboarding model managed by `OnboardingService`.

At requirement level, backoffice must assume store states similar to:

- requested
- pending approval
- provisioning
- onboarding
- ready
- active
- suspended
- archived

Exact status names can be aligned later with the onboarding implementation, but store lifecycle is required as a first-class concept.

### Store creation expectation

Creating a store is a request into the onboarding pipeline, not a direct synchronous shell action.

The backoffice must support:

- submit store request
- review approval state
- show provisioning/readiness state
- open only stores that are ready and resolvable
- show when a request is queued, blocked, failed, or waiting for approval

Backoffice must not assume that a newly submitted store request is immediately usable.
The workspace can exist before any store becomes selectable.

Store creation may require admin approval when:

- workspace quota is exceeded
- database capacity is unavailable
- a higher-risk placement is requested
- policy requires manual review

### Provisioning dependency

Backoffice store-scoped access depends on onboarding publishing tenant placement metadata.

That means:

- tenant existence alone is not enough
- a store request is not enough until onboarding completes
- a store cannot be opened until `pdtenantdb` can resolve placement
- missing placement should be treated as a readiness problem
- onboarding owns placement publication, not backoffice
- if placement is missing, the UI should surface onboarding status instead of presenting the store as openable

## Bootstrap and readiness requirements

Bootstrap and readiness checks are required at both levels:

- tenant level
- store level

This means:

- tenant infrastructure readiness alone is not sufficient
- store readiness alone is not sufficient
- placement resolution must be available before store-scoped operations are treated as ready
- backoffice runtime must validate readiness before allowing store-scoped navigation
- readiness checks must distinguish between:
  - tenant exists
  - store request exists
  - placement exists
  - store is openable

## Data isolation requirements

Target isolation direction is `store-level`.

This does not necessarily require a separate physical database per store immediately.

But the architecture must support store-level isolation semantics in:

- application contracts
- repository boundaries
- read/write filtering
- audit visibility
- lifecycle checks

Current tenant-level-only routing is not enough as the target architecture.

## Runtime and deployment routing requirements

Backoffice multitenancy has two different routing layers that must both be modeled explicitly.

### 1. Network or edge routing

At the infrastructure layer, requests must be routed to the correct `backoffice` runtime pool for the tenant.

This is an edge/runtime concern, not a business-usecase concern.

Examples of future production routing targets:

- tenant-to-namespace mapping in Kubernetes
- tenant-to-runtime pool mapping
- tenant-aware ingress or gateway rules
- sticky or deterministic tenant placement toward the correct backoffice shard/pool

This means the architecture must support:

- stateless backoffice API pods inside a tenant-assigned pool
- scale-out and scale-in without breaking tenant routing semantics
- pool-aware routing driven by external placement metadata, not hard-coded in business logic

### 2. Application-to-data routing

After the request reaches the correct backoffice runtime pool, the application must still route tenant traffic to the correct data placement.

This is the `application -> database placement` step.

The architecture must support:

- tenant-to-cluster routing
- tenant-to-database or tenant-to-schema routing
- store-scoped filtering inside the tenant placement
- readiness/bootstrap checks before use

In current runtime terms, this corresponds to:

- edge routes tenant into the correct backoffice runtime pool
- backoffice runtime resolves tenant placement
- repositories route into the correct DB/schema
- store-scoped operations then execute inside that tenant placement

### Scale requirements

Scale behavior must remain safe when backoffice scales horizontally.

Requirements:

- API instances must stay stateless with respect to tenant ownership
- runtime pool assignment must come from external placement metadata
- scale-out must allow multiple equivalent pods within one tenant pool
- scale-in must not require application-level tenant rebinding logic
- database placement routing must remain deterministic regardless of pod count

So the target model is:

- `tenant routed to runtime pool at edge`
- `tenant routed to data placement in app runtime`
- `store scoped inside tenant placement`

## Cross-store operations

Cross-store operations are valid.

However, they must be controlled by permissions.

This means:

- the system must support aggregate or cross-store views
- cross-store access is not automatically allowed
- users need explicit permission for cross-store visibility or actions

Examples of possible future cross-store operations:

- finance summaries across stores
- operational dashboards across stores
- cross-store catalog usage
- cross-store partner analytics

## Configuration scope requirements

### Partner and cost rules

Partner and cost/routing configuration must support both:

- tenant-level defaults
- store-level overrides

This implies inheritance behavior is required later:

- tenant default
- optional store override

### Catalog and product setup

Catalog is shared at tenant level.

Product setup and operational store usage can still be store-scoped.

Target model:

- tenant shared catalog pool
- store-specific product setup or activation state where needed

## Future organization support

Organization-level hierarchy is required in the future.

That means the target architecture must not assume:

- tenant is always the top-most business boundary

It should stay compatible with:

`organization -> tenant -> store`

## Architecture implications

The backoffice architecture must move toward:

### 1. Explicit runtime scope

Backoffice should not rely only on ambient `tenantID` in context.

It should introduce explicit runtime scope concepts such as:

- `TenantContext`
- `StoreContext`

### 2. Store-scoped operations

Business contracts should become explicit about store scope.

Examples:

- routing usecases should become store-scoped by default
- operational reads should know both tenant and store
- cross-store reads should be modeled as explicit privileged operations

### 3. Store validation layer

Backoffice needs a store resolution and validation layer that can answer:

- does this store exist?
- does this store belong to the current tenant?
- is this store ready?
- is this store active?
- does the current user have access to this store?

### 4. Permission model compatibility

IAM and backoffice must converge on:

- tenant membership
- store-level authorization
- cross-store authorization as an explicit privileged capability

### 5. Tenant runtime and placement awareness

Backoffice should have an internal runtime layer that can answer:

- what tenant is the current request operating under?
- what store is currently selected?
- what placement owns this tenant?
- is this tenant ready in the current runtime pool?
- is this store valid inside the resolved tenant placement?

This runtime layer should sit between edge resolution and business usecases.
- role/permission combinations per store

### 5. Tenant-shared, store-scoped split

Backoffice must clearly distinguish:

- tenant-shared assets and defaults
- store-scoped operations and isolated workflows

## Current system vs target product

### Current system

- Backoffice is tenant-aware.
- Runtime tenant is resolved from auth session.
- Repository access already uses `pdtenantdb`.
- Store exists as a domain concept.
- Routing and operations largely behave as tenant-scoped workflows.

### Target product

- Backoffice is tenant-and-store scoped.
- Store is the primary operational isolation unit.
- Users can be scoped to subsets of stores within a tenant.
- Cross-store operations exist but are permission-controlled.
- Tenant-shared catalog and defaults coexist with store-scoped workflows.

### Gap

The main gaps are:

- no explicit store runtime context model
- no clear store-level permission architecture in backoffice
- domain contracts still lean too heavily on tenant ambient context
- no formal distinction between tenant-shared and store-scoped capabilities
- readiness/bootstrap semantics are not yet modeled for both tenant and store together

## Phase objective

This phase does **not** require full feature implementation.

The phase objective is:

- finalize the multitenancy business model
- align the backoffice architecture to that model
- prepare the codebase so later feature work follows the correct tenancy boundaries

## BA conclusion

Backoffice should no longer be treated as only `tenant-aware`.

It should be designed as:

`a tenant-owned, store-isolated operations workspace, where users enter through tenant identity, operate within explicit store scope, and can perform cross-store work only when permissions allow it.`
