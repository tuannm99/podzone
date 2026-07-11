import { createEffect, createResource, createSignal, on, type Accessor } from 'solid-js'
import { listTenantInvites, listTenantMembers, listUserTenants, type TenantMembership } from '@podzone/shared/services/iam'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'

export function createWorkspaceAccessViewModel(
    userID: number,
    activeTenantID: Accessor<string>,
    enabled: Accessor<boolean>,
    options?: {
        membersEnabled?: Accessor<boolean>
        invitesEnabled?: Accessor<boolean>
    }
) {
    const membersEnabled = options?.membersEnabled ?? enabled
    const invitesEnabled = options?.invitesEnabled ?? enabled
    const [selectedTenantID, setSelectedTenantID] = createSignal(activeTenantID())
    const [membershipsResource, { refetch: reloadMemberships }] = createResource(
        () => (enabled() && userID ? userID : undefined),
        async (currentUserID): Promise<TenantMembership[]> => {
            const result = await listUserTenants(currentUserID)
            if (!result.success) throw new Error(result.message)
            return result.data
        }
    )
    const memberships = () => membershipsResource.latest || []
    const members = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listTenantMembers(selectedTenantID().trim(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: () => membersEnabled() && Boolean(userID && selectedTenantID().trim()),
            dependency: selectedTenantID,
        }
    )
    const invites = createPaginatedResource(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listTenantInvites(selectedTenantID().trim(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: () => invitesEnabled() && Boolean(userID && selectedTenantID().trim()),
            dependency: selectedTenantID,
        }
    )
    const canRead = () => members.resolved()
    const canManage = () => invites.resolved()
    const loadingMemberships = () => membershipsResource.loading
    const loadingAccess = () => members.loading() || invites.loading()
    const error = () => {
        const resourceError = membershipsResource.error
        return resourceError instanceof Error ? resourceError.message : ''
    }
    const tenantOptions = () =>
        memberships().map((membership) => ({
            name: `${membership.tenantId} · ${membership.roleName}`,
            value: membership.tenantId,
        }))

    createEffect(() => {
        if (selectedTenantID().trim()) return
        const tenantID = activeTenantID() || memberships()[0]?.tenantId || ''
        if (tenantID) setSelectedTenantID(tenantID)
    })
    createEffect(
        on(
            selectedTenantID,
            () => {
                members.clear()
                invites.clear()
            },
            { defer: true }
        )
    )

    return {
        selectedTenantID,
        setSelectedTenantID,
        memberships,
        members,
        invites,
        canRead,
        canManage,
        loadingMemberships,
        loadingAccess,
        error,
        tenantOptions,
        reloadMemberships: async () => void (await reloadMemberships()),
        reloadAccess: async () => {
            await Promise.all([members.reload(), invites.reload()])
        },
    }
}

export type WorkspaceAccessViewModel = ReturnType<typeof createWorkspaceAccessViewModel>
