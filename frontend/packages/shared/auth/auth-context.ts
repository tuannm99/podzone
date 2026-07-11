import { createContext, useContext, type Context } from 'solid-js'

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

declare global {
    interface Window {
        __pz_auth_ctx__?: Context<AuthContext | undefined>
        __pz_auth_value__?: AuthContext
    }
}

// Global registry ensures one token survives across MFE bundle boundaries.
// Federation singletons don't apply to subpath exports (@podzone/shared/auth),
// so each remote bundle would otherwise get its own createContext() object.
if (!window.__pz_auth_ctx__) {
    window.__pz_auth_ctx__ = createContext<AuthContext>()
}
export const AuthContextToken: Context<AuthContext | undefined> = window.__pz_auth_ctx__

export function useAuthContext(): AuthContext {
    // MFE remotes run in a separate solid-js reactive scope; they cannot reach
    // the host's context provider tree, so we fall back to the global value
    // that AuthContextProvider sets on mount.
    if (window.__pz_auth_value__) return window.__pz_auth_value__
    const ctx = useContext(AuthContextToken)
    if (!ctx) throw new Error('useAuthContext must be used inside AuthContextProvider')
    return ctx
}
