import { createSignal, type Accessor } from 'solid-js'
import { listDirectoryUsers, listPermissions } from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'

export function createDirectoryViewModel(params: {
  allowed: Accessor<boolean>
  canManagePlatform: Accessor<boolean>
  selectedOrgId: Accessor<string>
  selectedTenantId: Accessor<string>
}) {
  const [tenantScopeId, setTenantScopeId] = createSignal(
    params.selectedTenantId()
  )
  const permissionScope = () =>
    params.canManagePlatform()
      ? { scope: 'platform' as const }
      : {
          scope: 'organization' as const,
          orgId: params.selectedOrgId(),
        }
  const permissionScopeKey = () =>
    `${permissionScope().scope}:${permissionScope().orgId || ''}`
  const permissions = createPaginatedResource(
    {
      page: 1,
      pageSize: 100,
      search: '',
      sortBy: 'name',
      sortDirection: 'SORT_DIRECTION_ASC',
    },
    async (query) => {
      const result = await listPermissions(query, permissionScope())
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () =>
        params.allowed() &&
        (params.canManagePlatform() || Boolean(params.selectedOrgId())),
      dependency: permissionScopeKey,
    }
  )

  const platformUsers = createPaginatedResource(
    {
      page: 1,
      pageSize: 20,
      search: '',
      sortBy: 'username',
      sortDirection: 'SORT_DIRECTION_ASC',
    },
    async (query) => {
      const result = await listDirectoryUsers(query, { scope: 'platform' })
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () => params.allowed() && params.canManagePlatform(),
    }
  )

  const tenantUsers = createPaginatedResource(
    {
      page: 1,
      pageSize: 20,
      search: '',
      sortBy: 'username',
      sortDirection: 'SORT_DIRECTION_ASC',
    },
    async (query) => {
      const result = await listDirectoryUsers(query, {
        scope: 'tenant',
        tenantId: tenantScopeId(),
      })
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () => params.allowed() && Boolean(tenantScopeId().trim()),
      dependency: tenantScopeId,
    }
  )

  const organizationUsers = createPaginatedResource(
    {
      page: 1,
      pageSize: 20,
      search: '',
      sortBy: 'username',
      sortDirection: 'SORT_DIRECTION_ASC',
    },
    async (query) => {
      const result = await listDirectoryUsers(query, {
        scope: 'organization',
        orgId: params.selectedOrgId(),
      })
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () => params.allowed() && Boolean(params.selectedOrgId().trim()),
      dependency: params.selectedOrgId,
    }
  )

  const userOptions = (users: ReturnType<typeof platformUsers.items>) =>
    users.map((user) => ({
      value: String(user.id),
      label:
        user.displayName || user.username || user.email || `User ${user.id}`,
      description: [user.email, `ID ${user.id}`].filter(Boolean).join(' · '),
    }))

  return {
    permissionOptions: () =>
      permissions.items().map((permission) => ({
        name: `${permission.name} — ${permission.resource}:${permission.action}`,
        value: permission.name,
      })),
    permissionsLoading: permissions.loading,
    permissionsError: permissions.error,
    platformUserOptions: () => userOptions(platformUsers.items()),
    platformUsersLoading: platformUsers.loading,
    platformUsersError: platformUsers.error,
    searchPlatformUsers: (search: string) =>
      platformUsers.updateQuery({ search }),
    tenantUserOptions: () => userOptions(tenantUsers.items()),
    tenantUsersLoading: tenantUsers.loading,
    tenantUsersError: tenantUsers.error,
    searchTenantUsers: (search: string, tenantId?: string) => {
      setTenantScopeId(tenantId?.trim() || params.selectedTenantId().trim())
      tenantUsers.updateQuery({ search })
    },
    organizationUserOptions: () => userOptions(organizationUsers.items()),
    organizationUsersLoading: organizationUsers.loading,
    organizationUsersError: organizationUsers.error,
    searchOrganizationUsers: (search: string) =>
      organizationUsers.updateQuery({ search }),
  }
}
