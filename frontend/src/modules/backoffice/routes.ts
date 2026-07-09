import { createRoute, lazyRouteComponent, type AnyRoute } from '@tanstack/solid-router'

type Guards = {
    requireTenantAccess: (tenantId: string) => Promise<void>
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const tenantHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantHomePage')),
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
        component: lazyRouteComponent(() => import('./pages/TenantOrdersPage')),
    })

    const tenantOrderAuditRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/audit',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantOrderAuditPage')),
    })

    const tenantOrderFinanceRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/finance',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantOrderFinancePage')),
    })

    const tenantPartnersRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantPartnersPage')),
    })

    const tenantPartnerDetailRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners/$partnerId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantPartnerDetailPage')),
    })

    const tenantProductSetupRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/products/setup',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: lazyRouteComponent(() => import('./pages/TenantProductSetupPage')),
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
