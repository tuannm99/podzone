import { createContext, useContext } from 'solid-js'

export interface AuthContext {
    // Token / session
    getToken: () => string
    isAuthenticated: () => boolean
    getActiveTenantId: () => string
    getSessionId: () => string

    // User
    getUserId: () => number
    getUserEmail: () => string

    // Tenant side-channel — replaces tenantStorage.{set,get,clear}TenantID
    setActiveTenantId: (id: string) => void
    getLastKnownTenantId: () => string
    clearActiveTenantId: () => void

    // Store persistence per tenant — replaces storeStorage.*
    getStoreId: (tenantId: string) => string
    setStoreId: (tenantId: string, storeId: string) => void
    clearStoreId: (tenantId: string) => void
    clearAllStoreIds: () => void
}

const AuthContextToken = createContext<AuthContext>()

export function useAuthContext(): AuthContext {
    const ctx = useContext(AuthContextToken)
    if (!ctx) throw new Error('useAuthContext must be used inside AuthContextProvider')
    return ctx
}

export { AuthContextToken }
