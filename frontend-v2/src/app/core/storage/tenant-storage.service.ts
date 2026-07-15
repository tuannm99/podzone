import { Injectable } from '@angular/core';

const TENANT_KEY = 'x_tenant_id';

@Injectable({ providedIn: 'root' })
export class TenantStorageService {
  getTenantId(): string {
    return localStorage.getItem(TENANT_KEY) || '';
  }

  setTenantId(id: string): void {
    if (!id) return;
    localStorage.setItem(TENANT_KEY, id);
  }

  clearTenantId(): void {
    localStorage.removeItem(TENANT_KEY);
  }
}
