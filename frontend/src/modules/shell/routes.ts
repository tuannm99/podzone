import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireGuest: () => void
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

    return [loginRoute, registerRoute, googleCallbackRoute, acceptInviteRoute, devAuthBootstrapRoute]
}
