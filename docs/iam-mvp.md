# IAM MVP

This repository now contains an IAM MVP scaffold embedded in the auth service database and process.

## Core model

- `users`: existing principals for human identities
- `tenants`: tenant/workspace/account records
- `iam_roles`: system roles such as `tenant_owner`, `tenant_admin`, `tenant_editor`, `tenant_viewer`
- `iam_permissions`: normalized permission namespace such as `store:read`
- `iam_role_permissions`: role-to-permission mapping
- `tenant_memberships`: user membership inside a tenant with a role

## Current capabilities

- create tenant with owner membership
- add or update member role in a tenant
- resolve membership by `(tenant_id, user_id)`
- check or require permission by `(tenant_id, user_id, permission)`
- expose IAM MVP over `AuthService` gRPC + grpc-gateway HTTP:
  - `POST /auth/v1/iam/tenants`
  - `POST /auth/v1/iam/tenants/{tenant_id}/members`
  - `GET /auth/v1/iam/tenants/{tenant_id}/members/{user_id}`
  - `POST /auth/v1/iam/permissions:check`
  - `POST /auth/v1/iam/tenants:switch`

## Extension path

The MVP is intentionally shaped so future work can extend without breaking tables:

1. Add service accounts as another principal type
2. Add custom roles per tenant
3. Add policy statements and conditions for ABAC-style evaluation
4. Move backoffice authorization from local JWT-only checks to IAM membership + permission checks
5. Add token/session model that carries active tenant context safely

## Suggested next build steps

1. Integrate clients/UI with `tenants:switch` so they stop relying on `X-Tenant-ID` fallback
2. Add refresh tokens and session revocation
3. Add service accounts and inter-service auth
4. Add custom roles and policy conditions
