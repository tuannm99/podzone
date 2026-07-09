import { createRootRoute, createRoute, createRouter, lazyRouteComponent, redirect } from '@tanstack/solid-router'
import { createRouteTree as createBackofficeRouteTree } from '@backoffice/routes'
import { createRouteTree as createIamRouteTree } from '@iam/routes'
import { createRouteTree as createOnboardingRouteTree } from '@onboarding/routes'
import { createRouteTree as createShellRouteTree } from './modules/shell/routes'
import { ensureActiveTenant } from '@podzone/shared/services/auth'
import { tokenStorage } from '@podzone/shared/services/tokenStorage'
import Root from './solid/root'

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

const guards = { requireAuth, requireGuest, requireTenantAccess }

const rootRoute = createRootRoute({
    component: Root,
    notFoundComponent: lazyRouteComponent(() => import('./solid/routes/NotFoundRoute')),
})

const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    beforeLoad: () => {
        throw redirect({ to: tokenStorage.getToken() ? '/admin' : '/auth/login' })
    },
})

const routeTree = rootRoute.addChildren([
    indexRoute,
    ...createShellRouteTree(rootRoute, guards),
    ...createOnboardingRouteTree(rootRoute, guards),
    ...createIamRouteTree(rootRoute, guards),
    ...createBackofficeRouteTree(rootRoute, guards),
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
