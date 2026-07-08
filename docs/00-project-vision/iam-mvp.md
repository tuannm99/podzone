# IAM MVP

The repository now runs IAM as a separate gRPC service from auth, while both still share the same auth database for now.

## Service split

- `AuthService`
  - login, register, Google OAuth login
  - refresh token
  - logout
  - session lookup and revoke
  - active tenant switch
  - audit log listing for the authenticated actor
- `IAMService`
  - create tenant
  - tenant membership management
  - tenant permission checks
  - platform role management
  - platform permission checks

This is a deployment split first. Auth still depends on IAM domain logic internally for flows such as tenant bootstrap and active-tenant switching, but IAM RPCs are no longer served by `auth-service`.

## Core model

- `users`: existing principals for human identities
- `tenants`: tenant/workspace/account records
- `iam_roles`: system roles such as `tenant_owner`, `tenant_admin`, `tenant_editor`, `tenant_viewer`, `platform_owner`, `platform_admin`
- `iam_permissions`: normalized permission namespace such as `store:read`
- `iam_role_permissions`: role-to-permission mapping
- `tenant_memberships`: user membership inside a tenant with a role
- `user_platform_roles`: platform-wide role bindings
- `auth_sessions`: session lifecycle and active tenant context
- `auth_refresh_tokens`: refresh token rotation state
- `auth_audit_logs`: audit trail for sensitive auth and IAM actions

## Current HTTP surface

### AuthService via grpc-gateway

- `POST /auth/v1/login`
- `POST /auth/v1/register`
- `GET /auth/v1/google/login`
- `GET /auth/v1/google/callback`
- `POST /auth/v1/google/exchange`
- `POST /auth/v1/refresh`
- `POST /auth/v1/logout`
- `POST /auth/v1/iam/tenants:switch`
- `GET /auth/v1/sessions`
- `GET /auth/v1/sessions/{session_id}`
- `DELETE /auth/v1/sessions/{session_id}`
- `GET /auth/v1/audit-logs`

### IAMService via grpc-gateway

- `POST /auth/v1/iam/tenants`
- `GET /auth/v1/iam/users/{user_id}/tenants`
- `GET /auth/v1/iam/tenants/{tenant_id}/members`
- `POST /auth/v1/iam/tenants/{tenant_id}/members`
- `DELETE /auth/v1/iam/tenants/{tenant_id}/members/{user_id}`
- `GET /auth/v1/iam/tenants/{tenant_id}/members/{user_id}`
- `POST /auth/v1/iam/permissions:check`
- `POST /auth/v1/iam/platform-permissions:check`
- `GET /auth/v1/iam/platform-users/{target_user_id}/roles`
- `POST /auth/v1/iam/platform-users/{target_user_id}/roles`
- `DELETE /auth/v1/iam/platform-users/{target_user_id}/roles/{role_name}`

## Current runtime expectations

- frontend and backoffice must use bearer tokens
- tenant context comes from `active_tenant_id` in the access token
- `backoffice` uses `AuthService` for session validation and `IAMService` for membership/permission checks
- tenant and platform admin APIs derive actor identity from the JWT, not from request body fields

## Extension path

1. Finish the code-level split between auth domain and IAM domain, not just RPC/deployment split.
2. Add invite flow and email-based tenant onboarding.
3. Add custom roles per tenant.
4. Add service accounts and inter-service auth.
5. Add policy statements and conditions for ABAC-style evaluation.
6. Add richer audit browsing and export.
