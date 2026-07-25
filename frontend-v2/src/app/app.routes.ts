import { Routes } from '@angular/router';
import { AdminHomePage } from './features/onboarding/admin-home-page/admin-home-page.component';
import { StoreChooserPage } from './features/onboarding/store-chooser-page/store-chooser-page.component';
import { LoginPage } from './features/auth/login-page/login-page.component';
import { RegisterPage } from './features/auth/register-page/register-page.component';
import { TenantHomePage } from './features/backoffice/tenant-home-page/tenant-home-page.component';
import { StoreListPage } from './features/backoffice/store-list-page/store-list-page.component';
import { authGuard } from './core/auth/auth.guard';

// Workspace/store selection are bare screens (no Shell nav/topbar) — they
// run before there's a store context for the Shell's sidebar to be about.
// Shell wraps real app content once frontend-v2 hosts screens past the
// handoff (see PZEP-0008 M3).
export const routes: Routes = [
  { path: 'login', component: LoginPage },
  { path: 'register', component: RegisterPage },
  { path: '', component: AdminHomePage, canActivate: [authGuard] },
  { path: 't/:tenantId/stores', component: StoreChooserPage, canActivate: [authGuard] },
  // Minimal Backoffice shell — see PZEP-0011. More Backoffice pages are
  // ported one PZEP at a time.
  { path: 't/:tenantId/backoffice', component: TenantHomePage, canActivate: [authGuard] },
  // First declarative-collection screen — see ADR-0008.
  { path: 't/:tenantId/backoffice/stores', component: StoreListPage, canActivate: [authGuard] },
];
