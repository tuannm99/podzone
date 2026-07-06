import { createResource, type Accessor } from 'solid-js'
import { ensureActiveTenant } from '@/services/auth'
import { listStoreRequests } from '@/services/onboarding'
import { listStores } from '@/services/store'
import { tenantStorage } from '@/services/tenantStorage'
import { createPaginatedResource } from '@/solid/pagination'

export function createStoreCollectionsViewModel(
    selectedWorkspaceID: Accessor<string>,
    loadingWorkspaces: Accessor<boolean>
) {
    const [workspaceSession] = createResource(
        () => {
            const workspaceID = selectedWorkspaceID().trim()
            return workspaceID && !loadingWorkspaces() ? workspaceID : undefined
        },
        async (workspaceID) => {
            const switched = await ensureActiveTenant(workspaceID)
            if (!switched.success) {
                throw new Error(switched.data.message || 'Failed to load workspace')
            }
            tenantStorage.setTenantID(workspaceID)
            return workspaceID
        }
    )
    const workspaceReady = () => workspaceSession.latest === selectedWorkspaceID().trim()
    const stores = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listStores(query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: workspaceReady,
            dependency: selectedWorkspaceID,
        }
    )
    const requests = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'updatedAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listStoreRequests(selectedWorkspaceID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: workspaceReady,
            dependency: selectedWorkspaceID,
        }
    )
    const sessionError = () => (workspaceSession.error instanceof Error ? workspaceSession.error.message : '')

    return {
        stores,
        requests,
        storesError: () => sessionError() || stores.error(),
        requestsError: () => sessionError() || requests.error(),
    }
}

export type StoreCollectionsViewModel = ReturnType<typeof createStoreCollectionsViewModel>
