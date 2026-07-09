import { createResource, createSignal, type Accessor } from 'solid-js'
import {
    deleteInfrastructureConnection,
    getPlacementStatus,
    listInfrastructureConnections,
    reconcilePlacement,
    upsertInfrastructureConnection,
    type InfrastructureConnection,
    type PlacementStatus,
    type UpsertInfrastructureConnection,
} from '@podzone/shared/services/onboarding/provisioning'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'

export function createConnectionsViewModel(tenantId: Accessor<string>, enabled: Accessor<boolean>) {
    const [editor, setEditor] = createSignal<InfrastructureConnection>()
    const [creating, setCreating] = createSignal(false)
    const [saving, setSaving] = createSignal(false)
    const [deletingName, setDeletingName] = createSignal('')
    const [mutationError, setMutationError] = createSignal('')
    const [placementRepairing, setPlacementRepairing] = createSignal(false)
    const [placementError, setPlacementError] = createSignal('')
    const [placementResource, { refetch: loadPlacementStatus, mutate: setPlacementStatus }] = createResource(
        () => (enabled() && tenantId() ? tenantId() : undefined),
        async (currentTenantID): Promise<PlacementStatus | undefined> => {
            setPlacementError('')
            const result = await getPlacementStatus(currentTenantID)
            if (!result.success) {
                setPlacementError(result.message)
                return undefined
            }
            return result.data
        }
    )
    const connections = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'updatedAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listInfrastructureConnections(tenantId(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled, dependency: tenantId }
    )
    const closeEditor = () => {
        setEditor()
        setCreating(false)
    }
    const save = async (connection: UpsertInfrastructureConnection) => {
        setSaving(true)
        setMutationError('')
        try {
            const result = await upsertInfrastructureConnection(tenantId(), connection)
            if (!result.success) {
                setMutationError(result.message)
                return false
            }
            await connections.reload()
            closeEditor()
            return true
        } finally {
            setSaving(false)
        }
    }
    const remove = async (connection: InfrastructureConnection) => {
        setDeletingName(connection.name)
        setMutationError('')
        try {
            const result = await deleteInfrastructureConnection(tenantId(), connection.infra_type, connection.name)
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            await connections.reload()
            closeEditor()
        } finally {
            setDeletingName('')
        }
    }
    const reconcile = async () => {
        if (!tenantId()) return
        setPlacementRepairing(true)
        setPlacementError('')
        try {
            const result = await reconcilePlacement(tenantId())
            if (!result.success) {
                setPlacementError(result.message)
                return
            }
            setPlacementStatus(result.data.status)
            await connections.reload()
        } finally {
            setPlacementRepairing(false)
        }
    }

    return {
        connections,
        placementStatus: () => placementResource.latest,
        placementLoading: () => placementResource.loading,
        placementRepairing,
        placementError,
        loadPlacementStatus,
        reconcile,
        editor,
        creating,
        openCreate: () => {
            setEditor()
            setCreating(true)
        },
        openEdit: (connection: InfrastructureConnection) => {
            setEditor(connection)
            setCreating(false)
        },
        closeEditor,
        saving,
        deletingName,
        save,
        remove,
        error: () => mutationError() || connections.error(),
    }
}

export type ConnectionsViewModel = ReturnType<typeof createConnectionsViewModel>
