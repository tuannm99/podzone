import { Routes } from '@angular/router';
import { AdminHomePage } from './features/onboarding/admin-home-page/admin-home-page.component';
import { StoreChooserPage } from './features/onboarding/store-chooser-page/store-chooser-page.component';
import { LoginPage } from './features/auth/login-page/login-page.component';
import { RegisterPage } from './features/auth/register-page/register-page.component';
import { TenantHomePage } from './features/backoffice/tenant-home-page/tenant-home-page.component';
import { StoreListPage } from './features/backoffice/store-list-page/store-list-page.component';
import { Shell } from './shell/shell.component';
import { authGuard } from './core/auth/auth.guard';

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
    ],
  },
];
