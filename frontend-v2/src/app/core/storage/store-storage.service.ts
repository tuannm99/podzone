import { Injectable } from '@angular/core';

// Ported from frontend/packages/shared/services/storeStorage.ts - keeps the
// same localStorage key so a user who used the old SolidJS app doesn't lose
// their per-tenant active store on the switch to frontend-v2.
const STORE_KEY = 'x_store_id_by_tenant';

type StoreMap = Record<string, string>;

function readStoreMap(): StoreMap {
  try {
    const raw = localStorage.getItem(STORE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== 'object') return {};
    return parsed as StoreMap;
  } catch {
    return {};
  }
}

function writeStoreMap(value: StoreMap): void {
  localStorage.setItem(STORE_KEY, JSON.stringify(value));
}

@Injectable({ providedIn: 'root' })
export class StoreStorageService {
  getStoreId(tenantId: string): string {
    const normalizedTenantId = tenantId.trim();
    if (!normalizedTenantId) return '';
    return readStoreMap()[normalizedTenantId] || '';
  }

  setStoreId(tenantId: string, storeId: string): void {
    const normalizedTenantId = tenantId.trim();
    const normalizedStoreId = storeId.trim();
    if (!normalizedTenantId || !normalizedStoreId) return;
    const current = readStoreMap();
    current[normalizedTenantId] = normalizedStoreId;
    writeStoreMap(current);
  }

  clearStoreId(tenantId: string): void {
    const normalizedTenantId = tenantId.trim();
    if (!normalizedTenantId) return;
    const current = readStoreMap();
    delete current[normalizedTenantId];
    writeStoreMap(current);
  }

  clearAll(): void {
    localStorage.removeItem(STORE_KEY);
  }
}
