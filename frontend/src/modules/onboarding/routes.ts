import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireAuth: () => void
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const adminHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin',
        beforeLoad: guards.requireAuth,
        component: lazyRouteComponent(() => import('./pages/AdminHomePage')),
    })

    const adminSettingsRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/settings',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: lazyRouteComponent(() => import('./pages/AdminSettingsPage')),
    })

    const adminProvisioningRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/provisioning',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: lazyRouteComponent(() => import('./pages/AdminProvisioningPage')),
    })

    return [adminHomeRoute, adminSettingsRoute, adminProvisioningRoute]
}
