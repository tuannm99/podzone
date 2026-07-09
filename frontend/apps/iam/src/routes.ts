import { createRoute, type AnyRoute } from '@tanstack/solid-router'
import { remotePage } from '@/solid/remotePage'

type Guards = {
    requireAuth: () => void
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const adminIamRoute = createRoute({
        getParentRoute: () => parent,
        path: '/admin/iam',
        beforeLoad: guards.requireAuth,
        validateSearch: (search: Record<string, unknown>) => ({
            section: search.section as string | undefined,
        }),
        component: remotePage(
            // __MFE_IAM__ replaced at build time; dead branch is tree-shaken.
            __MFE_IAM__ ? () => import('iam/AdminIamPage') : () => import('./pages/AdminIamPage'),
            'iam',
        ),
    })

    return [adminIamRoute]
}
