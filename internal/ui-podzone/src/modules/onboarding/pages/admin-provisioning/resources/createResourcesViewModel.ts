import { createSignal } from 'solid-js'
import {
    checkDatabaseClusterHealth,
    deleteResource,
    listDatabaseClusters,
    listKubernetesClusters,
    listRuntimePools,
    upsertDatabaseCluster,
    upsertKubernetesCluster,
    upsertRuntimePool,
    type DatabaseClusterResource,
    type KubernetesClusterResource,
    type RuntimePoolResource,
} from '@/services/onboarding/provisioning'
import { createPaginatedResource } from '@/solid/pagination'

export type ResourceEditor =
    | { kind: 'database-clusters'; value?: DatabaseClusterResource }
    | { kind: 'kubernetes-clusters'; value?: KubernetesClusterResource }
    | { kind: 'runtime-pools'; value?: RuntimePoolResource }

export function createResourcesViewModel() {
    const [editor, setEditor] = createSignal<ResourceEditor>()
    const [saving, setSaving] = createSignal(false)
    const [deletingName, setDeletingName] = createSignal('')
    const [checkingName, setCheckingName] = createSignal('')
    const [mutationError, setMutationError] = createSignal('')
    const query = {
        page: 1,
        pageSize: 10,
        sortBy: 'updatedAt',
        sortDirection: 'SORT_DIRECTION_DESC' as const,
    }
    const databaseClusters = createPaginatedResource(query, async (next) => {
        const result = await listDatabaseClusters(next)
        if (!result.success) throw new Error(result.message)
        return result.data
    })
    const kubernetesClusters = createPaginatedResource(query, async (next) => {
        const result = await listKubernetesClusters(next)
        if (!result.success) throw new Error(result.message)
        return result.data
    })
    const runtimePools = createPaginatedResource(query, async (next) => {
        const result = await listRuntimePools(next)
        if (!result.success) throw new Error(result.message)
        return result.data
    })

    const reload = async (kind: ResourceEditor['kind']) => {
        if (kind === 'database-clusters') await databaseClusters.reload()
        if (kind === 'kubernetes-clusters') await kubernetesClusters.reload()
        if (kind === 'runtime-pools') await runtimePools.reload()
    }
    const saveDatabaseCluster = async (resource: DatabaseClusterResource) => {
        setSaving(true)
        setMutationError('')
        try {
            const result = await upsertDatabaseCluster(resource)
            if (!result.success) {
                setMutationError(result.message)
                return false
            }
            await reload('database-clusters')
            setEditor()
            return true
        } finally {
            setSaving(false)
        }
    }
    const saveKubernetesCluster = async (resource: KubernetesClusterResource) => {
        setSaving(true)
        setMutationError('')
        try {
            const result = await upsertKubernetesCluster(resource)
            if (!result.success) {
                setMutationError(result.message)
                return false
            }
            await reload('kubernetes-clusters')
            setEditor()
            return true
        } finally {
            setSaving(false)
        }
    }
    const saveRuntimePool = async (resource: RuntimePoolResource) => {
        setSaving(true)
        setMutationError('')
        try {
            const result = await upsertRuntimePool(resource)
            if (!result.success) {
                setMutationError(result.message)
                return false
            }
            await reload('runtime-pools')
            setEditor()
            return true
        } finally {
            setSaving(false)
        }
    }
    const remove = async (kind: ResourceEditor['kind'], name: string) => {
        setDeletingName(name)
        setMutationError('')
        try {
            const result = await deleteResource(kind, name)
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            await reload(kind)
            if (editor()?.value?.name === name) setEditor()
        } finally {
            setDeletingName('')
        }
    }
    const checkDatabaseHealth = async (name: string) => {
        setCheckingName(name)
        setMutationError('')
        try {
            const result = await checkDatabaseClusterHealth(name)
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            await databaseClusters.reload()
        } finally {
            setCheckingName('')
        }
    }

    return {
        databaseClusters,
        kubernetesClusters,
        runtimePools,
        editor,
        setEditor,
        saving,
        deletingName,
        checkingName,
        saveDatabaseCluster,
        saveKubernetesCluster,
        saveRuntimePool,
        remove,
        checkDatabaseHealth,
        error: () => mutationError() || databaseClusters.error() || kubernetesClusters.error() || runtimePools.error(),
    }
}

export type ResourcesViewModel = ReturnType<typeof createResourcesViewModel>
