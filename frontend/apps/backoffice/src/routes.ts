import { createRoute, type AnyRoute } from '@tanstack/solid-router'
import { remotePage } from '@/solid/remotePage'

type Guards = {
    requireTenantAccess: (tenantId: string) => Promise<void>
}

export function createRouteTree<TParent extends AnyRoute>(parent: TParent, guards: Guards) {
    const tenantHomeRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantHomePage')
                : () => import('./pages/TenantHomePage'),
            'backoffice',
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
            appliedQueueSearch:
                typeof search.appliedQueueSearch === 'string' ? search.appliedQueueSearch : '',
        }),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrdersPage')
                : () => import('./pages/TenantOrdersPage'),
            'backoffice',
        ),
    })

    const tenantOrderAuditRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/audit',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrderAuditPage')
                : () => import('./pages/TenantOrderAuditPage'),
            'backoffice',
        ),
    })

    const tenantOrderFinanceRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/orders/finance',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantOrderFinancePage')
                : () => import('./pages/TenantOrderFinancePage'),
            'backoffice',
        ),
    })

    const tenantPartnersRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantPartnersPage')
                : () => import('./pages/TenantPartnersPage'),
            'backoffice',
        ),
    })

    const tenantPartnerDetailRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/partners/$partnerId',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantPartnerDetailPage')
                : () => import('./pages/TenantPartnerDetailPage'),
            'backoffice',
        ),
    })

    const tenantProductSetupRoute = createRoute({
        getParentRoute: () => parent,
        path: '/t/$tenantId/products/setup',
        beforeLoad: async ({ params }) => guards.requireTenantAccess(params.tenantId),
        component: remotePage(
            __MFE_BACKOFFICE__
                ? () => import('backoffice/TenantProductSetupPage')
                : () => import('./pages/TenantProductSetupPage'),
            'backoffice',
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
