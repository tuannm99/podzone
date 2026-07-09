import { createMemo, createResource, createSignal, type Accessor } from 'solid-js'
import { listDirectoryUsers, listAllPermissions, listPolicies } from '@podzone/shared/services/iam'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'

export type DirectoryIdentity = {
    label: string
    description: string
}

export function createDirectoryViewModel(params: {
    allowed: Accessor<boolean>
    canManagePlatform: Accessor<boolean>
    selectedOrgId: Accessor<string>
    selectedTenantId: Accessor<string>
    visibleUserIds: Accessor<Array<number | string | undefined>>
}) {
    const [tenantScopeId, setTenantScopeId] = createSignal(params.selectedTenantId())
    const permissionScope = () =>
        params.canManagePlatform()
            ? { scope: 'platform' as const }
            : {
                  scope: 'organization' as const,
                  orgId: params.selectedOrgId(),
              }
    const permissionScopeKey = () => `${permissionScope().scope}:${permissionScope().orgId || ''}`
    const permissions = createPaginatedResource(
        {
            page: 1,
            pageSize: 100,
            search: '',
            sortBy: 'name',
            sortDirection: 'SORT_DIRECTION_ASC',
        },
        async (query) => {
            const result = await listAllPermissions(permissionScope())
            if (!result.success) throw new Error(result.message)
            return {
                items: result.data,
                pageInfo: {
                    total: result.data.length,
                    page: query.page,
                    pageSize: result.data.length || query.pageSize,
                    totalPages: result.data.length > 0 ? 1 : 0,
                    hasNext: false,
                    hasPrevious: false,
                },
            }
        },
        {
            enabled: () => params.allowed() && (params.canManagePlatform() || Boolean(params.selectedOrgId())),
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

    const createPolicyDirectory = (scope: 'platform' | 'organization' | 'tenant') =>
        createPaginatedResource(
            {
                page: 1,
                pageSize: 100,
                sortBy: 'name',
                sortDirection: 'SORT_DIRECTION_ASC' as const,
            },
            async (query) => {
                const result = await listPolicies(
                    query,
                    scope,
                    scope === 'organization' ? params.selectedOrgId() : undefined
                )
                if (!result.success) throw new Error(result.message)
                return result.data
            },
            {
                enabled: () =>
                    params.allowed() &&
                    (scope === 'organization' ? Boolean(params.selectedOrgId().trim()) : params.canManagePlatform()),
                dependency: scope === 'organization' ? params.selectedOrgId : undefined,
            }
        )
    const platformPolicies = createPolicyDirectory('platform')
    const organizationPolicies = createPolicyDirectory('organization')
    const tenantPolicies = createPolicyDirectory('tenant')
    const policyDirectory = {
        platform: platformPolicies,
        organization: organizationPolicies,
        tenant: tenantPolicies,
    }

    const visibleUserIDs = createMemo(() => [
        ...new Set(
            params
                .visibleUserIds()
                .map((userID) => String(userID || '').trim())
                .filter(Boolean)
        ),
    ])
    const visibleIdentityRequest = () => {
        const userIDs = visibleUserIDs()
        if (!params.allowed() || userIDs.length === 0) return undefined
        const scope = params.canManagePlatform()
            ? ({ scope: 'platform' } as const)
            : ({
                  scope: 'organization',
                  orgId: params.selectedOrgId(),
              } as const)
        if (scope.scope === 'organization' && !scope.orgId.trim()) return undefined
        return {
            key: `${scope.scope}:${scope.scope === 'organization' ? scope.orgId : ''}:${userIDs.join(',')}`,
            scope,
            userIDs,
        }
    }
    const [visibleIdentities] = createResource(visibleIdentityRequest, async (request) => {
        const result = await listDirectoryUsers(
            {
                page: 1,
                pageSize: Math.min(request.userIDs.length, 100),
                filters: [
                    {
                        field: 'id',
                        operator: 'FILTER_OPERATOR_IN',
                        values: request.userIDs,
                    },
                ],
                sortBy: 'username',
                sortDirection: 'SORT_DIRECTION_ASC',
            },
            request.scope
        )
        if (!result.success) throw new Error(result.message)
        return result.data.items
    })
    const identityByID = createMemo(() => {
        const identities = new Map<string, DirectoryIdentity>()
        for (const user of visibleIdentities.latest || []) {
            identities.set(String(user.id), {
                label: user.displayName || user.username || user.email || `User ${user.id}`,
                description: [user.email, `ID ${user.id}`].filter(Boolean).join(' · '),
            })
        }
        return identities
    })

    const userOptions = (users: ReturnType<typeof platformUsers.items>) =>
        users.map((user) => ({
            value: String(user.id),
            label: user.displayName || user.username || user.email || `User ${user.id}`,
            description: [user.email, `ID ${user.id}`].filter(Boolean).join(' · '),
        }))

    return {
        permissionOptions: () =>
            permissions.items().map((permission) => ({
                name: `${permission.name} — ${permission.resource}:${permission.action}`,
                value: permission.name,
                resource: permission.resource,
                action: permission.action,
            })),
        permissionsLoading: permissions.loading,
        permissionsError: permissions.error,
        platformUserOptions: () => userOptions(platformUsers.items()),
        platformUsersLoading: platformUsers.loading,
        platformUsersError: platformUsers.error,
        searchPlatformUsers: (search: string) => platformUsers.updateQuery({ search }),
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
        searchOrganizationUsers: (search: string) => organizationUsers.updateQuery({ search }),
        managedPolicyOptions: (scope: string) => {
            const policies = policyDirectory[scope as keyof typeof policyDirectory] || platformPolicies
            return [
                { name: 'Choose a managed policy', value: '' },
                ...policies.items().map((policy) => ({
                    name: `${policy.name} · ${policy.scope}`,
                    value: policy.name,
                })),
            ]
        },
        identityForUser: (userID: number | string): DirectoryIdentity =>
            identityByID().get(String(userID)) || {
                label: `User ${userID}`,
                description: `ID ${userID}`,
            },
        visibleIdentitiesLoading: () => visibleIdentities.loading,
    }
}
