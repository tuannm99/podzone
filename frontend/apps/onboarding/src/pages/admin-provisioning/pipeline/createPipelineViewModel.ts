import { createEffect, createSignal, type Accessor } from 'solid-js'
import { listStoreRequests, retryStoreRequest } from '@podzone/shared/services/onboarding'
import { listStoreRequestTransitions } from '@podzone/shared/services/onboarding/provisioning'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'

export function createPipelineViewModel(tenantId: Accessor<string>, enabled: Accessor<boolean>) {
    const [selectedRequestId, setSelectedRequestId] = createSignal('')
    const [retryingId, setRetryingId] = createSignal('')
    const [mutationError, setMutationError] = createSignal('')
    const requests = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'updatedAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listStoreRequests(tenantId(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled, dependency: tenantId }
    )
    const transitions = createPaginatedResource(
        {
            page: 1,
            pageSize: 100,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_ASC',
        },
        async (query) => {
            const result = await listStoreRequestTransitions(tenantId(), selectedRequestId(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: () => enabled() && Boolean(selectedRequestId()),
            dependency: () => `${tenantId()}:${selectedRequestId()}`,
        }
    )

    createEffect(() => {
        if (requests.loading()) return
        const items = requests.items()
        if (items.length === 0) return
        if (items.some((request) => request.id === selectedRequestId())) return
        setSelectedRequestId(items[0]?.id || '')
    })

    const selectedRequest = () => requests.items().find((request) => request.id === selectedRequestId())
    const retry = async (requestId: string) => {
        setRetryingId(requestId)
        setMutationError('')
        try {
            const result = await retryStoreRequest({
                tenantId: tenantId(),
                requestId,
            })
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            await Promise.all([requests.reload(), transitions.reload()])
        } finally {
            setRetryingId('')
        }
    }

    return {
        requests,
        transitions,
        selectedRequestId,
        setSelectedRequestId,
        selectedRequest,
        retryingId,
        retry,
        error: () => mutationError() || requests.error() || transitions.error(),
    }
}

export type PipelineViewModel = ReturnType<typeof createPipelineViewModel>
