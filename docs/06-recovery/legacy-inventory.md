# Legacy Inventory

Status: filled 2026-07-10.

Use this to classify the current code before refactoring.

## Services

| Service | Runs in Docker? | Has tests? | Owns data | Status | Notes |
| --- | --- | --- | --- | --- | --- |
| auth | Yes (services.yml) | Yes (17 test files) | Postgres — user/session/token data | Stabilize | JWT verifier, Google OAuth, worker publishes events via Kafka. Verify login + session flow end-to-end. |
| iam | Yes (services.yml) | Yes (11 test files) | Postgres — tenant, membership, policy, role, boundary | Stabilize | gRPC-only (no HTTP). Required by backoffice, onboarding, partner. Bootstrap permissions must be verified. |
| onboarding | Yes (services.yml) | Yes (14 test files) | Mongo — store requests, infrasmanager (placement allocation, resource inventory, route projection) | Stabilize | Core backbone dependency. Provisioning pipeline and placement backbone are the highest-risk area. |
| backoffice | Yes (services.yml) | Yes (18 test files) | Postgres (tenant DB via pdtenantdb) + Mongo (store-scoped runtime) | Stabilize | GraphQL API only. Must not open until store request status = ready AND placement route is live. |
| partner | Yes (services.yml) | Yes (2 test files, partial) | Postgres — partner records | Later | Not needed for first backbone slice. |
| catalog | Yes (services.yml) | No test files | Mongo — catalog data | Later | Not needed for first backbone slice. |
| grpcgateway | Yes (services.yml) | No test files | None | Stabilize | HTTP → gRPC adapter. Backbone entry point for FE API calls. |
| auth-worker | Yes (services.yml) | — | None (event consumer) | Stabilize | Kafka consumer for auth events. Required by iam and backoffice startup. |
| iam-worker | Yes (services.yml) | — | None (event consumer) | Stabilize | Kafka consumer for IAM events. Required by backoffice. |

## Packages

| Package | Used by | Keep/Rewrite/Delete | Notes |
| --- | --- | --- | --- |
| `pkg/pdtenantdb` | backoffice | Keep | Resolves tenant DB route from KV projection published by onboarding. Must not be bypassed. |
| `pkg/messaging` | workers, outbox relay | Keep | Commit-coupled Kafka events. Do not change publisher contracts. |
| `pkg/toolkit/kvstores` | onboarding, backoffice runtime | Keep | Route projection store abstraction (Redis/Valkey). Projection is rebuildable, not source of truth. |
| `pkg/api/proto` | auth, iam, partner, gateway | Keep | Generated protobuf. Do not hand-edit. Regenerate via `make proto`. |
| `pkg/pdauthn` | onboarding, backoffice, partner | Keep | JWT verifier used at service inbound boundaries. |
| `pkg/collection` | onboarding, backoffice, iam | Keep | Shared pagination types. |

## APIs

| API | Service | Status | Contract exists? | Notes |
| --- | --- | --- | --- | --- |
| Login / session (Google OAuth + JWT) | auth | ⚠️ Exists, unverified | Partial (transport-contracts.md) | gRPC proto defined. HTTP exposed via grpcgateway. Verify token flow in Docker. |
| Validate session / introspect token | auth | ⚠️ Exists, unverified | Partial | Used by service-level JWT middleware. |
| Resolve workspace membership | iam | ⚠️ Exists, unverified | Partial | gRPC only. Required by onboarding and backoffice guards. Verify bootstrap membership. |
| IAM permission check (gRPC PDP) | iam | ⚠️ Exists, unverified | Partial | Backend-only authorization. Not called by FE. |
| Store requests CRUD | onboarding | ⚠️ Exists, unverified | Partial | `GET/POST /requests`, `GET /requests/:id`. Returns full status lifecycle. |
| Tenant placement status | onboarding | ⚠️ Exists, unverified | No HTTP contract doc | `GetTenantPlacementStatus` usecase exists. Admin-only HTTP at `/infras/placements/:tenantId/status`. No user-facing readiness contract documented. |
| Combined store readiness check | onboarding | ❌ Missing | No | No single endpoint combines store request status + placement allocation ready + route ready. FE needs this to show ready/blocked/pending states. |
| Backoffice GraphQL protected read | backoffice | ⚠️ Exists, unverified | Partial (schema exists) | GraphQL schema in `internal/backoffice/graphql/`. Verify one query end-to-end with a ready store. |

## Data Stores

| Store/Table/Collection | Owner service | Used? | Tenant aware? | Source of truth? | Notes |
| --- | --- | --- | --- | --- | --- |
| Postgres (auth schema) | auth | Yes | user/session scoped | Yes | JWT keys, sessions, OAuth state. |
| Postgres (iam schema) | iam | Yes | org/tenant/user scoped | Yes | Tenant, membership, policy, role, boundary. Verify bootstrap seeding. |
| Mongo (onboarding DB) | onboarding | Yes | workspace/store scoped | Yes | Store requests, placement allocations, resource inventory, route projections. |
| Redis/Valkey (runtime KV) | onboarding → backoffice | Yes | tenant route scoped | No — rebuildable | Route projection published by onboarding worker. Read by `pkg/pdtenantdb`. |
| Postgres (tenant DB / per-schema) | backoffice | Yes | tenant/store scoped | Yes | Resolved at runtime from KV projection via `pdtenantdb`. |
| Mongo (backoffice store runtime) | backoffice | Yes | store scoped | Yes | Store-local operating data (orders, products, etc.). |

## Broken Or Risky Flows

| Flow | Symptom | Required doc | Next action |
| --- | --- | --- | --- |
| Sign in → choose workspace → open Backoffice | Not verified end-to-end in Docker dev | backbone-flow-refactor.md | Run Docker stack and record first failure point. |
| Store provisioning readiness (FE visible) | No combined readiness endpoint — FE cannot distinguish pending/blocked/ready without two calls | Onboarding readiness contract (Slice 0.3) | Define and implement combined readiness query. |
| IAM bootstrap permissions | Bootstrap creates membership but permission rows may be missing for backoffice API guard | IAM bootstrap task | Verify with a real login after `dev-bootstrap` runs. |
| `pdtenantdb` route resolution | Backoffice cannot open if route projection (KV) is not yet published by onboarding | Placement backbone flow | Verify onboarding publishes route projection before backoffice is opened. |
| partner-service depends on iam+auth startup | Docker `depends_on` order may cause race before IAM bootstraps permissions | services.yml startup | Monitor logs after `docker compose --profile backoffice up`. |

## Inventory Tasks

- [x] Record services from Docker services.yml and test file counts.
- [ ] Run `docker compose -f infras.yml -f services.yml --profile backoffice up` and record first failing step.
- [ ] Verify `dev-bootstrap` completes and seeds correct IAM permissions.
- [ ] Call `GET /requests` after bootstrap and confirm store request status reaches `ready`.
- [ ] Verify KV projection is published and backoffice can resolve tenant DB.
- [ ] Call one protected backoffice GraphQL query and confirm auth guard works.
