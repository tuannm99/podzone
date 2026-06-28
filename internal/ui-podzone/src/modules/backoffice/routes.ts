export const backofficeRouteComponents = {
  tenantHome: () => import('./pages/TenantHomePage'),
  tenantOrderAudit: () => import('./pages/TenantOrderAuditPage'),
  tenantOrderFinance: () => import('./pages/TenantOrderFinancePage'),
  tenantOrders: () => import('./pages/TenantOrdersPage'),
  tenantPartnerDetail: () => import('./pages/TenantPartnerDetailPage'),
  tenantPartners: () => import('./pages/TenantPartnersPage'),
  tenantProductSetup: () => import('./pages/TenantProductSetupPage'),
}
