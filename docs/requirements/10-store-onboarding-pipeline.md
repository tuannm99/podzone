# Store Onboarding Pipeline

## Purpose

This document defines the business rules for creating and activating a store inside a tenant workspace.

The goal is to remove ambiguity around:

- when a store is considered created
- when it is usable in Backoffice
- when DB placement exists
- when admin approval is required
- which service owns the provisioning pipeline

This is a requirements document, not an implementation plan.

## Business framing

### Workspace

A workspace is the merchant's tenant-level account.

It owns:

- memberships
- permissions
- one or more stores
- onboarding requests for infrastructure

### Store

A store is an operational unit inside a workspace.

A store:

- belongs to exactly one workspace
- is not usable until onboarding has completed
- requires a resolved runtime/data placement before store-scoped operations can run

### Store creation request

Creating a store is not a single instant action.

It is a request that may require:

- capacity checks
- database or schema allocation
- connection selection
- placement publication
- approval

The product must distinguish between:

- `store requested`
- `request queued`
- `pending approval`
- `store being provisioned`
- `store ready`
- `store active`

## Request tracking model

Every store provisioning request must be traceable.

Minimum fields:

- request ID
- workspace ID
- requested store name and slug
- requested by actor
- request timestamp
- current processing stage
- approval state
- selected connection or target pool
- selected placement mode
- last error, if any
- retry count

The UI and admin surfaces must be able to show:

- pending request
- approved but not yet provisioned
- provisioning in progress
- failed with reason
- ready to open

The user must never be left with an unlabelled failure state.

## Required workflow

### Store creation

1. A workspace owner or authorized operator submits a store creation request.
2. The request is persisted as a trackable onboarding item.
3. The system evaluates whether the request can be auto-approved.
4. If approval is required, the request enters a pending approval queue.
5. If approved, onboarding resolves the target infrastructure placement.
6. Onboarding provisions the required infrastructure and writes the placement metadata.
7. A background worker or queue consumer performs the long-running provisioning work.
8. Backoffice can open the store only after placement is resolvable and the store is ready.

### Approval paths

The system must support both:

- automatic approval
- admin approval

Automatic approval is allowed only when policy, quota, and infrastructure checks pass.

Admin approval is required when:

- tenant quota is exceeded
- database capacity is constrained
- a higher-risk configuration is requested
- platform policy requires manual review

Approval must be recorded as part of the request lifecycle, not as an implicit side effect.

## Provisioning checks

The onboarding pipeline must validate:

- database capacity
- cluster selection
- schema or database allocation mode
- connection availability
- tenant placement readiness
- store readiness state after publish
- connection health before accepting the request
- whether the request can be placed on an existing tenant pool or needs a new one
- whether the requested store violates tenant quotas or platform policy

The product should not assume a store can be opened just because the tenant exists.

The product should not assume a store can be provisioned synchronously from the UI request path.

## Placement requirements

A store becomes operational only when the tenant has a resolvable placement.

That placement must be published through onboarding infrastructure and consumed by `pdtenantdb`.

Requirements:

- placement must exist before store-scoped runtime can be entered
- missing placement should surface as a provisioning or readiness problem, not as a generic UI failure
- placement publication is part of onboarding, not a backoffice-side fallback
- placement resolution must point to the selected tenant DB, schema, or equivalent storage target
- placement must be refreshed or republished when the storage target changes

## Store lifecycle

The product should treat store lifecycle as:

- `requested`
- `queued`
- `pending_approval`
- `provisioning`
- `failed`
- `ready`
- `active`
- `suspended`
- `archived`

The exact technical status names can differ, but the product must preserve the lifecycle meaning.

## Visibility rules

- Requested stores may appear in admin surfaces as pending items.
- A store that is not ready must not appear as selectable for store-scoped operations.
- Backoffice should only open stores that are ready and resolvable.
- Admin surfaces must expose the reason a store is blocked, failed, or waiting.
- A store request may be visible even if the store itself is not yet selectable.

## Responsibilities by service

### IAM

IAM owns:

- workspace creation
- membership and permission assignment
- approval actor identity
- approval policy metadata if policy is defined at the IAM layer

IAM does not own placement publication.
IAM does not execute the long-running provisioning steps.

### Onboarding

Onboarding owns:

- store provisioning requests
- queueing or dispatching the provisioning work
- connection and placement publication
- infra readiness tracking
- approval-driven provisioning flow
- resource checks and placement selection

### Queue or worker runtime

A queue consumer or worker may execute the actual provisioning steps.

That runtime owns:

- polling or consuming pending requests
- calling the appropriate infra services
- recording success or failure
- republishing placement after successful changes

### Backoffice

Backoffice owns:

- displaying store readiness
- opening only ready stores
- reporting provisioning/readiness state to the user

## Product requirement summary

The user flow should feel like this:

1. Sign in.
2. Choose workspace.
3. Request or select a store.
4. If the store is pending, wait for onboarding or approval.
5. If the store is ready, enter Backoffice.

The user should never be forced to infer whether a store exists, whether placement exists, or whether the store is usable.
