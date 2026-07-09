import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireTenantAccess: (tenantId: string) => Promise<void>
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const tenantHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            // __MFE_BACKOFFICE__ replaced at build time; dead branch is tree-shaken.
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantHomePage')
                : () => import('@backoffice/pages/TenantHomePage')
        ),
    })

    const tenantOrdersRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        validateSearch: (search: Record<string, unknown>) => ({
            queueView: typeof search.queueView === 'string' ? search.queueView : 'all',
            queueSort: typeof search.queueSort === 'string' ? search.queueSort : 'priority',
            operatorLens: typeof search.operatorLens === 'string' ? search.operatorLens : '',
            queuePage: typeof search.queuePage === 'number' ? search.queuePage : 1,
            appliedQueueSearch: typeof search.appliedQueueSearch === 'string' ? search.appliedQueueSearch : '',
        }),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrdersPage')
                : () => import('@backoffice/pages/TenantOrdersPage')
        ),
    })

    const tenantOrderAuditRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/audit',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrderAuditPage')
                : () => import('@backoffice/pages/TenantOrderAuditPage')
        ),
    })

    const tenantOrderFinanceRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/finance',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrderFinancePage')
                : () => import('@backoffice/pages/TenantOrderFinancePage')
        ),
    })

    const tenantPartnersRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantPartnersPage')
                : () => import('@backoffice/pages/TenantPartnersPage')
        ),
    })

    const tenantPartnerDetailRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners/$partnerId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantPartnerDetailPage')
                : () => import('@backoffice/pages/TenantPartnerDetailPage')
        ),
    })

    const tenantProductSetupRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/products/setup',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantProductSetupPage')
                : () => import('@backoffice/pages/TenantProductSetupPage')
        ),
    })

    return [
        tenantHomeRoute,
        tenantOrdersRoute,
        tenantOrderAuditRoute,
        tenantOrderFinanceRoute,
        tenantPartnersRoute,
        tenantPartnerDetailRoute,
        tenantProductSetupRoute,
    ]
}
