import { createResource, type Accessor } from 'solid-js'
import { getStoreReadiness, type StoreReadinessUiState } from '@podzone/shared/services/onboarding'

export function createStoreReadinessViewModel(tenantId: Accessor<string>, requestIds: Accessor<string[]>) {
    const [readinessMap] = createResource(
        () => {
            const tenant = tenantId().trim()
            const ids = requestIds()
            return tenant && ids.length > 0 ? { tenant, ids } : undefined
        },
        async ({ tenant, ids }) => {
            const entries = await Promise.all(
                ids.map(async (id) => {
                    const result = await getStoreReadiness(tenant, id)
                    return [id, result.success ? result.data.ui_state : undefined] as const
                })
            )
            return Object.fromEntries(entries) as Record<string, StoreReadinessUiState | undefined>
        }
    )

    return {
        uiStateFor: (requestId: string): StoreReadinessUiState | undefined => readinessMap()?.[requestId],
    }
}

export type StoreReadinessViewModel = ReturnType<typeof createStoreReadinessViewModel>
