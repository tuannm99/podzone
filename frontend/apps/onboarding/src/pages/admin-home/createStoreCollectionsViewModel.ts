import { createResource, type Accessor } from 'solid-js'
import { ensureActiveTenant } from '@podzone/shared/services/auth'
import { listStoreRequests } from '@podzone/shared/services/onboarding'
import { listStores } from '@podzone/shared/services/store'
import { useAuthContext } from '@podzone/shared/auth'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'

export function createStoreCollectionsViewModel(
    selectedWorkspaceID: Accessor<string>,
    loadingWorkspaces: Accessor<boolean>
) {
    const auth = useAuthContext()
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
            auth.setActiveTenantId(workspaceID)
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
