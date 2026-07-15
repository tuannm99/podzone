import { Routes } from '@angular/router';
import { Shell } from './shell/shell.component';
import { AdminHomePage } from './features/onboarding/admin-home-page/admin-home-page.component';
import { LoginPage } from './features/auth/login-page/login-page.component';
import { RegisterPage } from './features/auth/register-page/register-page.component';
import { authGuard } from './core/auth/auth.guard';

export const routes: Routes = [
  { path: 'login', component: LoginPage },
  { path: 'register', component: RegisterPage },
  {
    path: '',
    component: Shell,
    canActivate: [authGuard],
    children: [{ path: '', component: AdminHomePage }],
  },
];
