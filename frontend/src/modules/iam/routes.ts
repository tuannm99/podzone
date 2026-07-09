import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

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
        component: lazyRouteComponent(
            // __MFE_IAM__ replaced at build time; dead branch is tree-shaken.
            __MFE_IAM__ ? () => import('iam/AdminIamPage') : () => import('@iam/pages/AdminIamPage')
        ),
    })

    return [adminIamRoute]
}
