import { Routes } from '@angular/router';
import { AdminHomePage } from './features/onboarding/admin-home-page/admin-home-page.component';
import { StoreChooserPage } from './features/onboarding/store-chooser-page/store-chooser-page.component';
import { LoginPage } from './features/auth/login-page/login-page.component';
import { RegisterPage } from './features/auth/register-page/register-page.component';
import { TenantHomePage } from './features/backoffice/tenant-home-page/tenant-home-page.component';
import { StoreListPage } from './features/backoffice/store-list-page/store-list-page.component';
import { RoutedOrderListPage } from './features/backoffice/routed-order-list-page/routed-order-list-page.component';
import { Shell } from './shell/shell.component';
import { authGuard } from './core/auth/auth.guard';
import { platformAdminGuard } from './core/auth/platform-admin.guard';
import { AdminShell } from './features/admin/admin-shell/admin-shell.component';
import { OnboardingReviewPage } from './features/admin/onboarding-review-page/onboarding-review-page.component';
import { ReadinessDetailPage } from './features/admin/onboarding-review-page/readiness-detail-page/readiness-detail-page.component';
import { OrganizationsPage } from './features/admin/iam/organizations/organizations-page.component';
import { PlatformRolesPage } from './features/admin/iam/platform-roles/platform-roles-page.component';
import { PoliciesPage } from './features/admin/iam/policies/policies-page.component';

// Workspace/store selection are bare screens (no Shell nav/topbar) — they
// run before there's a store context for the Shell's sidebar to be about.
// Backoffice routes (past the handoff, PZEP-0008 M3) are nested under
// Shell — see PZEP-0013.
export const routes: Routes = [
  { path: 'login', component: LoginPage },
  { path: 'register', component: RegisterPage },
  { path: '', component: AdminHomePage, canActivate: [authGuard] },
  { path: 't/:tenantId/stores', component: StoreChooserPage, canActivate: [authGuard] },
  {
    path: 't/:tenantId',
    component: Shell,
    canActivate: [authGuard],
    children: [
      // Minimal Backoffice shell — see PZEP-0011.
      { path: 'backoffice', component: TenantHomePage },
      // First declarative-collection screen — see ADR-0008/PZEP-0012.
      { path: 'backoffice/stores', component: StoreListPage },
      // Second declarative-collection screen — see PZEP-0014.
      { path: 'backoffice/orders', component: RoutedOrderListPage },
    ],
  },
  // Platform admin console — sibling of t/:tenantId, not nested under it:
  // /admin is not tenant-scoped. See PZEP-0020.
  {
    path: 'admin',
    component: AdminShell,
    canActivate: [authGuard],
    children: [
      { path: '', redirectTo: 'onboarding', pathMatch: 'full' },
      {
        path: 'onboarding',
        component: OnboardingReviewPage,
        canActivate: [platformAdminGuard('store:read_all')],
      },
      {
        path: 'onboarding/:id',
        component: ReadinessDetailPage,
        canActivate: [platformAdminGuard('store:read_all')],
      },
      // IAM section — organizations/platform roles/policies only, all
      // backed by real legacy IAMService REST mappings; a platform-wide
      // "tenant directory" was dropped (no ListTenants-style RPC exists —
      // see PZEP-0020's Open Questions). Gated by platform:manage_roles
      // (existing permission, platform_owner only by default) since IAM
      // management is more sensitive than onboarding review.
      {
        path: 'iam/organizations',
        component: OrganizationsPage,
        canActivate: [platformAdminGuard('platform:manage_roles')],
      },
      {
        path: 'iam/platform-roles',
        component: PlatformRolesPage,
        canActivate: [platformAdminGuard('platform:manage_roles')],
      },
      {
        path: 'iam/policies',
        component: PoliciesPage,
        canActivate: [platformAdminGuard('platform:manage_roles')],
      },
    ],
  },
];
