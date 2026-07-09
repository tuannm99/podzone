import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireAuth: () => void
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const adminHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin',
        beforeLoad: guards.requireAuth,
        component: lazyRouteComponent(
            // __MFE_ONBOARDING__ replaced at build time; dead branch is tree-shaken.
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminHomePage')
                : () => import('@onboarding/pages/AdminHomePage')
        ),
    })

    const adminSettingsRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/settings',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: lazyRouteComponent(
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminSettingsPage')
                : () => import('@onboarding/pages/AdminSettingsPage')
        ),
    })

    const adminProvisioningRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/provisioning',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: lazyRouteComponent(
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminProvisioningPage')
                : () => import('@onboarding/pages/AdminProvisioningPage')
        ),
    })

    return [adminHomeRoute, adminSettingsRoute, adminProvisioningRoute]
}
