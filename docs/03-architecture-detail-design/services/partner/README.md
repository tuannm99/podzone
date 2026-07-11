# Component: Partner Service

Parent index: [Services](../README.md).

Status: implemented, tagged **Later** in
[`docs/06-recovery/legacy-inventory.md`](../../../06-recovery/legacy-inventory.md)
— not needed for the current backbone recovery slice. This doc describes
current implemented state, not active development priority.

## Purpose

Own the store-scoped external business partner record — a print-on-demand,
production, or fulfillment supplier a tenant works with. Historical name in
code/migrations is "supplier"; product framing renamed it to "partner" (see
[`docs/00-project-vision/07-partner-mvp.md`](../../../00-project-vision/07-partner-mvp.md)).
This MVP is the minimum business object later flows (product feeds,
fulfillment routing, shipment tracking, settlement — see
[`docs/00-project-vision/08-partner-refactor-plan.md`](../../../00-project-vision/08-partner-refactor-plan.md))
attach to; none of those flows are implemented yet.

## Responsibilities

- CRUD for partner records: create, get, list, update, activate/deactivate.
- Enforce tenant-scoped uniqueness on partner `code`.
- Normalize partner type (`print_on_demand` | `fulfillment` |
  `dropship_supplier`), status (`active` | `inactive`), capability lists
  (supported product types/regions), and shipping cost rules.
- Authorize every call against the caller's session + tenant membership +
  IAM permission (`partner:read` / `partner:manage`) before touching data.

## Non-Responsibilities

- No product feed, product setup, fulfillment routing, shipment tracking,
  or settlement logic — out of scope per the MVP doc above.
- No frontend surface — see "Frontend Surface" below.
- Does not own session/JWT issuance or IAM permission evaluation — it is a
  caller of both (`auth` and `iam` gRPC), not the source of truth for
  either.

## Owned Data

| Data / Table / Resource | Access | Notes |
|---|---|---|
| `partners` (Postgres) | read/write | See [DB Design](./db-design.md). Single shared database, `tenant_id` is a plain scoping column — not a per-tenant routed schema like `pkg/pdtenantdb`. |

## Interfaces

### Inbound APIs

gRPC only (`pkg/api/proto/partner/v1`), no HTTP/GraphQL of its own — exposed
over HTTP via `internal/grpcgateway` (`registrar_partner.go`).

| API | Contract | Caller | Notes |
|---|---|---|---|
| `CreatePartner` | `pbpartnerv1.CreatePartnerRequest` → `Partner` | backoffice, gateway | Requires `partner:manage`. |
| `GetPartner` | `GetPartnerRequest` → `Partner` | backoffice, gateway | Requires `partner:read`, checked against the fetched record's `tenant_id` (not the caller-supplied one). |
| `ListPartners` | `ListPartnersRequest` → `ListPartnersResponse` | backoffice, gateway | Requires `partner:read`. Paginated via `pkg/collection`. |
| `UpdatePartner` | `UpdatePartnerRequest` → `Partner` | backoffice, gateway | Requires `partner:manage`. |
| `UpdatePartnerStatus` | `UpdatePartnerStatusRequest` → `Partner` | backoffice, gateway | Requires `partner:manage`. Activate/deactivate. |

### Outbound Calls

| Target | Protocol | Reason | Notes |
|---|---|---|---|
| `auth` service | gRPC (`pbauthv1.AuthServiceClient`) | Validate session (`GetSession`) as part of every authorized call. | Connection built once in `NewTenantAuthorizer`, config via `partner.auth.grpc_host`/`grpc_port`. |
| `iam` service | gRPC (`pbiamv1.IAMQueryServiceClient`) | Validate tenant membership (`GetTenantMembership`) and evaluate permission (`CheckPermission`). | Config via `partner.iam.grpc_host`/`grpc_port`, defaults to same host as auth, port `50053`. |

## Dependencies

| Dependency | Type | Reason |
|---|---|---|
| Postgres | DB | `partners` table, via `pkg/pdsql` + `sqlx`, migrated with `goose` (`internal/partner/migrations`). |
| auth service | gRPC | Session validation on every authorized request. |
| iam service | gRPC | Membership + permission check on every authorized request. |

## Runtime Flows

Every inbound call goes through the same tenant-authorization flow
(`internal/partner/controller/grpchandler/authz.go`,
`authTenantAuthorizer.AuthorizeTenant`) — see
[api-design.md](./api-design.md) "C4: Sequences Per Usecase" for the
sequence diagram. `GetPartner`/`UpdatePartner`/`UpdatePartnerStatus` fetch
the record by ID *before* authorizing, then authorize against the record's
actual `tenant_id` — the caller cannot claim a different tenant to bypass
the check (see `partner_server.go`).

## Failure Modes

| Failure | Expected Behavior |
|---|---|
| Missing/invalid JWT, wrong signing key | `PermissionDenied` (surfaced as authorization failure, not a distinct auth error to the caller). |
| Session inactive / revoked | `PermissionDenied`. |
| Tenant membership inactive | `PermissionDenied`. |
| Permission check denied | `PermissionDenied`. |
| Partner not found | `NotFound` (`ErrPartnerNotFound`). |
| Duplicate `(tenant_id, code)` | `AlreadyExists` (`ErrPartnerCodeTaken`) — enforced by the DB unique constraint, surfaced through the repository. |
| Invalid input (empty name/code/tenant, bad type/status) | `InvalidArgument`. |
| Auth/IAM gRPC unreachable | Falls through to `Internal` via `mapAuthzError`/default case — not distinguished from other backend faults today. |

## Security

- Authentication: Bearer JWT, HS256, shared secret from
  `partner.auth.jwt_secret` (same secret family as `auth`/`iam`/`backoffice`
  — see the JWT v3→v5 migration note in `docs/08-adr/` if one exists for
  that change, otherwise `pkg/pdauthn`).
- Authorization: `partner:read` / `partner:manage` permission strings,
  evaluated by IAM, not locally — this service never decides permissions
  itself.
- Permission: two granularities only (`read`, `manage`) — no finer action
  split (e.g. no separate `create` vs `update` vs `deactivate`
  permission).
- Tenant/workspace/store isolation: `tenant_id` column-scoped; every
  mutating/reading call is authorized against the tenant the record
  actually belongs to.
- Sensitive data: `contact_email` is PII; no field is currently flagged or
  redacted in logs — treat as sensitive if adding logging.

## Observability

- Logs: `pkg/pdlog`, standard fx-wired logger (`pdlog.Logger`), no
  partner-specific structured fields added beyond default request/error
  logging.
- Metrics: none added.
- Traces: none added.
- Alerts: none added.

## Config

`internal/partner/config` (`partner.*` keys via `pkg/pdconfig`/koanf):

| Key | Purpose | Default |
|---|---|---|
| `partner.auth.jwt_secret` | JWT verification secret | required, no default |
| `partner.auth.jwt_key` | Expected `key` claim value | — |
| `partner.auth.grpc_host` | Auth service host | `localhost` |
| `partner.auth.grpc_port` | Auth service port | `50051` |
| `partner.iam.jwt_secret` / `jwt_key` | (present, unused directly — IAM calls go through the auth-issued token, not a separate IAM token) | — |
| `partner.iam.grpc_host` | IAM service host | same as `auth.grpc_host` |
| `partner.iam.grpc_port` | IAM service port | `50053` |

Per `docs/00-governance/twelve-factor.md`: all of the above come from
environment/config, no hardcoded host/port in code except the listed
defaults, which are dev-only fallbacks.

## Frontend Surface

**None.** No dedicated MFE remote exists (`frontend/apps/` only has `iam`,
`backoffice`, `onboarding`). Partner records are currently managed only via
direct gRPC/HTTP-gateway calls. The only in-repo caller is
`internal/backoffice/infrastructure/partnerdirectory/adapter.go` — backoffice
calls this service as a backend dependency (partner directory lookups for
routing/fulfillment features), not a UI. If a partner-management UI is
built, per `docs/03-architecture-detail-design/14-mfe-federation-contract.md`
it would either be a new MFE remote or a section inside the `backoffice`
remote — no decision has been made (would need a PZEP/ADR).

## Agent Rules

- Do not put business logic in `controller/grpchandler` — it only maps
  transport and calls the usecase (see `partner_server.go`).
- Do not bypass `authz.go`'s `TenantAuthorizer` — every mutating and most
  read calls must go through it.
- Do not add a new permission string without updating both this service's
  calls to `AuthorizeTenant` and the IAM permission catalog
  (`docs/03-architecture-detail-design/services/iam/`).
- Do not change public proto contracts without a PZEP (per
  `docs/00-governance/agent-working-rule.md`).
