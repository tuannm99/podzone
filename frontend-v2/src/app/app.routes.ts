import { Routes } from '@angular/router';
import { Shell } from './shell/shell.component';
import { AdminHomePage } from './features/onboarding/admin-home-page/admin-home-page.component';

export const routes: Routes = [
  {
    path: '',
    component: Shell,
    children: [{ path: '', component: AdminHomePage }],
  },
];
