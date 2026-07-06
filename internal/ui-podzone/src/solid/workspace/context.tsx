import {
    createContext,
    createEffect,
    createMemo,
    createResource,
    createSignal,
    type ParentProps,
    useContext,
} from 'solid-js'
import { tenantStorage } from '../../services/tenantStorage'
import { getStore, listStores, type StoreInfo } from '../../services/store'
import { storeStorage } from '../../services/storeStorage'

type WorkspaceContextValue = {
    tenantId: () => string
    stores: () => StoreInfo[]
    currentStoreId: () => string
    currentStore: () => StoreInfo | undefined
    loading: () => boolean
    error: () => string
    setCurrentStoreId: (storeId: string) => void
}

const WorkspaceContext = createContext<WorkspaceContextValue>()

function initialStoreID(tenantId: string) {
    const requested = new URLSearchParams(window.location.search).get('storeId') || ''
    return requested.trim() || storeStorage.getStoreID(tenantId)
}

function syncStoreIdToURL(storeId: string) {
    const normalizedStoreId = storeId.trim()
    if (!normalizedStoreId || !window.location.pathname.startsWith('/t/')) {
        return
    }
    const params = new URLSearchParams(window.location.search)
    params.set('storeId', normalizedStoreId)
    const query = params.toString()
    const nextURL = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash}`
    window.history.replaceState(window.history.state, '', nextURL)
}

export function TenantWorkspaceProvider(props: ParentProps<{ tenantId: string }>) {
    const [currentStoreId, setCurrentStoreIdState] = createSignal('')
    const tenantId = createMemo(() => props.tenantId.trim())
    const [storeResource, { mutate: clearStore }] = createResource(
        () => {
            const currentTenantID = tenantId()
            return currentTenantID ? { tenantId: currentTenantID, storeId: currentStoreId() } : undefined
        },
        async ({ tenantId: currentTenantID, storeId }) => {
            tenantStorage.setTenantID(currentTenantID)
            if (storeId) {
                const result = await getStore(storeId)
                if (result.success) return result.data
                if (!result.message.toLowerCase().includes('not found')) {
                    throw new Error(result.message)
                }
            }
            const result = await listStores({
                page: 1,
                pageSize: 1,
                sortBy: 'createdAt',
                sortDirection: 'SORT_DIRECTION_DESC',
            })
            if (!result.success) throw new Error(result.message)
            return result.data.items[0]
        }
    )

    const setCurrentStoreId = (storeId: string) => {
        const normalizedStoreId = storeId.trim()
        if (!tenantId() || !normalizedStoreId) return
        setCurrentStoreIdState(normalizedStoreId)
        storeStorage.setStoreID(tenantId(), normalizedStoreId)
        syncStoreIdToURL(normalizedStoreId)
    }

    let loadedTenantID = ''
    createEffect(() => {
        const nextTenantId = tenantId()
        if (nextTenantId === loadedTenantID) return
        loadedTenantID = nextTenantId
        clearStore(undefined)
        if (!nextTenantId) {
            setCurrentStoreIdState('')
            return
        }
        setCurrentStoreIdState(initialStoreID(nextTenantId))
    })

    createEffect(() => {
        const store = storeResource.latest
        if (!store || currentStoreId() === store.id) return
        setCurrentStoreId(store.id)
    })

    const stores = () => (storeResource.latest ? [storeResource.latest] : [])
    const currentStore = () => storeResource.latest
    const loading = () => storeResource.loading
    const error = () => (storeResource.error instanceof Error ? storeResource.error.message : '')

    return (
        <WorkspaceContext.Provider
            value={{
                tenantId,
                stores,
                currentStoreId,
                currentStore,
                loading,
                error,
                setCurrentStoreId,
            }}
        >
            {props.children}
        </WorkspaceContext.Provider>
    )
}

export function useTenantWorkspace() {
    return useContext(WorkspaceContext) || null
}
