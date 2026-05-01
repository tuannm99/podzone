const TENANT_KEY = 'x_tenant_id';

export const tenantStorage = {
  getTenantID(): string {
    return localStorage.getItem(TENANT_KEY) || '';
  },
  setTenantID(id: string): void {
    if (!id) return;
    localStorage.setItem(TENANT_KEY, id);
  },
  clearTenantID(): void {
    localStorage.removeItem(TENANT_KEY);
  },
};
