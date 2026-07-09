import { createContext, useContext } from 'solid-js'
import type { StoreInfo } from '../services/store'

export type WorkspaceContextValue = {
    tenantId: () => string
    stores: () => StoreInfo[]
    currentStoreId: () => string
    currentStore: () => StoreInfo | undefined
    loading: () => boolean
    error: () => string
    setCurrentStoreId: (storeId: string) => void
}

export const WorkspaceContext = createContext<WorkspaceContextValue>()

export function useTenantWorkspace() {
    return useContext(WorkspaceContext) || null
}
