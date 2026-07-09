import { createRoute, type AnyRoute } from '@tanstack/solid-router'
import { remotePage } from '@/solid/remotePage'

type Guards = {
    requireAuth: () => void
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const adminHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin',
        beforeLoad: guards.requireAuth,
        component: remotePage(
            // __MFE_ONBOARDING__ replaced at build time; dead branch is tree-shaken.
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminHomePage')
                : () => import('./pages/AdminHomePage'),
            'onboarding',
        ),
    })

    const adminSettingsRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/settings',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: remotePage(
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminSettingsPage')
                : () => import('./pages/AdminSettingsPage'),
            'onboarding',
        ),
    })

    const adminProvisioningRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/provisioning',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            tab: search.tab as string | undefined,
        }),
        component: remotePage(
            __MFE_ONBOARDING__
                ? () => import('onboarding/AdminProvisioningPage')
                : () => import('./pages/AdminProvisioningPage'),
            'onboarding',
        ),
    })

    return [adminHomeRoute, adminSettingsRoute, adminProvisioningRoute]
}
