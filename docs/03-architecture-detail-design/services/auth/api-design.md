# Auth Service — API Design

Parent: [Services Index](../README.md) · [auth README](./README.md) · [DB Design](./db-design.md)

## C3: Component View

```mermaid
flowchart TB
    subgraph controller["controller/grpchandler"]
        AuthServer["AuthServer (AuthService gRPC)"]
    end

    subgraph domain["domain (authInteractorImpl)"]
        Credentials["auth_credentials.go\nLogin/Register/RefreshAccessToken/SwitchActiveTenant/Logout"]
        OAuth["auth_oauth.go\nGoogle OAuth code flow"]
        AssumeRole["auth_assume_role.go\nAssumeRole/ClearAssumedRole"]
        SessionPolicy["auth_session.go\nAssumeSessionPolicy/ClearSessionPolicy"]
        TokenUC["token_interactor.go\nJWT issuance (pkg/pdauthn)"]
        UserUC["user_interactor.go\nuser directory usecases"]
    end

    subgraph infra["infrastructure"]
        UserRepo["repository (users, sessions, refresh tokens, audit)"]
        OauthStateRepo["oauth_state_repo_impl.go (Redis)"]
        GoogleClient["oauth_external_impl.go (Google OAuth HTTPS)"]
        IAMClient["iamclient (AccountBootstrapper, RoleAssumer, TenantAccessChecker)"]
    end

    subgraph events["controller/eventhandler/iamprojection"]
        Projection["Kafka consumer: tenant.created, tenant.member.added"]
    end

    AuthServer --> Credentials
    AuthServer --> OAuth
    AuthServer --> AssumeRole
    AuthServer --> SessionPolicy
    Credentials --> UserRepo
    Credentials --> TokenUC
    Credentials --> IAMClient
    OAuth --> GoogleClient
    OAuth --> OauthStateRepo
    OAuth --> UserRepo
    AssumeRole --> IAMClient
    AssumeRole --> UserRepo
    Projection --> UserRepo
    TokenUC --> UserRepo
```

Two binaries build from this codebase: `cmd/auth` (gRPC API, everything
above except `events`) and `cmd/auth-worker` (`events` only, no inbound
gRPC) — see `01-modules.md`.

## gRPC API Surface

Service `auth.v1.AuthService` (`api/proto/auth/v1/{auth,auth_session}.proto`),
also exposed over HTTP via `grpcgateway` (`google.api.http` annotations
below). 19 RPCs.

| RPC | HTTP | Request | Response | Errors |
|---|---|---|---|---|
| `Login` | `POST /auth/v1/login` | `username, password` | `jwt_token, user_info, refresh_token` | `Unauthenticated` (wrong password/not found) |
| `Register` | `POST /auth/v1/register` | `username, password, email` | same as `Login` | `InvalidArgument` (duplicate username/email) |
| `GoogleLogin` | `GET /auth/v1/google/login` | `redirect_after_login` | `redirect_url` | — |
| `GoogleCallback` | `GET /auth/v1/google/callback` | `state, code` | `exchange_code, redirect_url, user_info` | `Internal` (OAuth exchange failure) |
| `ExchangeGoogleLogin` | `POST /auth/v1/google/exchange` | `exchange_code` | same as `Login` | `Unauthenticated`/`Internal` |
| `RefreshToken` | `POST /auth/v1/refresh` | `refresh_token` | same as `Login` | `Unauthenticated` (`ErrRefreshTokenInvalid`/`ErrRefreshTokenExpired`/`ErrSessionRevoked`) |
| `Logout` | `POST /auth/v1/logout` | `token` | `success, redirect_url` | — |
| `SwitchActiveTenant` | `POST /auth/v1/iam/tenants:switch` | `user_id, tenant_id, access_token` | `jwt_token, user_info, refresh_token` | `InvalidArgument` (`ErrInvalidUserID`), `Unauthenticated` (`ErrSessionRevoked`) |
| `AssumeSessionPolicy` | `POST /auth/v1/sessions:assume-policy` | `access_token, statements` | `jwt_token, session` | `InvalidArgument` (`ErrInvalidSessionPolicy`) |
| `ClearSessionPolicy` | `POST /auth/v1/sessions:clear-policy` | `access_token` | `jwt_token, session` | — |
| `AssumeRole` | `POST /auth/v1/sessions:assume-role` | `access_token, role_name, tenant_id, session_policy, external_id, session_name, source_identity, duration_seconds, service_principal, session_tags` | `jwt_token, session` | `InvalidArgument`, propagates IAM `RoleAssumer` errors |
| `ClearAssumedRole` | `POST /auth/v1/sessions:clear-assumed-role` | `access_token` | `jwt_token, session` | `InvalidArgument` (`ErrInvalidUserID`) |
| `GetSession` | `GET /auth/v1/sessions/{session_id}` | `session_id` | `session` | `NotFound` |
| `ListSessions` | `GET /auth/v1/sessions` | `collection` (pagination) | `sessions, page_info` | — |
| `RevokeSession` | `DELETE /auth/v1/sessions/{session_id}` | `session_id` | `{}` | `NotFound` |
| `ListAuditLogs` | `GET /auth/v1/audit-logs` | `page_size` (legacy), `collection` | `logs, page_info` | — |
| `GetUserByIdentity` | `GET /auth/v1/users:by-identity` | `identity` | `user_info` | `NotFound` |
| `EnsureUserByEmail` | `POST /auth/v1/users:ensure-by-email` | `email` | `user_info, created` | — |
| `GetUserByID` | `GET /auth/v1/users/{user_id}` | `user_id` | `user_info` | `NotFound` |
| `ListUsers` | internal only, no HTTP annotation | `collection` | `users, page_info` | — (see proto comment: "Internal directory query. Public IAM management APIs authorize callers before proxying directory results.") |

Error mapping (`authStatusError` in `auth_server.go`): `ErrUserNotFound` →
`NotFound`; `ErrWrongPassword` → `Unauthenticated`;
`ErrUserAlreadyExists`/`ErrUsernameExisted`/`ErrEmailExisted`/
`ErrInvalidSessionPolicy`/`ErrInvalidUserID` → `InvalidArgument`;
`ErrSessionNotFound`/`ErrSessionRevoked`/`ErrRefreshTokenInvalid`/
`ErrRefreshTokenExpired` → `Unauthenticated`; everything else → `Internal`.

Kafka inbound (not gRPC): `tenant.created`, `tenant.member.added` from
IAM — see [IAM service API design](../iam/api-design.md) "IAM Event
Projected into Auth" for the producer-side sequence (this doc doesn't
duplicate it).

## C4: Sequences Per Usecase

### Login (username/password)

```mermaid
sequenceDiagram
    participant FE as Frontend (HOST shell)
    participant Auth as Auth Service
    participant AuthDB as Auth DB
    participant IAM as IAM Service

    FE->>Auth: Login(username, password)
    Auth->>AuthDB: GetByUsernameOrEmail
    Auth->>Auth: CheckPassword (bcrypt compare)
    Auth->>AuthDB: create auth_session + refresh_token
    Auth->>Auth: issue JWT
    alt user.InitialFrom == "podzone" (first login)
        Auth->>IAM: EnsureRootOrganization(userID, username)
    end
    Auth-->>FE: jwt_token, refresh_token, user_info
```

### Register

```mermaid
sequenceDiagram
    participant FE as Frontend (HOST shell)
    participant Auth as Auth Service
    participant AuthDB as Auth DB
    participant IAM as IAM Service

    FE->>Auth: Register(username, password, email)
    Auth->>AuthDB: Create user
    Auth->>AuthDB: UpdateById(InitialFrom = "podzone")
    Auth->>AuthDB: create auth_session + refresh_token
    Auth->>Auth: issue JWT
    Auth->>IAM: EnsureRootOrganization(userID, username)
    Auth-->>FE: jwt_token, refresh_token, user_info
```

### Switch Active Tenant

```mermaid
sequenceDiagram
    participant UI as UI
    participant Auth as Auth Service
    participant LocalProj as Auth IAM Projection
    participant IAM as IAM Service
    participant AuthDB as Auth DB

    UI->>Auth: SwitchActiveTenant(userID, tenantID, accessToken)
    Auth->>LocalProj: EnsureActiveMembership(tenantID, userID)
    alt projection hit
        LocalProj-->>Auth: membership (checked active)
    else projection miss (no local row)
        Auth->>IAM: GetTenantMembership(tenantID, userID)
        IAM-->>Auth: membership (checked active) or NotFound
    end
    Auth->>AuthDB: GetByID(userID), load session from access token
    Auth->>Auth: validate session ownership, active, not expired
    Auth->>AuthDB: UpdateActiveTenant(sessionID, tenantID)
    Auth->>Auth: issue JWT for updated session
    Auth-->>UI: new access token
```

### Refresh Token (rotation)

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant Auth as Auth Service
    participant AuthDB as Auth DB

    FE->>Auth: RefreshToken(refresh_token)
    Auth->>Auth: hash token (SHA-256)
    Auth->>AuthDB: GetByTokenHash
    Auth->>Auth: check not revoked, not expired
    Auth->>AuthDB: GetByID(session), validate active
    Auth->>AuthDB: GetByID(user)
    Auth->>Auth: mint new refresh token
    Auth->>AuthDB: Revoke(old token, replaced_by_token_id = new)
    Auth->>AuthDB: Create(new refresh token)
    Auth->>Auth: issue new access token
    Auth-->>FE: new jwt_token, new refresh_token, user_info
```

Reused (raw) token is a rotation chain — a replayed/reused old refresh
token is rejected once `replaced_by_token_id` marks it superseded (see
[DB Design](./db-design.md) `auth_refresh_tokens`).

### Logout

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant Auth as Auth Service
    participant AuthDB as Auth DB

    FE->>Auth: Logout(access_token)
    Auth->>Auth: resolve session from access token
    Auth->>AuthDB: Revoke(session)
    Auth->>AuthDB: RevokeBySession(all refresh tokens for session)
    Auth-->>FE: success, redirect_url
```

### Assume Role

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant Auth as Auth Service
    participant AuthDB as Auth DB
    participant IAM as IAM Service (via iamclient.RoleAssumer)

    FE->>Auth: AssumeRole(access_token, role_name, tenant_id, ...)
    Auth->>Auth: loadOwnedActiveSession(userID, accessToken)
    Auth->>IAM: AssumeRole(userID, roleName, tenantID, sessionTags, ...)
    IAM-->>Auth: assumedRole { roleID, roleScope, tenantID, expiresAt, sessionTags }
    Auth->>Auth: apply assumed-role fields to session (+ active_tenant_id if role scope is tenant)
    Auth->>AuthDB: UpdateSessionPolicy
    Auth->>AuthDB: UpdateAssumedRole
    Auth->>Auth: issue JWT reflecting assumed role
    Auth-->>FE: new jwt_token, session
```

`ClearAssumedRole` is the inverse: same session load, zero out all
`assumed_role_*` fields + `session_tags`, `UpdateAssumedRole`, re-issue
JWT. No call to IAM on clear.

### Google OAuth Login

See [auth README](./README.md) "Runtime Flows" — the full
`GoogleLogin -> GoogleCallback -> ExchangeGoogleLogin` sequence lives
there already; not duplicated here.

## Cross-Service Dependencies

**Auth calls out to:**
- IAM (`iamclient`, gRPC): `AccountBootstrapper.EnsureRootOrganization`
  (first login/register), `RoleAssumer.AssumeRole` (role assumption).
  `TenantAccessChecker.EnsureActiveMembership` is called by
  `SwitchActiveTenant` — implementation reads the local IAM projection
  tables (`iam_tenants_projection`, `iam_tenant_memberships_projection`),
  not a live IAM gRPC call, per [DB Design](./db-design.md).
- Google OAuth (HTTPS): code exchange, user info fetch.
- Redis (`redis-auth`): OAuth CSRF state, TTL-bound.

**Auth is called by:**
- Frontend (HOST shell), via `grpcgateway`/APISIX, for all user-facing
  RPCs above.
- Other services for internal directory/session lookups: `partner`
  calls `GetSession` for authz (`internal/partner/controller/grpchandler/authz.go`);
  IAM and other services may call `GetUserByIdentity`/`EnsureUserByEmail`/
  `GetUserByID`/`ListUsers` for directory resolution.
- IAM (indirectly, via Kafka): `tenant.created`/`tenant.member.added`
  events consumed by `cmd/auth-worker` to maintain the local projection —
  see [IAM API design](../iam/api-design.md).
