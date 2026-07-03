import {
  createEffect,
  createResource,
  createSignal,
  on,
  type Accessor,
} from 'solid-js'
import {
  checkPermission,
  listTenantInvites,
  listTenantMembers,
  listUserTenants,
  type TenantMembership,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'

export function createWorkspaceAccessViewModel(
  userID: number,
  activeTenantID: Accessor<string>
) {
  const [selectedTenantID, setSelectedTenantID] = createSignal(activeTenantID())
  const [membershipsResource, { refetch: reloadMemberships }] = createResource(
    () => userID || undefined,
    async (currentUserID): Promise<TenantMembership[]> => {
      const result = await listUserTenants(currentUserID)
      if (!result.success) throw new Error(result.message)
      return result.data
    }
  )
  const [accessResource, { refetch: reloadAccess }] = createResource(
    () => {
      const tenantID = selectedTenantID().trim()
      return userID && tenantID ? { userID, tenantID } : undefined
    },
    async ({ userID: currentUserID, tenantID }) => {
      const [readResult, manageResult] = await Promise.all([
        checkPermission({
          tenantId: tenantID,
          userId: currentUserID,
          permission: 'tenant:read',
        }),
        checkPermission({
          tenantId: tenantID,
          userId: currentUserID,
          permission: 'tenant:manage_members',
        }),
      ])
      if (!readResult.success) throw new Error(readResult.message)
      if (!manageResult.success) throw new Error(manageResult.message)

      return {
        canRead: readResult.data,
        canManage: manageResult.data,
      }
    }
  )

  const memberships = () => membershipsResource.latest || []
  const canRead = () =>
    !accessResource.loading && Boolean(accessResource.latest?.canRead)
  const canManage = () =>
    !accessResource.loading && Boolean(accessResource.latest?.canManage)
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
      enabled: () => canRead() && Boolean(selectedTenantID().trim()),
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
      enabled: () => canManage() && Boolean(selectedTenantID().trim()),
      dependency: selectedTenantID,
    }
  )
  const loadingMemberships = () => membershipsResource.loading
  const loadingAccess = () => accessResource.loading
  const error = () => {
    const resourceError = membershipsResource.error || accessResource.error
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
    reloadAccess: async () => void (await reloadAccess()),
  }
}

export type WorkspaceAccessViewModel = ReturnType<
  typeof createWorkspaceAccessViewModel
>
