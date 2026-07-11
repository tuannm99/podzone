import { createContext, useContext, type Context } from 'solid-js'
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

declare global {
    interface Window {
        __pz_workspace_ctx__?: Context<WorkspaceContextValue | undefined>
    }
}

if (!window.__pz_workspace_ctx__) {
    window.__pz_workspace_ctx__ = createContext<WorkspaceContextValue>()
}
export const WorkspaceContext: Context<WorkspaceContextValue | undefined> = window.__pz_workspace_ctx__

export function useTenantWorkspace() {
    return useContext(WorkspaceContext) || null
}
