import { createEffect, onMount } from 'solid-js'
import { tokenStorage } from '@/services/tokenStorage'
import { AdminIamView } from './admin-iam/AdminIamView'
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
} from './admin-iam/presentation'
import { useAdminIamActions } from './admin-iam/useAdminIamActions'
import { useAdminIamLoaders } from './admin-iam/useAdminIamLoaders'
import { useAdminIamState } from './admin-iam/useAdminIamState'

export function createAdminIamViewModel() {
  const userID = tokenStorage.getUserID() || 0
  const state = useAdminIamState(userID)
  const loaders = useAdminIamLoaders(state, userID)
  const actions = useAdminIamActions(state, loaders)

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

  createEffect(() => {
    void state.principalMode()
    void state.principalPlatformUserId()
    void state.principalTenantId()
    void state.principalTenantUserId()
    if (state.allowed()) void loaders.loadPrincipalControls()
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
    pageError: state.pageError,
    pageMessage: state.pageMessage,
    loading: state.loading,
    allowed: state.allowed,
    sectionLinks,
    policyContextValue,
    groupContextValue,
    principalContextValue,
    trustSimContextValue,
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
    organizations: state.organizations,
    orgPolicies: state.orgPolicies,
    handleDetachTenantFromOrg: actions.handleDetachTenantFromOrg,
    handleDetachScp: actions.handleDetachScp,
    shortcutPlatformUserId: state.shortcutPlatformUserId,
    setShortcutPlatformUserId: state.setShortcutPlatformUserId,
    shortcutPlatformRoleName: state.shortcutPlatformRoleName,
    setShortcutPlatformRoleName: state.setShortcutPlatformRoleName,
    platformRoleOptions,
    handleAssignPlatformRole: actions.handleAssignPlatformRole,
    handleRemovePlatformRoleShortcut: actions.handleRemovePlatformRoleShortcut,
    shortcutTenantId: state.shortcutTenantId,
    setShortcutTenantId: state.setShortcutTenantId,
    shortcutTenantUserId: state.shortcutTenantUserId,
    setShortcutTenantUserId: state.setShortcutTenantUserId,
    shortcutTenantRoleName: state.shortcutTenantRoleName,
    setShortcutTenantRoleName: state.setShortcutTenantRoleName,
    tenantRoleOptions,
    handleAssignTenantRole: actions.handleAssignTenantRole,
    handleRemoveTenantMembershipShortcut:
      actions.handleRemoveTenantMembershipShortcut,
  }
}

export type AdminIamViewModel = ReturnType<typeof createAdminIamViewModel>

export default function AdminIamPage() {
  const viewModel = createAdminIamViewModel()
  return <AdminIamView model={viewModel} />
}
