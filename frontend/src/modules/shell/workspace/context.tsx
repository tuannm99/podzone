import { useNavigate, useSearch } from '@tanstack/solid-router'
import { createEffect, createMemo, createResource, createSignal, type ParentProps } from 'solid-js'
import { tenantStorage } from '@/services/tenantStorage'
import { getStore, listStores } from '@/services/store'
import { storeStorage } from '@/services/storeStorage'
import { WorkspaceContext } from '@/solid/context/workspace-context'

function initialStoreID(tenantId: string, requestedStoreId: string) {
    const requested = requestedStoreId || ''
    return requested.trim() || storeStorage.getStoreID(tenantId)
}

export function TenantWorkspaceProvider(props: ParentProps<{ tenantId: string }>) {
    const navigate = useNavigate()
    const search = useSearch({ strict: false }) as () => Record<string, unknown>
    const [currentStoreId, setCurrentStoreIdState] = createSignal('')
    const tenantId = createMemo(() => props.tenantId.trim())
    const requestedStoreID = () => {
        const raw = search().storeId
        return typeof raw === 'string' ? raw.trim() : ''
    }
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
        void navigate({
            search: (previous: Record<string, unknown>) => ({ ...previous, storeId: normalizedStoreId }),
            replace: true,
        } as unknown as Parameters<typeof navigate>[0])
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
        setCurrentStoreIdState(initialStoreID(nextTenantId, requestedStoreID()))
    })

    createEffect(() => {
        const requested = requestedStoreID()
        const currentTenantID = tenantId()
        if (!currentTenantID || !requested || requested === currentStoreId()) return
        setCurrentStoreIdState(requested)
        storeStorage.setStoreID(currentTenantID, requested)
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

export { useTenantWorkspace } from '@/solid/context/workspace-context'
