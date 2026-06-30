import {
  createEffect,
  createResource,
  createSignal,
  type Accessor,
} from 'solid-js'
import {
  checkPermission,
  listTenantInvites,
  listTenantMembers,
  listUserTenants,
  type TenantMembership,
} from '@/services/iam'

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

      const [membersResult, invitesResult] = await Promise.all([
        readResult.data ? listTenantMembers(tenantID) : undefined,
        manageResult.data ? listTenantInvites(tenantID) : undefined,
      ])
      if (membersResult && !membersResult.success) {
        throw new Error(membersResult.message)
      }
      if (invitesResult && !invitesResult.success) {
        throw new Error(invitesResult.message)
      }
      return {
        canRead: readResult.data,
        canManage: manageResult.data,
        members: membersResult?.success ? membersResult.data : [],
        invites: invitesResult?.success ? invitesResult.data : [],
      }
    }
  )

  const memberships = () => membershipsResource.latest || []
  const members = () => accessResource.latest?.members || []
  const invites = () => accessResource.latest?.invites || []
  const canRead = () => accessResource.latest?.canRead || false
  const canManage = () => accessResource.latest?.canManage || false
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
