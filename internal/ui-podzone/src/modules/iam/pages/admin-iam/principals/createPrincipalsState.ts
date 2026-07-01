import { createSignal } from 'solid-js'
import type {
  PermissionBoundary,
  PolicyInfo,
  UserInlinePolicy,
} from '@/services/iam'
import { prettyJSON } from '../presentation'

export function createPrincipalsState(userID: number) {
  const [platformUserPolicies, setPlatformUserPolicies] = createSignal<
    PolicyInfo[]
  >([])
  const [tenantUserPolicies, setTenantUserPolicies] = createSignal<
    PolicyInfo[]
  >([])
  const [platformUserInlinePolicies, setPlatformUserInlinePolicies] =
    createSignal<UserInlinePolicy[]>([])
  const [tenantUserInlinePolicies, setTenantUserInlinePolicies] = createSignal<
    UserInlinePolicy[]
  >([])
  const [platformUserBoundary, setPlatformUserBoundary] =
    createSignal<PermissionBoundary | null>(null)
  const [tenantUserBoundary, setTenantUserBoundary] =
    createSignal<PermissionBoundary | null>(null)
  const [principalMode, setPrincipalMode] = createSignal<'platform' | 'tenant'>(
    'platform'
  )
  const [principalPlatformUserId, setPrincipalPlatformUserId] = createSignal(
    userID ? String(userID) : ''
  )
  const [principalTenantId, setPrincipalTenantId] = createSignal('')
  const [principalTenantUserId, setPrincipalTenantUserId] = createSignal('')
  const [principalManagedPolicyName, setPrincipalManagedPolicyName] =
    createSignal('')
  const [principalBoundaryPolicyName, setPrincipalBoundaryPolicyName] =
    createSignal('')
  const [principalInlinePolicyName, setPrincipalInlinePolicyName] =
    createSignal('')
  const [
    principalInlinePolicyDescription,
    setPrincipalInlinePolicyDescription,
  ] = createSignal('')
  const [principalInlinePolicyJson, setPrincipalInlinePolicyJson] =
    createSignal(
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
    platformUserPolicies,
    setPlatformUserPolicies,
    tenantUserPolicies,
    setTenantUserPolicies,
    platformUserInlinePolicies,
    setPlatformUserInlinePolicies,
    tenantUserInlinePolicies,
    setTenantUserInlinePolicies,
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
