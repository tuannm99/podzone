import { createEffect, createSignal, type Accessor } from 'solid-js'
import {
    deleteInfrastructureConnection,
    getPlacementStatus,
    listInfrastructureConnections,
    reconcilePlacement,
    upsertInfrastructureConnection,
    type InfrastructureConnection,
    type PlacementStatus,
    type UpsertInfrastructureConnection,
} from '@/services/onboarding/provisioning'
import { createPaginatedResource } from '@/solid/pagination'

export function createConnectionsViewModel(tenantId: Accessor<string>, enabled: Accessor<boolean>) {
    const [editor, setEditor] = createSignal<InfrastructureConnection>()
    const [creating, setCreating] = createSignal(false)
    const [saving, setSaving] = createSignal(false)
    const [deletingName, setDeletingName] = createSignal('')
    const [mutationError, setMutationError] = createSignal('')
    const [placementStatus, setPlacementStatus] = createSignal<PlacementStatus>()
    const [placementLoading, setPlacementLoading] = createSignal(false)
    const [placementRepairing, setPlacementRepairing] = createSignal(false)
    const [placementError, setPlacementError] = createSignal('')
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
    const loadPlacementStatus = async () => {
        if (!enabled() || !tenantId()) {
            setPlacementStatus()
            setPlacementError('')
            return
        }
        setPlacementLoading(true)
        setPlacementError('')
        try {
            const result = await getPlacementStatus(tenantId())
            if (!result.success) {
                setPlacementStatus()
                setPlacementError(result.message)
                return
            }
            setPlacementStatus(result.data)
        } finally {
            setPlacementLoading(false)
        }
    }
    createEffect(() => {
        const currentTenant = tenantId()
        if (!enabled() || !currentTenant) {
            setPlacementStatus()
            setPlacementError('')
            return
        }
        void loadPlacementStatus()
    })
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
        placementStatus,
        placementLoading,
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
