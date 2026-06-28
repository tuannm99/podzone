const STORE_KEY = 'x_store_id_by_tenant'

type StoreMap = Record<string, string>

function readStoreMap(): StoreMap {
  try {
    const raw = localStorage.getItem(STORE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object') return {}
    return parsed as StoreMap
  } catch {
    return {}
  }
}

function writeStoreMap(value: StoreMap) {
  localStorage.setItem(STORE_KEY, JSON.stringify(value))
}

export const storeStorage = {
  getStoreID(tenantId: string): string {
    const normalizedTenantId = tenantId.trim()
    if (!normalizedTenantId) return ''
    return readStoreMap()[normalizedTenantId] || ''
  },
  setStoreID(tenantId: string, storeId: string): void {
    const normalizedTenantId = tenantId.trim()
    const normalizedStoreId = storeId.trim()
    if (!normalizedTenantId || !normalizedStoreId) return
    const current = readStoreMap()
    current[normalizedTenantId] = normalizedStoreId
    writeStoreMap(current)
  },
  clearStoreID(tenantId: string): void {
    const normalizedTenantId = tenantId.trim()
    if (!normalizedTenantId) return
    const current = readStoreMap()
    delete current[normalizedTenantId]
    writeStoreMap(current)
  },
  clearAll(): void {
    localStorage.removeItem(STORE_KEY)
  },
}
