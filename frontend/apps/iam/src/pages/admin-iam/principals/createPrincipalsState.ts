import { createSignal, type Accessor } from 'solid-js'
import {
    listPlatformUserInlinePolicies,
    listPlatformUserPolicies,
    listTenantUserInlinePolicies,
    listTenantUserPolicies,
    type PermissionBoundary,
    type PolicyInfo,
    type UserInlinePolicy,
} from '@podzone/shared/services/iam'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'
import { prettyJSON } from '../presentation'

export function createPrincipalsState(userID: number, enabled: Accessor<boolean>) {
    const [platformUserBoundary, setPlatformUserBoundary] = createSignal<PermissionBoundary | null>(null)
    const [tenantUserBoundary, setTenantUserBoundary] = createSignal<PermissionBoundary | null>(null)
    const [principalMode, setPrincipalMode] = createSignal<'platform' | 'tenant'>('platform')
    const [principalPlatformUserId, setPrincipalPlatformUserId] = createSignal(userID ? String(userID) : '')
    const [principalTenantId, setPrincipalTenantId] = createSignal('')
    const [principalTenantUserId, setPrincipalTenantUserId] = createSignal('')
    const platformUserID = () => Number.parseInt(principalPlatformUserId().trim(), 10)
    const tenantUserID = () => Number.parseInt(principalTenantUserId().trim(), 10)
    const collectionEnabled = () =>
        enabled() &&
        (principalMode() === 'platform'
            ? Number.isFinite(platformUserID()) && platformUserID() > 0
            : principalTenantId().trim().length > 0 && Number.isFinite(tenantUserID()) && tenantUserID() > 0)
    const managedPolicies = createPaginatedResource<PolicyInfo>(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result =
                principalMode() === 'platform'
                    ? await listPlatformUserPolicies(platformUserID(), query)
                    : await listTenantUserPolicies(principalTenantId().trim(), tenantUserID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled: collectionEnabled }
    )
    const inlinePolicies = createPaginatedResource<UserInlinePolicy>(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result =
                principalMode() === 'platform'
                    ? await listPlatformUserInlinePolicies(platformUserID(), query)
                    : await listTenantUserInlinePolicies(principalTenantId().trim(), tenantUserID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled: collectionEnabled }
    )
    const [principalManagedPolicyName, setPrincipalManagedPolicyName] = createSignal('')
    const [principalBoundaryPolicyName, setPrincipalBoundaryPolicyName] = createSignal('')
    const [principalInlinePolicyName, setPrincipalInlinePolicyName] = createSignal('')
    const [principalInlinePolicyDescription, setPrincipalInlinePolicyDescription] = createSignal('')
    const [principalInlinePolicyJson, setPrincipalInlinePolicyJson] = createSignal(
        prettyJSON([
            {
                effect: 'allow',
                actionPattern: 'order:read',
                resourcePattern: '*',
                conditions: [],
            },
        ])
    )

    return {
        principalManagedPolicies: managedPolicies.items,
        principalManagedPoliciesQuery: managedPolicies.query,
        principalManagedPoliciesPageInfo: managedPolicies.pageInfo,
        principalManagedPoliciesLoading: managedPolicies.loading,
        principalManagedPoliciesError: managedPolicies.error,
        updatePrincipalManagedPoliciesQuery: managedPolicies.updateQuery,
        reloadPrincipalManagedPolicies: managedPolicies.reload,
        clearPrincipalManagedPolicies: managedPolicies.clear,
        principalInlinePolicies: inlinePolicies.items,
        principalInlinePoliciesQuery: inlinePolicies.query,
        principalInlinePoliciesPageInfo: inlinePolicies.pageInfo,
        principalInlinePoliciesLoading: inlinePolicies.loading,
        principalInlinePoliciesError: inlinePolicies.error,
        updatePrincipalInlinePoliciesQuery: inlinePolicies.updateQuery,
        reloadPrincipalInlinePolicies: inlinePolicies.reload,
        clearPrincipalInlinePolicies: inlinePolicies.clear,
        platformUserBoundary,
        setPlatformUserBoundary,
        tenantUserBoundary,
        setTenantUserBoundary,
        principalMode,
        setPrincipalMode,
        principalPlatformUserId,
        setPrincipalPlatformUserId,
        principalTenantId,
        setPrincipalTenantId,
        principalTenantUserId,
        setPrincipalTenantUserId,
        principalManagedPolicyName,
        setPrincipalManagedPolicyName,
        principalBoundaryPolicyName,
        setPrincipalBoundaryPolicyName,
        principalInlinePolicyName,
        setPrincipalInlinePolicyName,
        principalInlinePolicyDescription,
        setPrincipalInlinePolicyDescription,
        principalInlinePolicyJson,
        setPrincipalInlinePolicyJson,
    }
}
