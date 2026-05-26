# Actors And Business Flows

## Actor model

## 1. Platform Owner

The platform owner manages the Podzone platform itself.

Responsibilities:

- approve or create stores when platform governance requires it
- manage platform administrators
- review audit logs
- configure platform-level policies

## 2. Store Owner

The store owner is the main merchant account for a specific store.

Responsibilities:

- create or claim a store
- manage team access
- configure store settings
- oversee product and order operations

## 3. Store Admin / Operations Manager

A team member responsible for day-to-day store operations.

Responsibilities:

- manage staff access if allowed
- maintain products
- process orders
- monitor store health

## 4. Store Staff

Internal operator with limited access.

Responsibilities:

- work on assigned operational tasks
- view or update records according to permissions

## 5. Print Or Fulfillment Partner

This actor is a core actor in the target POD model.

Responsibilities:

- receive production or fulfillment requests
- confirm execution readiness
- ship fulfilled orders
- update shipment or fulfillment state

Current status:

- conceptually required, but not yet implemented as a polished product surface

## 6. End Customer

The buyer who purchases from a merchant storefront.

## 7. Internal Support / Compliance / Finance Roles

These are likely future platform roles.

Responsibilities:

- support issue resolution
- audit store activity
- manage risk or compliance actions
- support billing and settlement

## Actor relationship model

- one user can belong to multiple stores
- one store can have many users
- platform roles are global
- store roles are scoped to one store
- a user acts inside one `active store context` at a time

## High-level business flows

## Flow A. User authentication and session establishment

### Goal

Let a user sign in and obtain a valid session for platform access.

### Current flow

1. User signs in via username/password or Google OAuth.
2. Auth service issues access token and refresh token.
3. Session is created and tracked.
4. User enters the admin surface.

## Flow B. Store creation

### Goal

Let an authorized actor create a new store workspace.

### Current flow

1. Authenticated actor invokes create tenant.
2. IAM checks platform permission `tenant:create`.
3. System creates store/workspace record.
4. System creates owner membership for the actor.

### Business meaning

This is the merchant onboarding entry point.

## Flow C. Enter store workspace

### Goal

Let a user open one of the stores they belong to.

### Current flow

1. User selects a store.
2. System switches `active_tenant_id`.
3. New token/session state reflects the active store.
4. Backoffice uses that active store context for data and permission checks.

## Flow D. Team access management

### Goal

Let a store owner or admin manage who can access a store.

### Current flow

1. Authorized actor opens team access management.
2. Actor adds member directly or by identity.
3. Actor assigns role.
4. Actor can remove member.

## Flow E. Invite and accept team access

### Goal

Let a store owner invite a new or existing user by email.

### Current flow

1. Authorized actor creates invite.
2. System stores invite token hash and metadata.
3. Invitee opens accept link.
4. Invitee authenticates if needed.
5. System validates invite token and email match.
6. Membership is created or updated.
7. Invitee enters store context.

## Flow F. Store-scoped authorization

### Goal

Ensure the user can only see or change data in the current store according to role.

### Current flow

1. Backoffice reads user identity and active store from token.
2. Backoffice calls IAM for membership and permission validation.
3. Backoffice routes DB access through multi-tenant context.

## Flow G. Partner onboarding

### Goal

Make a print or fulfillment partner available to a store.

### Target future flow

1. Merchant or platform creates partner record.
2. Partner details are stored.
3. Partner is activated for the selected store.

## Flow H. Product setup and publishing

### Goal

Let the merchant prepare products and publish them into store operations.

### Target future flow

1. Merchant creates or imports product setup data.
2. Merchant edits sales-facing details.
3. Merchant publishes products.

## Flow I. Order intake and fulfillment routing

### Goal

Turn customer demand into operational work.

### Target future flow

1. Customer order is received.
2. Store order is recorded.
3. Order is routed to the appropriate partner or internal path.
4. Shipment and status updates flow back to the merchant.

## Flow J. Margin and settlement visibility

### Goal

Let merchants understand whether their operations are commercially healthy.

### Target future flow

1. Revenue, partner cost, shipping cost, and platform fee are tracked.
2. Margin is computed.
3. Settlement status becomes visible.
