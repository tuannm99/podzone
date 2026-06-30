import { createSignal } from 'solid-js'
import type {
  GroupInfo,
  GroupInlinePolicy,
  OrganizationInfo,
  PermissionBoundary,
  PolicyAttachmentInfo,
  PolicyInfo,
  PolicyVersionInfo,
  RolePermissionBoundary,
  SimulateAccessResult,
  TenantMembership,
  UserInlinePolicy,
} from '@/services/iam'
import {
  platformRoleOptions,
  prettyJSON,
  tenantRoleOptions,
} from './presentation'

export function createAdminIamState(userID: number) {
  const [pageError, setPageError] = createSignal('')
  const [pageMessage, setPageMessage] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [allowed, setAllowed] = createSignal(false)
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([])
  const [organizations, setOrganizations] = createSignal<OrganizationInfo[]>([])
  const [policies, setPolicies] = createSignal<PolicyInfo[]>([])
  const [groups, setGroups] = createSignal<GroupInfo[]>([])
  const [selectedOrgId, setSelectedOrgId] = createSignal('')
  const [selectedPolicyName, setSelectedPolicyName] = createSignal('')
  const [selectedGroupId, setSelectedGroupId] = createSignal('')
  const [policyDetail, setPolicyDetail] = createSignal<PolicyInfo>()
  const [policyVersions, setPolicyVersions] = createSignal<PolicyVersionInfo[]>(
    []
  )
  const [policyAttachments, setPolicyAttachments] = createSignal<
    PolicyAttachmentInfo[]
  >([])
  const [orgPolicies, setOrgPolicies] = createSignal<PolicyInfo[]>([])
  const [groupMembers, setGroupMembers] = createSignal<number[]>([])
  const [groupPolicies, setGroupPolicies] = createSignal<PolicyInfo[]>([])
  const [groupInlinePolicies, setGroupInlinePolicies] = createSignal<
    GroupInlinePolicy[]
  >([])
  const [roleBoundary, setRoleBoundary] =
    createSignal<RolePermissionBoundary | null>(null)
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
  const [simulation, setSimulation] = createSignal<SimulateAccessResult>()
  const [orgName, setOrgName] = createSignal('')
  const [orgSlug, setOrgSlug] = createSignal('')
  const [orgTenantId, setOrgTenantId] = createSignal('')
  const [orgPolicyName, setOrgPolicyName] = createSignal('')
  const [policyScope, setPolicyScope] = createSignal('platform')
  const [policyName, setPolicyName] = createSignal('')
  const [policyDescription, setPolicyDescription] = createSignal('')
  const [policyStatementsJson, setPolicyStatementsJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const [policyVersionJson, setPolicyVersionJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:update',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const [groupScope, setGroupScope] = createSignal('platform')
  const [groupTenantId, setGroupTenantId] = createSignal('')
  const [groupName, setGroupName] = createSignal('')
  const [groupDescription, setGroupDescription] = createSignal('')
  const [groupMemberUserId, setGroupMemberUserId] = createSignal('')
  const [groupPolicyName, setGroupPolicyName] = createSignal('')
  const [groupInlinePolicyName, setGroupInlinePolicyName] = createSignal('')
  const [groupInlinePolicyDescription, setGroupInlinePolicyDescription] =
    createSignal('')
  const [groupInlinePolicyJson, setGroupInlinePolicyJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const [shortcutPlatformUserId, setShortcutPlatformUserId] = createSignal(
    userID ? String(userID) : ''
  )
  const [shortcutPlatformRoleName, setShortcutPlatformRoleName] = createSignal(
    platformRoleOptions[1].value
  )
  const [shortcutTenantId, setShortcutTenantId] = createSignal('')
  const [shortcutTenantUserId, setShortcutTenantUserId] = createSignal('')
  const [shortcutTenantRoleName, setShortcutTenantRoleName] = createSignal(
    tenantRoleOptions[1].value
  )
  const [trustRoleName, setTrustRoleName] = createSignal('tenant_admin')
  const [trustBoundaryPolicyName, setTrustBoundaryPolicyName] = createSignal('')
  const [trustJson, setTrustJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        principalType: 'service',
        principalPattern: 'backoffice.podzone.internal',
        tenantPattern: '*',
        externalIdPattern: '',
      },
    ])
  )
  const [simScope, setSimScope] = createSignal('tenant')
  const [simTenantId, setSimTenantId] = createSignal('')
  const [simTargetUserId, setSimTargetUserId] = createSignal(
    userID ? String(userID) : ''
  )
  const [simAction, setSimAction] = createSignal('order:update')
  const [simResource, setSimResource] = createSignal('*')
  const [simServicePrincipal, setSimServicePrincipal] = createSignal('')
  const [simAttributesJson, setSimAttributesJson] = createSignal(prettyJSON({}))
  const [simSessionTagsJson, setSimSessionTagsJson] = createSignal(
    prettyJSON({ team: 'ops', lane: 'priority' })
  )
  const [simSessionPolicyJson, setSimSessionPolicyJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:update',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const [simAssumedRoleId, setSimAssumedRoleId] = createSignal('')
  const [simAssumedRoleScope, setSimAssumedRoleScope] = createSignal('tenant')
  const [simAssumedRoleName, setSimAssumedRoleName] = createSignal('')
  const [simAssumedRoleTenantId, setSimAssumedRoleTenantId] = createSignal('')
  const [simAssumedRoleSessionName, setSimAssumedRoleSessionName] =
    createSignal('')
  const [simAssumedRoleSourceIdentity, setSimAssumedRoleSourceIdentity] =
    createSignal('')
  const [simAssumedRoleServicePrincipal, setSimAssumedRoleServicePrincipal] =
    createSignal('')
  const [simAssumedRoleExpiresAt, setSimAssumedRoleExpiresAt] = createSignal('')
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

  const tenantOptions = () =>
    memberships().map((membership) => ({
      name: `${membership.tenantId} · ${membership.roleName}`,
      value: membership.tenantId,
    }))
  const policyOptions = () =>
    policies().map((item) => ({
      name: `${item.name} · ${item.scope}`,
      value: item.name,
    }))
  const organizationOptions = () =>
    organizations().map((item) => ({
      name: `${item.slug} · ${item.name}`,
      value: item.id,
    }))
  const groupOptions = () =>
    groups().map((item) => ({
      name: `${item.name}${item.tenantId ? ` · ${item.tenantId}` : ''}`,
      value: String(item.id || ''),
    }))

  return {
    pageError,
    setPageError,
    pageMessage,
    setPageMessage,
    loading,
    setLoading,
    allowed,
    setAllowed,
    memberships,
    setMemberships,
    organizations,
    setOrganizations,
    policies,
    setPolicies,
    groups,
    setGroups,
    selectedOrgId,
    setSelectedOrgId,
    selectedPolicyName,
    setSelectedPolicyName,
    selectedGroupId,
    setSelectedGroupId,
    policyDetail,
    setPolicyDetail,
    policyVersions,
    setPolicyVersions,
    policyAttachments,
    setPolicyAttachments,
    orgPolicies,
    setOrgPolicies,
    groupMembers,
    setGroupMembers,
    groupPolicies,
    setGroupPolicies,
    groupInlinePolicies,
    setGroupInlinePolicies,
    roleBoundary,
    setRoleBoundary,
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
    simulation,
    setSimulation,
    orgName,
    setOrgName,
    orgSlug,
    setOrgSlug,
    orgTenantId,
    setOrgTenantId,
    orgPolicyName,
    setOrgPolicyName,
    policyScope,
    setPolicyScope,
    policyName,
    setPolicyName,
    policyDescription,
    setPolicyDescription,
    policyStatementsJson,
    setPolicyStatementsJson,
    policyVersionJson,
    setPolicyVersionJson,
    groupScope,
    setGroupScope,
    groupTenantId,
    setGroupTenantId,
    groupName,
    setGroupName,
    groupDescription,
    setGroupDescription,
    groupMemberUserId,
    setGroupMemberUserId,
    groupPolicyName,
    setGroupPolicyName,
    groupInlinePolicyName,
    setGroupInlinePolicyName,
    groupInlinePolicyDescription,
    setGroupInlinePolicyDescription,
    groupInlinePolicyJson,
    setGroupInlinePolicyJson,
    shortcutPlatformUserId,
    setShortcutPlatformUserId,
    shortcutPlatformRoleName,
    setShortcutPlatformRoleName,
    shortcutTenantId,
    setShortcutTenantId,
    shortcutTenantUserId,
    setShortcutTenantUserId,
    shortcutTenantRoleName,
    setShortcutTenantRoleName,
    trustRoleName,
    setTrustRoleName,
    trustBoundaryPolicyName,
    setTrustBoundaryPolicyName,
    trustJson,
    setTrustJson,
    simScope,
    setSimScope,
    simTenantId,
    setSimTenantId,
    simTargetUserId,
    setSimTargetUserId,
    simAction,
    setSimAction,
    simResource,
    setSimResource,
    simServicePrincipal,
    setSimServicePrincipal,
    simAttributesJson,
    setSimAttributesJson,
    simSessionTagsJson,
    setSimSessionTagsJson,
    simSessionPolicyJson,
    setSimSessionPolicyJson,
    simAssumedRoleId,
    setSimAssumedRoleId,
    simAssumedRoleScope,
    setSimAssumedRoleScope,
    simAssumedRoleName,
    setSimAssumedRoleName,
    simAssumedRoleTenantId,
    setSimAssumedRoleTenantId,
    simAssumedRoleSessionName,
    setSimAssumedRoleSessionName,
    simAssumedRoleSourceIdentity,
    setSimAssumedRoleSourceIdentity,
    simAssumedRoleServicePrincipal,
    setSimAssumedRoleServicePrincipal,
    simAssumedRoleExpiresAt,
    setSimAssumedRoleExpiresAt,
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
    tenantOptions,
    policyOptions,
    organizationOptions,
    groupOptions,
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
