# Frontend Solid Audit

Audit scope: `internal/ui-podzone/src`.

This document records migration work. Stable conventions live in
`agent/SOLID_STYLE_GUIDE.md`.

## Current Strengths

- Feature routes are separated into `shell`, `onboarding`, `iam`, and
  `backoffice`.
- Route components are lazy-loaded and tenant routes have access guards.
- IAM and Orders service facades are split into feature/query/command files.
- Typed form stores replace many groups of field-level signals.
- Feature contexts exist for IAM and Orders subtrees.
- Solid control-flow components are used consistently.
- Event listeners and recurring timers generally register cleanup.
- Shared table, pagination, feedback, field, and layout primitives exist.
- Feature view-model contexts are inferred from typed owner primitives.
- Order detail components read reactive props without destructuring.

## Resolved Findings

- Removed `any` from IAM, Admin Home, and Admin Settings view-model boundaries.
- Removed reactive prop destructuring from the Orders detail component tree.
- Defined frontend verification without requiring unit tests.
- Moved onboarding workspace, tenant access, session, and audit reads to Solid
  resources with resource-owned loading and read errors.
- Migrated Auth sessions and audit logs to the common server collection contract,
  resource `.latest`, bounded scroll regions, and stable pagination.
- Split Admin Settings into namespaced Sessions, Audit, Team Access, Invites,
  Platform Roles, and shared workspace-access ViewModels.
- Namespaced the IAM root model, added explicit principal loading, and rejected
  stale IAM selection responses.
- Moved Partner collection/form orchestration into a feature ViewModel and made
  Backoffice Audit filters explicit instead of fetching on every keystroke.
- Fixed Product Setup status feedback so failed mutations cannot report success.
- Unified frontend formatting under the frontend Prettier configuration.
- Reduced every Backoffice route file to a composition wrapper and moved each
  implementation into its feature folder, matching the Onboarding module
  boundary.
- Split IAM state and remote loaders by organization, policy, group, principal,
  assignment, trust/simulation, and shell ownership. The IAM root now composes
  those feature contracts instead of declaring all signals and reads itself.
- Split IAM mutations by the same feature boundaries; no IAM action owner spans
  unrelated organizations, policies, groups, principals, assignments, and
  trust/simulation workflows.
- Moved Home, Product Setup, Order Audit, and Order Finance remote state,
  mutation lifecycle, stale-request handling, and derived operational data into
  dedicated feature ViewModels.
- Added the common collection contract to Partner REST/gRPC and routed Orders
  GraphQL, including database count, search, field filters, sort whitelists, and
  page metadata. The Partners screen now uses backend pagination directly.
- Added the common collection contract to IAM organizations, managed policies,
  and groups. Their feature resources now own backend search, filters, sort,
  loading, errors, and pagination instead of slicing full arrays in the browser.
- Migrated IAM policy versions and attachments, group members and policies, and
  platform/tenant principal policies to the same server collection contract.
  IAM no longer uses client pagination for operational collections.
- Migrated tenant members, tenant invites, and platform role memberships to the
  common server collection contract. Admin Settings now owns separate paginated
  resources for each collection and renders operational tables instead of
  unbounded cards.
- Migrated onboarding store requests and Backoffice stores to the common server
  collection contract. Store selection and provisioning now use dedicated
  paginated resources and operational tables; dashboard metrics retain an
  explicit compatibility traversal until backend projections replace it.
- Promoted the paginated resource owner to `solid/pagination` so Onboarding and
  Backoffice use the same resource behavior.
- Removed direct IAM permission probes from Onboarding, Admin Settings, and the
  IAM console. UI capabilities now come only from protected business responses;
  backend services remain the authorization boundary.

## P0: Correctness And Safety

### Remote Fetching Is Effect-Driven

Several pages manually combine data, loading, error, and dependency tracking.
Requests can race when filters, principal selection, tenant, or store changes.

Onboarding is migrated. Remaining pages:

- `modules/backoffice/pages/home/TenantHomeView.tsx`
- `modules/backoffice/pages/orders/TenantOrdersPageView.tsx`
- `modules/backoffice/pages/order-audit/TenantOrderAuditView.tsx`
- `modules/backoffice/pages/order-finance/TenantOrderFinanceView.tsx`
- `modules/backoffice/pages/product-setup/TenantProductSetupView.tsx`

Move primary reads to typed router queries/preloads or resource primitives. Add
stale-response and cancellation behavior.

## P1: Architecture And Scale

### Aggregate Metrics Still Traverse Pages

Orders, store attention, and workspace summary consumers intentionally walk all
pages until their aggregate metrics move to backend projections. Operational
tables use backend page/search/filter/sort state directly.

### Remaining Collection Contract Gaps

The active UI still has three non-common read shapes:

- user workspace memberships are returned as one array and feed workspace
  selectors across Onboarding and IAM;
- organization service-control-policy attachments are returned as one nested
  array;
- Product Setup returns draft and candidate arrays in one snapshot.

The routed-order activity feed already has bounded cursor pagination and does
not need offset pagination. Migrate the three array contracts before adding
more consumers; do not replace them with client-side slicing.

### Route State Is Mostly Local

IAM active section uses a manual hash. Partner filters, pagination, Orders queue
state, Audit filters, and Finance selection are not consistently represented by
typed search parameters.

Update router search schemas and use router navigation instead of manual history
or local-only state.

### Internal Navigation Uses Raw Anchors

`Button` renders a raw anchor for `href`, and shared navigation components also
render raw anchors. Internal navigation can trigger full document loads.

Update:

- `solid/components/common/Primitives.tsx`
- `solid/components/common/Navigation.tsx`
- `solid/components/common/Breadcrumbs.tsx`
- `solid/components/common/MegaMenu.tsx`
- `solid/components/common/Overlay.tsx`
- `solid/components/common/display/ListGroup.tsx`
- `modules/shell/pages/auth/DevAuthBootstrapPage.tsx`

Introduce a router-aware link primitive for internal routes.

### Admin Workflows Still Mix Editors And Lists

Tables and pagination were added, but create/edit controls remain permanently
mounted in several workspaces.

Update:

- IAM policies, groups, principals, organizations, role assignments, trust, and
  simulator
- Partners editor/list
- Product setup draft/candidate workflow
- Orders detail editor
- Audit and Finance collection views

Use list plus detail route/drawer and create/edit modal/drawer. Mount only the
active editor.

### Large Owners And Mixed Responsibilities

Highest-priority splits:

- `modules/onboarding/pages/AdminHomePage.tsx`
- `modules/iam/pages/admin-iam/trust-simulation/TrustSimulationPanel.tsx`
- `modules/backoffice/pages/order-audit/TenantOrderAuditView.tsx`
- `modules/backoffice/pages/product-setup/TenantProductSetupView.tsx`
- `modules/backoffice/pages/home/TenantHomeSections.tsx`
- `modules/backoffice/pages/orders/CreateRoutedOrderPanel.tsx`
- `solid/components/common/Primitives.tsx`

Split by workflow or primitive family, not arbitrary line ranges.

### Form Submission Is Inconsistent

Some submit handlers clear submitting state without `finally`, including IAM,
onboarding, product setup, and order paths. A thrown transport error can leave a
form stuck.

Standardize mutation state and backend field-error mapping.

## P2: Consistency And Maintainability

### Unused Or Parallel Dependencies

`@apollo/client` has no source imports while GraphQL uses a custom adapter.
Review and remove it unless a planned migration has an owner. Also review the
scope of Flowbite and the large generic component inventory.

### Browser Storage Is Split Across Layers

Token, tenant, and store adapters exist, but Orders accesses local storage
directly.

Move queue presets and templates behind a typed feature storage adapter. Define
key versioning, tenant/store namespace, and malformed-data behavior.

### IDs And Route Search Types Are Inconsistent

IAM uses numeric user and group IDs while route/search parsing often starts from
`Record<string, unknown>`.

Add typed route search schemas and normalize boundary IDs before they enter
feature state.

### Accessibility Needs Workflow Tests

Tables and controls have basic semantics, but drawer/dialog focus management,
internal link behavior, keyboard table actions, and status announcements are not
tested.

Add component and E2E assertions before introducing more overlays.

## Recommended Migration Order

1. Move IAM and Backoffice page fetch effects to route/resource data.
2. Add server pagination contracts for IAM, Orders, Partners, and onboarding.
3. Move shareable list state into typed route search parameters.
4. Convert remaining card collections and permanent editors to list/detail
   workflows.
5. Add E2E smoke coverage for critical workflows.
6. Unify formatting and remove unused dependencies.
