const TENANT_KEY = 'x_tenant_id';

export const tenantStorage = {
  getTenantID() {
    return localStorage.getItem(TENANT_KEY) || '';
  },
  setTenantID(id) {
    if (!id) return;
    localStorage.setItem(TENANT_KEY, id);
  },
  clearTenantID() {
    localStorage.removeItem(TENANT_KEY);
  },
};
