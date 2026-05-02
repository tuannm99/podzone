import {
  createRootRoute,
  createRoute,
  createRouter,
  lazyRouteComponent,
  redirect,
} from '@tanstack/solid-router';
import { ensureActiveTenant } from '../services/auth';
import Root from './root';
import { tokenStorage } from '../services/tokenStorage';

function requireAuth() {
  if (!tokenStorage.getToken()) {
    throw redirect({ to: '/auth/login' });
  }
}

function requireGuest() {
  if (tokenStorage.getToken()) {
    throw redirect({ to: '/admin' });
  }
}

async function requireTenantAccess(tenantId: string) {
  requireAuth();

  const { success } = await ensureActiveTenant(tenantId);
  if (!success) {
    throw redirect({ to: '/admin' });
  }
}

const rootRoute = createRootRoute({
  component: Root,
  notFoundComponent: lazyRouteComponent(() => import('./routes/NotFoundRoute')),
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    throw redirect({ to: tokenStorage.getToken() ? '/admin' : '/auth/login' });
  },
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/login',
  beforeLoad: requireGuest,
  component: lazyRouteComponent(() => import('./pages/podzone/LoginPage')),
});

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/register',
  beforeLoad: requireGuest,
  component: lazyRouteComponent(() => import('./pages/podzone/RegisterPage')),
});

const googleCallbackRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/google/callback',
  beforeLoad: requireGuest,
  component: lazyRouteComponent(
    () => import('./pages/podzone/GoogleCallbackPage')
  ),
});

const acceptInviteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/invite/accept',
  component: lazyRouteComponent(
    () => import('./pages/podzone/AcceptInvitePage')
  ),
});

const adminHomeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  beforeLoad: requireAuth,
  component: lazyRouteComponent(() => import('./pages/podzone/AdminHomePage')),
});

const adminSettingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin/settings',
  beforeLoad: requireAuth,
  component: lazyRouteComponent(
    () => import('./pages/podzone/AdminSettingsPage')
  ),
});

const tenantHomeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/t/$tenantId',
  beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
  component: lazyRouteComponent(() => import('./pages/podzone/TenantHomePage')),
});

const tenantOrdersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/t/$tenantId/orders',
  beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
  component: lazyRouteComponent(
    () => import('./pages/podzone/TenantOrdersPage')
  ),
});

const tenantPartnersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/t/$tenantId/partners',
  beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
  component: lazyRouteComponent(
    () => import('./pages/podzone/TenantPartnersPage')
  ),
});

const tenantProductSetupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/t/$tenantId/products/setup',
  beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
  component: lazyRouteComponent(
    () => import('./pages/podzone/TenantProductSetupPage')
  ),
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  registerRoute,
  googleCallbackRoute,
  acceptInviteRoute,
  adminHomeRoute,
  adminSettingsRoute,
  tenantHomeRoute,
  tenantOrdersRoute,
  tenantPartnersRoute,
  tenantProductSetupRoute,
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  scrollRestoration: false,
});

declare module '@tanstack/solid-router' {
  interface Register {
    router: typeof router;
  }
}
