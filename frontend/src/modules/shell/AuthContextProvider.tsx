import { type ParentProps } from 'solid-js'
import { storeStorage } from '@podzone/shared/services/storeStorage'
import { tenantStorage } from '@podzone/shared/services/tenantStorage'
import { tokenStorage } from '@podzone/shared/services/tokenStorage'
import { type AuthContext, AuthContextToken } from '@podzone/shared/auth'

export function AuthContextProvider(props: ParentProps) {
    const ctx: AuthContext = {
        getToken: () => tokenStorage.getToken(),
        isAuthenticated: () => Boolean(tokenStorage.getToken()),
        getActiveTenantId: () => tokenStorage.getActiveTenantID(),
        getSessionId: () => tokenStorage.getSessionID(),
        getUserId: () => tokenStorage.getUserID() ?? 0,
        getUserEmail: () => tokenStorage.getUser()?.email ?? '',

        setActiveTenantId: (id) => tenantStorage.setTenantID(id),
        getLastKnownTenantId: () => tenantStorage.getTenantID(),
        clearActiveTenantId: () => tenantStorage.clearTenantID(),

        getStoreId: (tenantId) => storeStorage.getStoreID(tenantId),
        setStoreId: (tenantId, storeId) => storeStorage.setStoreID(tenantId, storeId),
        clearStoreId: (tenantId) => storeStorage.clearStoreID(tenantId),
        clearAllStoreIds: () => storeStorage.clearAll(),
    }

    return <AuthContextToken.Provider value={ctx}>{props.children}</AuthContextToken.Provider>
}
