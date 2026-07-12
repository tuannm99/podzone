import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireGuest: () => void
    requireAuth: () => void
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const loginRoute = createRoute({
        getParentRoute: () => parent,
        path: '/auth/login',
        beforeLoad: guards.requireGuest,
        component: lazyRouteComponent(() => import('./pages/auth/LoginPage')),
    })

    const registerRoute = createRoute({
        getParentRoute: () => parent,
        path: '/auth/register',
        beforeLoad: guards.requireGuest,
        component: lazyRouteComponent(() => import('./pages/auth/RegisterPage')),
    })

    const googleCallbackRoute = createRoute({
        getParentRoute: () => parent,
        path: '/auth/google/callback',
        beforeLoad: guards.requireGuest,
        component: lazyRouteComponent(() => import('./pages/auth/GoogleCallbackPage')),
    })

    const acceptInviteRoute = createRoute({
        getParentRoute: () => parent,
        path: '/auth/invite/accept',
        component: lazyRouteComponent(() => import('./pages/auth/AcceptInvitePage')),
    })

    const devAuthBootstrapRoute = createRoute({
        getParentRoute: () => parent,
        path: '/auth/dev/bootstrap',
        component: lazyRouteComponent(() => import('./pages/auth/DevAuthBootstrapPage')),
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

    return [
        loginRoute,
        registerRoute,
        googleCallbackRoute,
        acceptInviteRoute,
        devAuthBootstrapRoute,
        adminSettingsRoute,
    ]
}
