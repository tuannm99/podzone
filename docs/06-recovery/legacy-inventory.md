# Legacy Inventory

Status: draft template.

Use this to classify the current code before refactoring.

## Services

| Service | Runs in Docker? | Has tests? | Owns data | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| auth | Unknown | Yes | auth/session data | Stabilize | Verify login/session flow. |
| iam | Unknown | Yes | IAM policy/membership data | Stabilize | Keep service-side authorization. |
| onboarding | Unknown | Yes | store requests, placement allocation, resource inventory | Stabilize | Backbone dependency. |
| backoffice | Unknown | Yes | store-scoped operating data | Stabilize | Must open only ready store. |
| grpcgateway | Unknown | Yes | none | Stabilize | HTTP entry to gRPC services. |
| partner | Unknown | Partial | partner data | Later | Not first backbone. |
| catalog | Unknown | Partial | catalog data | Later | Not first backbone. |
| storefront | Unknown | Unknown | storefront runtime | Later | Out of MVP recovery. |

Status values:

- Keep
- Stabilize
- Replace
- Archive
- Unknown

## Packages

| Package | Used by | Keep/Rewrite/Delete | Notes |
| --- | --- | --- | --- |
| `pkg/pdtenantdb` | backoffice/runtime | Keep | Must resolve from onboarding route projection. |
| `pkg/messaging` | workers/outbox | Keep | Use for commit-coupled events only where needed. |
| `pkg/toolkit/kvstores` | runtime route projection | Keep | Projection store abstraction. |
| `pkg/api/proto` | gRPC/gateway | Keep | Generated, do not hand edit. |

## APIs

| API | Service | Status | Contract exists? | Notes |
| --- | --- | --- | --- | --- |
| Login/session | auth | Unknown | Partial | Verify first. |
| IAM permission checks | iam | Unknown | Partial | Backend-only authorization path. |
| Store requests | onboarding | Unknown | Partial | Backbone dependency. |
| Placement readiness | onboarding | Missing/partial | No | Needs contract. |
| Backoffice GraphQL protected read | backoffice | Unknown | Partial | Use as first protected API. |

## Data Stores

| Store/Table/Collection | Owner service | Used? | Tenant aware? | Source of truth? | Notes |
| --- | --- | --- | --- | --- | --- |
| auth DB | auth | Yes | user/session scoped | Yes | Verify schemas. |
| iam DB | iam | Yes | org/tenant scoped | Yes | Verify bootstrap permissions. |
| onboarding Mongo | onboarding | Yes | workspace/store/placement scoped | Yes | Placement allocation truth. |
| Mongo runtime KV | projection | Yes | tenant route scoped | No | Rebuildable route projection. |
| backoffice tenant DB/schema | backoffice | Yes | tenant/store scoped | Yes | Resolved through `pdtenantdb`. |

## Broken Or Risky Flows

| Flow | Symptom | Required doc | Next action |
| --- | --- | --- | --- |
| Sign in -> choose store -> open Backoffice | Not reliably working | Backbone requirement/design | Stabilize first. |
| Store provisioning readiness | Missing full readiness contract | Onboarding readiness contract | Define before code. |
| IAM permission UX | Manual IDs/permissions still exist in places | IAM role matrix contract | Later sprint. |
| FE long admin screens | Some screens still long/mounted | UI state specs | Later sprint. |

## Inventory Tasks

- Run Docker stack and record which services start.
- Record first failing request for the backbone flow.
- Map each failing request to owner service.
- Update this file with status and next slice.
