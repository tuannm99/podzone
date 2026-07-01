import { createEffect, onMount } from 'solid-js'
import { tokenStorage } from '@/services/tokenStorage'
import {
  attachmentColor,
  groupScopeOptions,
  platformRoleOptions,
  policyScopeOptions,
  sectionLinks,
  simulationLayerTone,
  simulationSourceColor,
  statementSourceLabel,
  tenantRoleOptions,
} from './presentation'
import { createAdminIamActions } from './createAdminIamActions'
import { createAdminIamLoaders } from './createAdminIamLoaders'
import { createAdminIamState } from './createAdminIamState'

export function createAdminIamViewModel() {
  const userID = tokenStorage.getUserID() || 0
  const state = createAdminIamState(userID)
  const loaders = createAdminIamLoaders(state, userID)
  const actions = createAdminIamActions(state, loaders)

  createEffect(() => {
    const firstPolicy = state.policies()[0]
    if (state.allowed() && !state.selectedPolicyName() && firstPolicy) {
      state.setSelectedPolicyName(firstPolicy.name)
    }
  })

  createEffect(() => {
    const firstOrganization = state.organizations()[0]
    if (state.allowed() && !state.selectedOrgId() && firstOrganization) {
      state.setSelectedOrgId(firstOrganization.id)
    }
  })

  createEffect(() => {
    void state.selectedPolicyName()
    if (state.allowed()) void loaders.loadSelectedPolicy()
  })

  createEffect(() => {
    void state.selectedOrgId()
    if (state.allowed()) void loaders.loadSelectedOrganization()
  })

  createEffect(() => {
    void state.selectedGroupId()
    if (state.allowed()) void loaders.loadSelectedGroup()
  })

  createEffect(() => {
    void state.groupScope()
    void state.groupTenantId()
    if (state.allowed()) void loaders.loadGroupsForScope()
  })

  onMount(() => {
    void loaders.loadBootstrap()
  })

  const policyContextValue = {
    policyScopeOptions,
    policyScope: state.policyScope,
    setPolicyScope: state.setPolicyScope,
    policyName: state.policyName,
    setPolicyName: state.setPolicyName,
    policyDescription: state.policyDescription,
    setPolicyDescription: state.setPolicyDescription,
    policyStatementsJson: state.policyStatementsJson,
    setPolicyStatementsJson: state.setPolicyStatementsJson,
    policyVersionJson: state.policyVersionJson,
    setPolicyVersionJson: state.setPolicyVersionJson,
    selectedPolicyName: state.selectedPolicyName,
    setSelectedPolicyName: state.setSelectedPolicyName,
    policyOptions: state.policyOptions,
    policies: state.policies,
    query: state.policiesQuery,
    pageInfo: state.policiesPageInfo,
    loading: state.policiesLoading,
    error: state.policiesError,
    updateQuery: state.updatePoliciesQuery,
    policyDetail: state.policyDetail,
    policyVersions: state.policyVersions,
    policyAttachments: state.policyAttachments,
    attachmentColor,
    submitCreatePolicy: actions.submitCreatePolicy,
    createPolicyFromForm: actions.createPolicyFromForm,
    createPolicyVersionFromForm: actions.createPolicyVersionFromForm,
    handleCreatePolicyVersion: actions.handleCreatePolicyVersion,
    handleDeletePolicy: actions.handleDeletePolicy,
    handleSetDefaultVersion: actions.handleSetDefaultVersion,
    handleDeleteVersion: actions.handleDeleteVersion,
  }

  const groupContextValue = {
    groupScopeOptions,
    groupScope: state.groupScope,
    setGroupScope: state.setGroupScope,
    groupTenantId: state.groupTenantId,
    setGroupTenantId: state.setGroupTenantId,
    tenantOptions: state.tenantOptions,
    groupName: state.groupName,
    setGroupName: state.setGroupName,
    groupDescription: state.groupDescription,
    setGroupDescription: state.setGroupDescription,
    submitCreateGroup: actions.submitCreateGroup,
    createGroupFromForm: actions.createGroupFromForm,
    groupOptions: state.groupOptions,
    groups: state.groups,
    query: state.groupsQuery,
    pageInfo: state.groupsPageInfo,
    loading: state.groupsLoading,
    error: state.groupsError,
    updateQuery: state.updateGroupsQuery,
    selectedGroupId: state.selectedGroupId,
    setSelectedGroupId: state.setSelectedGroupId,
    groupMemberUserId: state.groupMemberUserId,
    setGroupMemberUserId: state.setGroupMemberUserId,
    groupPolicyName: state.groupPolicyName,
    setGroupPolicyName: state.setGroupPolicyName,
    addGroupMemberFromForm: actions.addGroupMemberFromForm,
    handleAddGroupMember: actions.handleAddGroupMember,
    attachGroupPolicyFromForm: actions.attachGroupPolicyFromForm,
    handleAttachGroupPolicy: actions.handleAttachGroupPolicy,
    handleDeleteGroup: actions.handleDeleteGroup,
    groupMembers: state.groupMembers,
    handleRemoveGroupMember: actions.handleRemoveGroupMember,
    groupPolicies: state.groupPolicies,
    handleDetachGroupPolicy: actions.handleDetachGroupPolicy,
    groupInlinePolicyName: state.groupInlinePolicyName,
    setGroupInlinePolicyName: state.setGroupInlinePolicyName,
    groupInlinePolicyDescription: state.groupInlinePolicyDescription,
    setGroupInlinePolicyDescription: state.setGroupInlinePolicyDescription,
    groupInlinePolicyJson: state.groupInlinePolicyJson,
    setGroupInlinePolicyJson: state.setGroupInlinePolicyJson,
    saveGroupInlinePolicyFromForm: actions.saveGroupInlinePolicyFromForm,
    handleSaveGroupInlinePolicy: actions.handleSaveGroupInlinePolicy,
    groupInlinePolicies: state.groupInlinePolicies,
    handleDeleteGroupInlinePolicy: actions.handleDeleteGroupInlinePolicy,
  }

  const principalContextValue = {
    principalMode: state.principalMode,
    setPrincipalMode: state.setPrincipalMode,
    principalPlatformUserId: state.principalPlatformUserId,
    setPrincipalPlatformUserId: state.setPrincipalPlatformUserId,
    principalTenantId: state.principalTenantId,
    setPrincipalTenantId: state.setPrincipalTenantId,
    principalTenantUserId: state.principalTenantUserId,
    setPrincipalTenantUserId: state.setPrincipalTenantUserId,
    loadPrincipalControls: loaders.loadPrincipalControls,
    tenantOptions: state.tenantOptions,
    principalManagedPolicyName: state.principalManagedPolicyName,
    setPrincipalManagedPolicyName: state.setPrincipalManagedPolicyName,
    attachPrincipalManagedPolicyFromForm:
      actions.attachPrincipalManagedPolicyFromForm,
    handleAttachPrincipalManagedPolicy:
      actions.handleAttachPrincipalManagedPolicy,
    currentManagedPolicies: () =>
      state.principalMode() === 'platform'
        ? state.platformUserPolicies()
        : state.tenantUserPolicies(),
    handleDetachPrincipalManagedPolicy:
      actions.handleDetachPrincipalManagedPolicy,
    principalBoundaryPolicyName: state.principalBoundaryPolicyName,
    setPrincipalBoundaryPolicyName: state.setPrincipalBoundaryPolicyName,
    savePrincipalBoundaryFromForm: actions.savePrincipalBoundaryFromForm,
    handleSavePrincipalBoundary: actions.handleSavePrincipalBoundary,
    handleDeletePrincipalBoundary: actions.handleDeletePrincipalBoundary,
    currentBoundary: () =>
      state.principalMode() === 'platform'
        ? state.platformUserBoundary()
        : state.tenantUserBoundary(),
    principalInlinePolicyName: state.principalInlinePolicyName,
    setPrincipalInlinePolicyName: state.setPrincipalInlinePolicyName,
    principalInlinePolicyDescription: state.principalInlinePolicyDescription,
    setPrincipalInlinePolicyDescription:
      state.setPrincipalInlinePolicyDescription,
    principalInlinePolicyJson: state.principalInlinePolicyJson,
    setPrincipalInlinePolicyJson: state.setPrincipalInlinePolicyJson,
    savePrincipalInlinePolicyFromForm:
      actions.savePrincipalInlinePolicyFromForm,
    handleSavePrincipalInlinePolicy: actions.handleSavePrincipalInlinePolicy,
    currentInlinePolicies: () =>
      state.principalMode() === 'platform'
        ? state.platformUserInlinePolicies()
        : state.tenantUserInlinePolicies(),
    handleDeletePrincipalInlinePolicy:
      actions.handleDeletePrincipalInlinePolicy,
  }

  const trustSimContextValue = {
    trustRoleName: state.trustRoleName,
    setTrustRoleName: state.setTrustRoleName,
    loadTrustPolicy: loaders.loadTrustPolicy,
    handleSaveTrustPolicy: actions.handleSaveTrustPolicy,
    trustBoundaryPolicyName: state.trustBoundaryPolicyName,
    setTrustBoundaryPolicyName: state.setTrustBoundaryPolicyName,
    handleSaveRoleBoundary: actions.handleSaveRoleBoundary,
    handleDeleteRoleBoundary: actions.handleDeleteRoleBoundary,
    roleBoundary: state.roleBoundary,
    trustJson: state.trustJson,
    setTrustJson: state.setTrustJson,
    simScope: state.simScope,
    setSimScope: state.setSimScope,
    policyScopeOptions,
    simTargetUserId: state.simTargetUserId,
    setSimTargetUserId: state.setSimTargetUserId,
    tenantOptions: state.tenantOptions,
    simTenantId: state.simTenantId,
    setSimTenantId: state.setSimTenantId,
    simAction: state.simAction,
    setSimAction: state.setSimAction,
    simResource: state.simResource,
    setSimResource: state.setSimResource,
    simServicePrincipal: state.simServicePrincipal,
    setSimServicePrincipal: state.setSimServicePrincipal,
    applyServiceAssumePreset: actions.applyServiceAssumePreset,
    applyTenantAssumePreset: actions.applyTenantAssumePreset,
    applyScopeDownDenyPreset: actions.applyScopeDownDenyPreset,
    simAttributesJson: state.simAttributesJson,
    setSimAttributesJson: state.setSimAttributesJson,
    simSessionTagsJson: state.simSessionTagsJson,
    setSimSessionTagsJson: state.setSimSessionTagsJson,
    simSessionPolicyJson: state.simSessionPolicyJson,
    setSimSessionPolicyJson: state.setSimSessionPolicyJson,
    simAssumedRoleId: state.simAssumedRoleId,
    setSimAssumedRoleId: state.setSimAssumedRoleId,
    simAssumedRoleScope: state.simAssumedRoleScope,
    setSimAssumedRoleScope: state.setSimAssumedRoleScope,
    simAssumedRoleName: state.simAssumedRoleName,
    setSimAssumedRoleName: state.setSimAssumedRoleName,
    simAssumedRoleTenantId: state.simAssumedRoleTenantId,
    setSimAssumedRoleTenantId: state.setSimAssumedRoleTenantId,
    simAssumedRoleSessionName: state.simAssumedRoleSessionName,
    setSimAssumedRoleSessionName: state.setSimAssumedRoleSessionName,
    simAssumedRoleSourceIdentity: state.simAssumedRoleSourceIdentity,
    setSimAssumedRoleSourceIdentity: state.setSimAssumedRoleSourceIdentity,
    simAssumedRoleServicePrincipal: state.simAssumedRoleServicePrincipal,
    setSimAssumedRoleServicePrincipal: state.setSimAssumedRoleServicePrincipal,
    simAssumedRoleExpiresAt: state.simAssumedRoleExpiresAt,
    setSimAssumedRoleExpiresAt: state.setSimAssumedRoleExpiresAt,
    handleSimulate: actions.handleSimulate,
    simulation: state.simulation,
    simulationSourceColor,
    statementSourceLabel,
    simulationLayerTone,
  }

  return {
    feedback: {
      error: state.pageError,
      message: state.pageMessage,
      loading: state.loading,
      allowed: state.allowed,
    },
    sectionLinks,
    policies: policyContextValue,
    groups: groupContextValue,
    principals: principalContextValue,
    trustSimulation: trustSimContextValue,
    organizations: {
      organizationOptions: state.organizationOptions,
      selectedOrgId: state.selectedOrgId,
      setSelectedOrgId: state.setSelectedOrgId,
      submitCreateOrganization: actions.submitCreateOrganization,
      orgName: state.orgName,
      setOrgName: state.setOrgName,
      orgSlug: state.orgSlug,
      setOrgSlug: state.setOrgSlug,
      orgTenantId: state.orgTenantId,
      setOrgTenantId: state.setOrgTenantId,
      orgPolicyName: state.orgPolicyName,
      setOrgPolicyName: state.setOrgPolicyName,
      tenantOptions: state.tenantOptions,
      handleAttachTenantToOrg: actions.handleAttachTenantToOrg,
      handleAttachScp: actions.handleAttachScp,
      items: state.organizations,
      query: state.organizationsQuery,
      pageInfo: state.organizationsPageInfo,
      loading: state.organizationsLoading,
      error: state.organizationsError,
      updateQuery: state.updateOrganizationsQuery,
      orgPolicies: state.orgPolicies,
      handleDetachTenantFromOrg: actions.handleDetachTenantFromOrg,
      handleDetachScp: actions.handleDetachScp,
    },
    assignments: {
      platformUserId: state.shortcutPlatformUserId,
      setPlatformUserId: state.setShortcutPlatformUserId,
      platformRoleName: state.shortcutPlatformRoleName,
      setPlatformRoleName: state.setShortcutPlatformRoleName,
      platformRoleOptions,
      assignPlatformRole: actions.handleAssignPlatformRole,
      removePlatformRole: actions.handleRemovePlatformRoleShortcut,
      tenantId: state.shortcutTenantId,
      setTenantId: state.setShortcutTenantId,
      tenantUserId: state.shortcutTenantUserId,
      setTenantUserId: state.setShortcutTenantUserId,
      tenantRoleName: state.shortcutTenantRoleName,
      setTenantRoleName: state.setShortcutTenantRoleName,
      tenantOptions: state.tenantOptions,
      tenantRoleOptions,
      assignTenantRole: actions.handleAssignTenantRole,
      removeTenantMembership: actions.handleRemoveTenantMembershipShortcut,
    },
  }
}

export type AdminIamViewModel = ReturnType<typeof createAdminIamViewModel>
