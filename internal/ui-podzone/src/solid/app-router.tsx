import { createRootRoute, createRoute, createRouter, lazyRouteComponent, redirect } from '@tanstack/solid-router'
import { backofficeRouteComponents } from '../modules/backoffice/routes'
import { iamRouteComponents } from '../modules/iam/routes'
import { onboardingRouteComponents } from '../modules/onboarding/routes'
import { shellRouteComponents } from '../modules/shell/routes'
import { ensureActiveTenant } from '../services/auth'
import Root from './root'
import { tokenStorage } from '../services/tokenStorage'

function requireAuth() {
    if (!tokenStorage.getToken()) {
        throw redirect({ to: '/auth/login' })
    }
}

function requireGuest() {
    if (tokenStorage.getToken()) {
        throw redirect({ to: '/admin' })
    }
}

async function requireTenantAccess(tenantId: string) {
    requireAuth()

    const { success } = await ensureActiveTenant(tenantId)
    if (!success) {
        throw redirect({ to: '/admin' })
    }
}

const rootRoute = createRootRoute({
    component: Root,
    notFoundComponent: lazyRouteComponent(() => import('./routes/NotFoundRoute')),
})

const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    beforeLoad: () => {
        throw redirect({ to: tokenStorage.getToken() ? '/admin' : '/auth/login' })
    },
})

const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/login',
    beforeLoad: requireGuest,
    component: lazyRouteComponent(shellRouteComponents.login),
})

const registerRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/register',
    beforeLoad: requireGuest,
    component: lazyRouteComponent(shellRouteComponents.register),
})

const googleCallbackRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/google/callback',
    beforeLoad: requireGuest,
    component: lazyRouteComponent(shellRouteComponents.googleCallback),
})

const acceptInviteRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/invite/accept',
    component: lazyRouteComponent(shellRouteComponents.acceptInvite),
})

const devAuthBootstrapRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/dev/bootstrap',
    component: lazyRouteComponent(shellRouteComponents.devAuthBootstrap),
})

const adminHomeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin',
    beforeLoad: requireAuth,
    component: lazyRouteComponent(onboardingRouteComponents.adminHome),
})

const adminSettingsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin/settings',
    beforeLoad: requireAuth,
    validateSearch: (search: Record<string, unknown>) => ({
        tab: search.tab as string | undefined,
    }),
    component: lazyRouteComponent(onboardingRouteComponents.adminSettings),
})

const adminProvisioningRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin/provisioning',
    beforeLoad: requireAuth,
    validateSearch: (search: Record<string, unknown>) => ({
        tab: search.tab as string | undefined,
    }),
    component: lazyRouteComponent(onboardingRouteComponents.adminProvisioning),
})

const adminIamRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/admin/iam',
    beforeLoad: requireAuth,
    validateSearch: (search: Record<string, unknown>) => ({
        section: search.section as string | undefined,
    }),
    component: lazyRouteComponent(iamRouteComponents.adminIam),
})

const tenantHomeRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantHome),
})

const tenantOrdersRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/orders',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    validateSearch: (search: Record<string, unknown>) => ({
        queueView: typeof search.queueView === 'string' ? search.queueView : 'all',
        queueSort: typeof search.queueSort === 'string' ? search.queueSort : 'priority',
        operatorLens: typeof search.operatorLens === 'string' ? search.operatorLens : '',
        queuePage: typeof search.queuePage === 'number' ? search.queuePage : 1,
        appliedQueueSearch: typeof search.appliedQueueSearch === 'string' ? search.appliedQueueSearch : '',
    }),
    component: lazyRouteComponent(backofficeRouteComponents.tenantOrders),
})

const tenantOrderAuditRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/orders/audit',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantOrderAudit),
})

const tenantOrderFinanceRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/orders/finance',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantOrderFinance),
})

const tenantPartnersRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/partners',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantPartners),
})

const tenantPartnerDetailRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/partners/$partnerId',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantPartnerDetail),
})

const tenantProductSetupRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/t/$tenantId/products/setup',
    beforeLoad: async ({ params }) => requireTenantAccess(params.tenantId),
    component: lazyRouteComponent(backofficeRouteComponents.tenantProductSetup),
})

const routeTree = rootRoute.addChildren([
    indexRoute,
    loginRoute,
    registerRoute,
    googleCallbackRoute,
    acceptInviteRoute,
    devAuthBootstrapRoute,
    adminHomeRoute,
    adminProvisioningRoute,
    adminSettingsRoute,
    adminIamRoute,
    tenantHomeRoute,
    tenantOrdersRoute,
    tenantOrderAuditRoute,
    tenantOrderFinanceRoute,
    tenantPartnersRoute,
    tenantPartnerDetailRoute,
    tenantProductSetupRoute,
])

export const router = createRouter({
    routeTree,
    defaultPreload: 'intent',
    scrollRestoration: false,
})

declare module '@tanstack/solid-router' {
    interface Register {
        router: typeof router
    }
}
