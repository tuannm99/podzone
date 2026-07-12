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

Each requirement is its own file under a per-domain folder, so it can carry
detail (status, acceptance nuance, linked docs) without bloating this index.
Read `traceability-matrix.md` alongside this table for design/architecture/
contract/sprint links per requirement.

| SRS ID | Title | Status |
| --- | --- | --- |
| [SRS-AUTH-001](./auth/SRS-AUTH-001-authentication.md) | Authentication | Active |
| [SRS-IAM-001](./iam/SRS-IAM-001-centralized-authorization.md) | Centralized Authorization | Active |
| [SRS-IAM-002](./iam/SRS-IAM-002-permission-authoring.md) | Permission Authoring | Active |
| [SRS-IAM-003](./iam/SRS-IAM-003-platform-vs-organization-administration-surface.md) | Platform Vs Organization Administration Surface | Planned, post-backbone |
| [SRS-IAM-004](./iam/SRS-IAM-004-platform-admin-bootstrap-and-recovery.md) | Platform Admin Bootstrap And Recovery | Planned, post-backbone |
| [SRS-ONB-001](./onboarding/SRS-ONB-001-workspace-and-store-entry.md) | Workspace And Store Entry | Active |
| [SRS-ONB-002](./onboarding/SRS-ONB-002-store-provisioning-workflow.md) | Store Provisioning Workflow | Active |
| [SRS-ONB-003](./onboarding/SRS-ONB-003-placement-source-of-truth.md) | Placement Source Of Truth | Active |
| [SRS-ONB-004](./onboarding/SRS-ONB-004-store-request-manual-approval.md) | Store Request Manual Approval | Planned, post-backbone |
| [SRS-BO-001](./backoffice/SRS-BO-001-store-scoped-backoffice.md) | Store-Scoped Backoffice | Active |
| [SRS-CAT-001](./backoffice/SRS-CAT-001-product-setup-readiness.md) | Product Setup Readiness | Active |
| [SRS-ORDER-001](./backoffice/SRS-ORDER-001-order-operations.md) | Order Operations | Active |
| [SRS-SETTLEMENT-001](./backoffice/SRS-SETTLEMENT-001-margin-and-settlement-visibility.md) | Margin And Settlement Visibility | Active |
| [SRS-PARTNER-001](./partner/SRS-PARTNER-001-partner-connection.md) | Partner Connection | Active |

`SRS-CAT-001`, `SRS-ORDER-001`, and `SRS-SETTLEMENT-001` live under
`backoffice/` because they are DDD subdomains inside the single
`internal/backoffice` service (catalog, routing/order, settlement), not
separate deployable services — see
`../03-architecture-detail-design/services/backoffice/README.md`.

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
