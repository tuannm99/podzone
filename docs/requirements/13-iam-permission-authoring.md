# IAM Permission Authoring And Matrix

Status: policy permission matrix implemented; role matrix contract proposed.

## Problem

IAM administrators currently have to know permission names, role names, and
principal IDs before assigning access. Raw JSON is useful for advanced policy
work, but it cannot be the default authoring experience.

The product must provide:

- a searchable permission catalog owned by IAM;
- feature-oriented permission selection;
- searchable user, group, policy, and role selectors;
- a visual role-permission matrix;
- an advanced JSON editor without creating a second policy format.

The UI must never hardcode permissions as authorization truth. It renders the
catalog and protected resources returned by IAM; backend guards remain the
security boundary.

## Domain Language

| Term                | Meaning                                                          |
| ------------------- | ---------------------------------------------------------------- |
| Feature             | Human-facing grouping backed by a permission resource            |
| Permission          | Registered action identified by a stable namespaced name         |
| Role                | Named, scope-bound permission set                                |
| Policy statement    | Allow or deny rule with action, resource, and conditions         |
| Managed policy      | Versioned reusable policy                                        |
| Inline policy       | Policy owned by one principal or group                           |
| Permission boundary | Maximum permissions a principal or role may receive              |
| Matrix selection    | Simple allow statement for one exact permission on all resources |
| Scoped selection    | Exact permission with deny, resource scope, or conditions        |

## Current Implementation

The IAM API already exposes a paginated permission catalog containing:

- `id`
- `name`
- `resource`
- `action`

Policy forms now provide three synchronized authoring modes:

1. **Permission matrix** groups permissions by feature/resource and supports
   search, single selection, and feature-level selection.
2. **Builder** owns effect, exact or wildcard action, resource patterns, and
   conditions.
3. **JSON** preserves the wire-compatible statement array for advanced edits.

Matrix changes create or remove only exact permission statements. They never
silently remove wildcard statements. A permission with a deny rule, resource
scope, or conditions is marked as scoped and remains editable in Builder.
Invalid JSON remains visible as a draft and must not clear the parsed policy.

## Functional Requirements

### FR-1 Permission Catalog

- IAM is the source of truth for permission metadata.
- Catalog queries support scope, page, page size, search, filters, and sort.
- Search covers permission name, resource, action, and optional display label.
- The UI must fetch every catalog page needed by the matrix; a page-size limit
  must never silently hide permissions.
- Unknown or inactive permissions remain visible on existing policies as
  unresolved values so administrators can repair them.
- A future generic IAM platform registers application permissions through the
  application schema registry, not Podzone migrations.

### FR-2 Policy Permission Matrix

- Rows represent features/resources.
- Cells represent actions and display the stable permission name.
- Administrators can search, select one permission, or select all visible
  permissions in one feature.
- Selection creates `allow + exact permission + resource * + no condition`.
- Deselection removes exact statements for that permission only.
- Scoped, denied, conditional, and wildcard statements are identified but
  edited through Builder or JSON.
- Matrix, Builder, and JSON use one form value and remain synchronized.

### FR-3 Role Catalog

IAM must replace frontend role constants with:

```text
ListRoles(scope, organization?, tenant?, collection)
GetRole(role_id)
```

Each role read model returns:

- stable role ID and name;
- display name and description;
- scope and owning organization/application;
- system/custom status;
- revision;
- assigned principal count;
- permission count.

All list operations use the common collection contract.

### FR-4 Role-Permission Queries

IAM must expose:

```text
GetRolePermissionSet(role_id) -> role, revision, permission_ids
ListRolePermissionSets(role_ids, permission_ids?) -> bounded matrix slice
```

The batch query prevents one request per matrix cell. The response includes the
catalog revision and role revisions used to render the matrix.

### FR-5 Role-Permission Commands

IAM must expose:

```text
ReplaceRolePermissionSet(
  role_id,
  permission_ids,
  expected_revision,
  reason
) -> new_revision
```

- The command is atomic.
- `expected_revision` enforces optimistic concurrency.
- The backend validates role scope and permission scope.
- System roles may be immutable or require a dedicated system-admin action.
- An empty permission set requires explicit confirmation.
- Every change writes an audit record and an outbox event in the same
  transaction.

Incremental grant/revoke commands may be added for automation, but the UI uses
replace-with-revision to avoid partial matrix saves.

### FR-6 Role-Permission Matrix

- The administrator selects one scope before loading roles.
- Rows are features and permissions; columns are roles.
- Search and filters apply to role, feature, action, and permission name.
- Column and row bulk selection are explicit commands.
- Unsaved cells are visually distinct from persisted cells.
- Save shows a change summary: granted, revoked, affected roles, and reason.
- Conflict responses reload current revisions without losing the local draft.
- Large matrices use server-side role pagination and row virtualization.
- The screen provides read-only mode when the caller can view but not manage.

### FR-7 Principal Assignment

- User, group, policy, and role inputs are directory-backed selectors.
- IDs remain visible as secondary diagnostic text, not primary input.
- Role assignment uses the role catalog, never frontend constants.
- Permission denial stays within the current screen and identifies the missing
  permission and resource.

## Authorization And Audit

Required management actions should be separated:

| Operation                    | Proposed action           |
| ---------------------------- | ------------------------- |
| View catalog and matrix      | `iam:read_permissions`    |
| Create or edit custom role   | `iam:manage_roles`        |
| Edit system role             | `iam:manage_system_roles` |
| Assign role to principal     | `iam:assign_roles`        |
| Edit managed/inline policy   | `iam:manage_policies`     |
| View detailed policy explain | `iam:explain_access`      |

The exact names are registered in the IAM application schema. Each command
audit record contains actor, organization/application, target, before/after
revision, permission delta, reason, request ID, and timestamp.

## Non-Functional Requirements

- Default deny and fail closed on unknown permissions.
- No N+1 requests for users, roles, or matrix cells.
- Permission catalog reads should be cacheable by revision and ETag.
- A 500-permission by 50-role matrix must remain interactive.
- Keyboard navigation, labels, focus, and non-color status are required.
- JSON parsing and policy validation errors must not destroy the draft.
- Management writes are strongly consistent; decision projections expose their
  applied revision for read-after-write checks.
- No browser-side permission-check probes are allowed.

## Delivery Plan

### Phase 1: Policy Authoring

- Load the IAM permission catalog.
- Add matrix, Builder, and JSON modes with one synchronized statement value.
- Replace raw user and policy IDs with directory-backed selectors.
- Preserve protected-screen UI for expected authorization errors.

Exit: administrators can author exact permissions without memorizing strings.

### Phase 2: Role Contract

- Add role list/detail and role-permission command/query contracts.
- Add repository methods and optimistic revision storage.
- Add guards, audit records, outbox events, and contract tests.
- Migrate system roles from frontend constants to IAM responses.

Exit: no role or permission assignment depends on hardcoded UI data.

### Phase 3: Role Matrix

- Add paginated role columns and virtualized permission rows.
- Add draft, delta preview, reason, atomic save, and conflict recovery.
- Add read-only and missing-permission states.
- Add desktop/mobile interaction smoke tests.

Exit: role permissions can be reviewed and changed from one auditable screen.

### Phase 4: Generic IAM Platform

- Move permission registration to application schemas.
- Replace Podzone-specific scope and identity assumptions with organization,
  application, opaque principal, action, and resource contracts.
- Add SDK and gateway adapters without moving enforcement into the frontend.

Exit: another product can register a catalog and use the same authoring flow.

## Acceptance Criteria

- No primary IAM form asks an administrator to remember a permission or user ID.
- All 18 current Podzone permissions are visible under their six resources.
- Selecting a matrix permission produces the same statement accepted by the
  existing policy API.
- Builder and JSON edits are reflected when returning to the matrix.
- Invalid JSON shows an inline error and does not replace the current policy
  with an empty statement list.
- Permission-denied responses identify the required permission and keep the
  current IAM screen mounted.
- The future role matrix cannot ship until role catalog, batch query, revision,
  audit, and atomic replace contracts are implemented.
