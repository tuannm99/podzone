import { createContext, useContext } from 'solid-js'

export interface AuthContext {
    getToken: () => string
    isAuthenticated: () => boolean
    getActiveTenantId: () => string
    getSessionId: () => string
    getUserId: () => number
    getUserEmail: () => string
    setActiveTenantId: (id: string) => void
    getLastKnownTenantId: () => string
    clearActiveTenantId: () => void
    getStoreId: (tenantId: string) => string
    setStoreId: (tenantId: string, storeId: string) => void
    clearStoreId: (tenantId: string) => void
    clearAllStoreIds: () => void
}

export const AuthContextToken = createContext<AuthContext>()

export function useAuthContext(): AuthContext {
    const ctx = useContext(AuthContextToken)
    if (!ctx) throw new Error('useAuthContext must be used inside AuthContextProvider')
    return ctx
}
