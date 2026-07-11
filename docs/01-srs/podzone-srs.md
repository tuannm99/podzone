# Podzone Software Requirements Specification

Status: draft baseline for AI-assisted delivery.

## 1. Purpose

Podzone is a multi-tenant POD-first commerce operations platform. It lets store
owners and their teams manage workspace access, store onboarding, product setup,
partner coordination, order routing, fulfillment visibility, and settlement
readiness.

This SRS defines the system-level requirements that future design, architecture,
sprint planning, and agent tasks must trace to.

## 2. Scope

In scope:

- authentication and session lifecycle;
- centralized IAM and permission administration;
- workspace and store selection;
- store onboarding, provisioning, placement, and readiness;
- store-scoped Backoffice operations;
- product setup and catalog readiness;
- partner connection and routing;
- order operations;
- fulfillment exceptions;
- margin and settlement visibility;
- platform/infrastructure administration needed to operate the above.

Out of scope for the current MVP:

- complete public storefront commerce;
- complete payment processor integration;
- production Terraform/cloud provider automation;
- extraction of IAM into an independent open-source repository;
- full marketplace or app ecosystem.

## 3. Actors

| Actor | Description |
| --- | --- |
| Store owner | Owns one workspace and operates one or more stores. |
| Store operator | Works inside a store with assigned permissions. |
| Organization root | First durable owner of one organization/workspace. |
| Platform operator | Operates infrastructure, onboarding, and support workflows. |
| System admin | Deployment/system-level administrator, separate from tenant owner. |
| Partner operator | Represents production, print, or fulfillment partner workflows. |
| AI coding agent | Implements bounded tasks from approved docs and contracts. |

Detailed actor flows live in `../00-project-vision/02-actors-and-business-flows.md`.

## 4. System Context

The C4 context and container architecture are defined in:

- `../02-architecture-overall/01-c4.md`
- `../02-architecture-overall/02-system-context.md`
- `../02-architecture-overall/03-containers.md`

## 5. Functional Requirements

### SRS-AUTH-001 Authentication

The system shall allow users to sign in, maintain sessions, refresh tokens, and
sign out.

Linked docs:

- `../00-project-vision/iam-mvp.md`
- `../02-architecture-overall/04-data-ownership.md`

### SRS-IAM-001 Centralized Authorization

The system shall enforce authorization at backend service boundaries using IAM
decisions. Frontend permission checks shall not be treated as security
boundaries.

Linked docs:

- `../00-project-vision/13-iam-permission-authoring.md`
- `../03-architecture-detail-design/11-iam-platform.md`

### SRS-IAM-002 Permission Authoring

The system shall let authorized administrators manage policies and permissions
without memorizing raw permission strings or principal IDs.

Linked docs:

- `../00-project-vision/13-iam-permission-authoring.md`

### SRS-ONB-001 Workspace And Store Entry

The system shall require users to choose a workspace and then a ready store
before entering store-scoped Backoffice operations.

Linked docs:

- `../00-project-vision/09-backoffice-multitenancy.md`
- `../00-project-vision/10-store-onboarding-pipeline.md`

### SRS-ONB-002 Store Provisioning Workflow

The system shall treat store creation as a tracked onboarding workflow, not a
single synchronous create operation.

Linked docs:

- `../00-project-vision/10-store-onboarding-pipeline.md`

### SRS-ONB-003 Placement Source Of Truth

The system shall treat onboarding placement allocation as the source of truth.
Runtime KV entries are projections that can be rebuilt.

Linked docs:

- `../../internal/onboarding/README.md`
- `../02-architecture-overall/04-data-ownership.md`

### SRS-BO-001 Store-Scoped Backoffice

The system shall expose Backoffice as a store-scoped operating surface after
store readiness is verified.

Linked docs:

- `../00-project-vision/09-backoffice-multitenancy.md`
- `../00-project-vision/11-backoffice-event-storming.md`

### SRS-CAT-001 Product Setup Readiness

The system shall let a store define product setup state required before order
operations become meaningful.

Linked docs:

- `../00-project-vision/03-domain-map.md`
- `../00-project-vision/05-epics-and-user-stories.md`

### SRS-PARTNER-001 Partner Connection

The system shall model print, production, or fulfillment partners used by the
store.

Linked docs:

- `../00-project-vision/07-partner-mvp.md`
- `../00-project-vision/08-partner-refactor-plan.md`

### SRS-ORDER-001 Order Operations

The system shall support store operators managing order routing and fulfillment
state.

Linked docs:

- `../00-project-vision/11-backoffice-event-storming.md`
- `../00-project-vision/12-backoffice-ddd-discovery.md`

### SRS-SETTLEMENT-001 Margin And Settlement Visibility

The system shall help merchants and platform operators understand commercial
health and settlement readiness.

Linked docs:

- `../00-project-vision/04-dropship-operating-model.md`
- `../00-project-vision/05-epics-and-user-stories.md`

## 6. Non-Functional Requirements

### SRS-NFR-001 Tenant Isolation

Tenant and store-scoped data shall be isolated by the selected runtime
placement strategy and by service-layer authorization.

### SRS-NFR-002 Fail Closed Authorization

Unknown permissions, missing scope, or failed permission evaluation shall deny
the operation by default.

### SRS-NFR-003 Source-Of-Truth Integrity

Projection stores shall not become authoritative for business or placement
state.

### SRS-NFR-004 Auditability

Sensitive actions and provisioning state changes shall be auditable.

### SRS-NFR-005 Operability

Operators shall be able to inspect blocked onboarding, route projection drift,
capacity state, and failure reasons without direct database edits.

### SRS-NFR-006 AI-Agent Safety

AI agents shall implement only scoped tasks traceable to approved requirements,
contracts, and sprint plans.

## 7. MVP Backbone Flow

The first reliable backbone flow is:

```text
sign in
  -> choose workspace
  -> request or select store
  -> onboarding placement resolves
  -> open store-scoped Backoffice
  -> call one protected business API
```

This flow should be stabilized before adding broad product surface area.

## 8. Traceability

Use `traceability-matrix.md` to map each SRS requirement to:

- requirement document;
- design/UI spec;
- architecture/C4 doc;
- API/DB contract;
- sprint;
- agent task;
- verification evidence.
