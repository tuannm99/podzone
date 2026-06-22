import { createEffect, onMount } from 'solid-js';
import {
  addGroupMember,
  attachGroupPolicy,
  attachPlatformUserPolicy,
  attachServiceControlPolicy,
  attachTenantToOrganization,
  attachTenantUserPolicy,
  createGroup,
  createOrganization,
  createPolicy,
  createPolicyVersion,
  deleteGroup,
  deleteGroupInlinePolicy,
  deletePolicy,
  deletePolicyVersion,
  deletePlatformUserInlinePolicy,
  deletePlatformUserPermissionBoundary,
  deleteTenantUserInlinePolicy,
  deleteTenantUserPermissionBoundary,
  detachServiceControlPolicy,
  detachGroupPolicy,
  detachTenantFromOrganization,
  detachPlatformUserPolicy,
  detachTenantUserPolicy,
  putGroupInlinePolicy,
  putPlatformUserInlinePolicy,
  putPlatformUserPermissionBoundary,
  putRolePermissionBoundary,
  putRoleTrustPolicy,
  putTenantUserInlinePolicy,
  putTenantUserPermissionBoundary,
  removeGroupMember,
  removePlatformRole,
  removeTenantMember,
  setDefaultPolicyVersion,
  simulateAccess,
  type PolicyStatement,
  type RoleTrustStatement,
  deleteRolePermissionBoundary,
  upsertPlatformRole,
  upsertTenantMember,
} from '@/services/iam';
import { tokenStorage } from '@/services/tokenStorage';
import { AdminIamView } from './admin-iam/AdminIamView';
import {
  attachmentColor,
  groupScopeOptions,
  parseJSONArray,
  parseJSONObject,
  platformRoleOptions,
  policyScopeOptions,
  prettyJSON,
  sectionLinks,
  simulationLayerTone,
  simulationSourceColor,
  statementSourceLabel,
  tenantRoleOptions,
} from './admin-iam/presentation';
import { useAdminIamLoaders } from './admin-iam/useAdminIamLoaders';
import { useAdminIamState } from './admin-iam/useAdminIamState';

export default function AdminIamPage() {
  const userID = tokenStorage.getUserID() || 0;
  const iamState = useAdminIamState(userID);
  const {
    pageError, setPageError, pageMessage, setPageMessage, loading,
    allowed, organizations, selectedOrgId,
    setSelectedOrgId, selectedPolicyName, setSelectedPolicyName,
    selectedGroupId, setSelectedGroupId, policyDetail, setPolicyDetail,
    policyVersions, setPolicyVersions, policyAttachments, setPolicyAttachments,
    orgPolicies, groupMembers, setGroupMembers, groupPolicies,
    setGroupPolicies, groupInlinePolicies, setGroupInlinePolicies, roleBoundary,
    setRoleBoundary, platformUserPolicies,
    tenantUserPolicies, platformUserInlinePolicies,
    tenantUserInlinePolicies, platformUserBoundary,
    tenantUserBoundary, simulation, setSimulation,
    orgName, setOrgName, orgSlug, setOrgSlug, orgTenantId, setOrgTenantId,
    orgPolicyName, setOrgPolicyName, policyScope, setPolicyScope, policyName,
    setPolicyName, policyDescription, setPolicyDescription,
    policyStatementsJson, setPolicyStatementsJson, policyVersionJson,
    setPolicyVersionJson, groupScope, setGroupScope, groupTenantId,
    setGroupTenantId, groupName, setGroupName, groupDescription,
    setGroupDescription, groupMemberUserId, setGroupMemberUserId,
    groupPolicyName, setGroupPolicyName, groupInlinePolicyName,
    setGroupInlinePolicyName, groupInlinePolicyDescription,
    setGroupInlinePolicyDescription, groupInlinePolicyJson,
    setGroupInlinePolicyJson, shortcutPlatformUserId,
    setShortcutPlatformUserId, shortcutPlatformRoleName,
    setShortcutPlatformRoleName, shortcutTenantId, setShortcutTenantId,
    shortcutTenantUserId, setShortcutTenantUserId, shortcutTenantRoleName,
    setShortcutTenantRoleName, trustRoleName, setTrustRoleName,
    trustBoundaryPolicyName, setTrustBoundaryPolicyName, trustJson,
    setTrustJson, simScope, setSimScope, simTenantId, setSimTenantId,
    simTargetUserId, setSimTargetUserId, simAction, setSimAction, simResource,
    setSimResource, simServicePrincipal, setSimServicePrincipal,
    simAttributesJson, setSimAttributesJson, simSessionTagsJson,
    setSimSessionTagsJson, simSessionPolicyJson, setSimSessionPolicyJson,
    simAssumedRoleId, setSimAssumedRoleId, simAssumedRoleScope,
    setSimAssumedRoleScope, simAssumedRoleName, setSimAssumedRoleName,
    simAssumedRoleTenantId, setSimAssumedRoleTenantId,
    simAssumedRoleSessionName, setSimAssumedRoleSessionName,
    simAssumedRoleSourceIdentity, setSimAssumedRoleSourceIdentity,
    simAssumedRoleServicePrincipal, setSimAssumedRoleServicePrincipal,
    simAssumedRoleExpiresAt, setSimAssumedRoleExpiresAt, principalMode,
    setPrincipalMode, principalPlatformUserId, setPrincipalPlatformUserId,
    principalTenantId, setPrincipalTenantId, principalTenantUserId,
    setPrincipalTenantUserId, principalManagedPolicyName,
    setPrincipalManagedPolicyName, principalBoundaryPolicyName,
    setPrincipalBoundaryPolicyName, principalInlinePolicyName,
    setPrincipalInlinePolicyName, principalInlinePolicyDescription,
    setPrincipalInlinePolicyDescription, principalInlinePolicyJson,
    setPrincipalInlinePolicyJson, tenantOptions, policyOptions,
    organizationOptions, groupOptions,
  } = iamState;
  const {
    loadBootstrap,
    loadGroupsForScope,
    loadSelectedPolicy,
    loadSelectedOrganization,
    loadSelectedGroup,
    loadTrustPolicy,
    loadPrincipalControls,
  } = useAdminIamLoaders(iamState, userID);

  const buildAssumedRoleSession = () => {
    const roleId = Number.parseInt(simAssumedRoleId().trim(), 10);
    if (!Number.isFinite(roleId) || roleId <= 0) return undefined;
    return {
      assumedRoleId: roleId,
      assumedRoleScope: simAssumedRoleScope().trim(),
      assumedRoleName: simAssumedRoleName().trim(),
      assumedRoleTenantId: simAssumedRoleTenantId().trim() || undefined,
      assumedRoleSessionName: simAssumedRoleSessionName().trim() || undefined,
      assumedRoleSourceIdentity:
        simAssumedRoleSourceIdentity().trim() || undefined,
      assumedRoleServicePrincipal:
        simAssumedRoleServicePrincipal().trim() || undefined,
      assumedRoleExpiresAt: simAssumedRoleExpiresAt().trim() || undefined,
      sessionTags: parseJSONObject(simSessionTagsJson(), 'Session tags'),
    };
  };

  createEffect(() => {
    if (
      simAssumedRoleScope() === 'tenant' &&
      simTenantId().trim() &&
      !simAssumedRoleTenantId().trim()
    ) {
      setSimAssumedRoleTenantId(simTenantId().trim());
    }
  });

  const applyServiceAssumePreset = () => {
    setSimScope('platform');
    setSimAction('order:update');
    setSimResource('*');
    setSimServicePrincipal('backoffice.podzone.internal');
    setSimAssumedRoleScope('platform');
    setSimAssumedRoleName('platform_admin');
    setSimAssumedRoleSourceIdentity('backoffice-admin');
    setSimAssumedRoleSessionName('service-assume');
    setSimAssumedRoleServicePrincipal('backoffice.podzone.internal');
    setSimAttributesJson(prettyJSON({ lane: 'priority' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', path: 'service-assume' }));
  };

  const applyTenantAssumePreset = () => {
    setSimScope('tenant');
    setSimAction('order:update');
    setSimResource('*');
    setSimAssumedRoleScope('tenant');
    setSimAssumedRoleName('tenant_admin');
    setSimAssumedRoleTenantId(simTenantId().trim());
    setSimAssumedRoleSessionName('tenant-admin-review');
    setSimAssumedRoleSourceIdentity('store-ops');
    setSimAssumedRoleServicePrincipal('');
    setSimAttributesJson(prettyJSON({ lane: 'priority', region: 'us' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', store: simTenantId().trim() || 'tenant' }));
  };

  const applyScopeDownDenyPreset = () => {
    setSimAction('order:update');
    setSimResource('*');
    setSimSessionPolicyJson(
      prettyJSON([
        {
          effect: 'deny',
          actionPattern: 'order:update',
          resourcePattern: '*',
          conditions: [],
        },
      ])
    );
    setSimAttributesJson(prettyJSON({ lane: 'restricted' }));
    setSimSessionTagsJson(prettyJSON({ team: 'ops', mode: 'scope-down' }));
  };

  createEffect(() => {
    void selectedPolicyName();
    if (allowed()) void loadSelectedPolicy();
  });

  createEffect(() => {
    void selectedOrgId();
    if (allowed()) void loadSelectedOrganization();
  });

  createEffect(() => {
    void selectedGroupId();
    if (allowed()) void loadSelectedGroup();
  });

  createEffect(() => {
    void groupScope();
    void groupTenantId();
    if (allowed()) void loadGroupsForScope();
  });

  createEffect(() => {
    void principalMode();
    void principalPlatformUserId();
    void principalTenantId();
    void principalTenantUserId();
    if (allowed()) void loadPrincipalControls();
  });

  onMount(() => {
    void loadBootstrap();
  });

  const runAction = async (work: () => Promise<void>) => {
    setPageError('');
    setPageMessage('');
    try {
      await work();
    } catch (error) {
      setPageError(error instanceof Error ? error.message : 'Action failed');
    }
  };

  const submitCreateOrganization = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createOrganization({
        name: orgName().trim(),
        slug: orgSlug().trim(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created organization ${result.data.organization?.slug || orgName()}.`);
      setOrgName('');
      setOrgSlug('');
      await loadBootstrap();
    });
  };

  const handleAttachTenantToOrg = async () => {
    await runAction(async () => {
      const result = await attachTenantToOrganization(selectedOrgId().trim(), orgTenantId().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached tenant ${orgTenantId().trim()} to organization.`);
      await loadBootstrap();
    });
  };

  const handleDetachTenantFromOrg = async (tenantId: string) => {
    await runAction(async () => {
      const result = await detachTenantFromOrganization(selectedOrgId().trim(), tenantId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached tenant ${tenantId} from organization.`);
      await loadBootstrap();
    });
  };

  const handleAttachScp = async () => {
    await runAction(async () => {
      const result = await attachServiceControlPolicy(selectedOrgId().trim(), orgPolicyName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached SCP ${orgPolicyName().trim()}.`);
      setOrgPolicyName('');
      await loadSelectedOrganization();
      await loadSelectedPolicy();
    });
  };

  const handleDetachScp = async (policyName: string) => {
    await runAction(async () => {
      const result = await detachServiceControlPolicy(selectedOrgId().trim(), policyName);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached SCP ${policyName}.`);
      await loadSelectedOrganization();
      await loadSelectedPolicy();
    });
  };

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(policyStatementsJson(), 'Policy statements');
      const result = await createPolicy({
        scope: policyScope(),
        name: policyName().trim(),
        description: policyDescription().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created policy ${policyName().trim()}.`);
      setPolicyName('');
      setPolicyDescription('');
      await loadBootstrap();
    });
  };

  const handleCreatePolicyVersion = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(policyVersionJson(), 'Policy version statements');
      const result = await createPolicyVersion({
        name: selectedPolicyName().trim(),
        statements,
        setAsDefault: false,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created a new version for ${selectedPolicyName().trim()}.`);
      await loadSelectedPolicy();
    });
  };

  const handleDeletePolicy = async () => {
    await runAction(async () => {
      const name = selectedPolicyName().trim();
      const result = await deletePolicy(name);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted policy ${name}.`);
      setSelectedPolicyName('');
      setPolicyDetail(undefined);
      setPolicyVersions([]);
      setPolicyAttachments([]);
      await loadBootstrap();
    });
  };

  const handleSetDefaultVersion = async (version: string) => {
    await runAction(async () => {
      const result = await setDefaultPolicyVersion(selectedPolicyName().trim(), version);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Set ${version} as default for ${selectedPolicyName().trim()}.`);
      await loadSelectedPolicy();
    });
  };

  const handleDeleteVersion = async (version: string) => {
    await runAction(async () => {
      const result = await deletePolicyVersion(selectedPolicyName().trim(), version);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted policy version ${version}.`);
      await loadSelectedPolicy();
    });
  };

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createGroup({
        scope: groupScope(),
        tenantId: groupScope() === 'tenant' ? groupTenantId().trim() : undefined,
        name: groupName().trim(),
        description: groupDescription().trim(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Created group ${groupName().trim()}.`);
      setGroupName('');
      setGroupDescription('');
      await loadGroupsForScope();
    });
  };

  const handleAddGroupMember = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const userId = Number.parseInt(groupMemberUserId().trim(), 10);
      const result = await addGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Added user ${userId} to group.`);
      setGroupMemberUserId('');
      await loadSelectedGroup();
    });
  };

  const handleRemoveGroupMember = async (userId: number) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await removeGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed user ${userId} from group.`);
      await loadSelectedGroup();
    });
  };

  const handleAttachGroupPolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await attachGroupPolicy(groupId, groupPolicyName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Attached policy ${groupPolicyName().trim()} to group.`);
      setGroupPolicyName('');
      await loadSelectedGroup();
      await loadSelectedPolicy();
    });
  };

  const handleDetachGroupPolicy = async (policyName: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await detachGroupPolicy(groupId, policyName);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Detached policy ${policyName} from group.`);
      await loadSelectedGroup();
      await loadSelectedPolicy();
    });
  };

  const handleDeleteGroup = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await deleteGroup(groupId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted group ${selectedGroupId().trim()}.`);
      setSelectedGroupId('');
      setGroupMembers([]);
      setGroupPolicies([]);
      setGroupInlinePolicies([]);
      await loadGroupsForScope();
    });
  };

  const handleSaveGroupInlinePolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const statements = parseJSONArray<PolicyStatement>(groupInlinePolicyJson(), 'Group inline policy');
      const result = await putGroupInlinePolicy(
        groupId,
        groupInlinePolicyName().trim(),
        groupInlinePolicyDescription().trim(),
        statements
      );
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved group inline policy ${groupInlinePolicyName().trim()}.`);
      setGroupInlinePolicyName('');
      setGroupInlinePolicyDescription('');
      await loadSelectedGroup();
    });
  };

  const handleDeleteGroupInlinePolicy = async (name: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(selectedGroupId().trim(), 10);
      const result = await deleteGroupInlinePolicy(groupId, name);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted group inline policy ${name}.`);
      await loadSelectedGroup();
    });
  };

  const handleSaveTrustPolicy = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<RoleTrustStatement>(trustJson(), 'Trust policy');
      const result = await putRoleTrustPolicy({
        roleName: trustRoleName().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved trust policy for role ${trustRoleName().trim()}.`);
      await loadTrustPolicy();
    });
  };

  const handleSaveRoleBoundary = async () => {
    await runAction(async () => {
      const result = await putRolePermissionBoundary(
        trustRoleName().trim(),
        trustBoundaryPolicyName().trim()
      );
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Saved role boundary for ${trustRoleName().trim()}.`);
      await loadTrustPolicy();
    });
  };

  const handleDeleteRoleBoundary = async () => {
    await runAction(async () => {
      const result = await deleteRolePermissionBoundary(trustRoleName().trim());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Deleted role boundary for ${trustRoleName().trim()}.`);
      setRoleBoundary(null);
      setTrustBoundaryPolicyName('');
    });
  };

  const handleSimulate = async () => {
    await runAction(async () => {
      const result = await simulateAccess({
        scope: simScope(),
        tenantId: simTenantId().trim() || undefined,
        userId: Number.parseInt(simTargetUserId().trim(), 10),
        action: simAction().trim(),
        resource: simResource().trim(),
        useAssumedRole: Boolean(buildAssumedRoleSession()),
        assumedRoleSession: buildAssumedRoleSession(),
        sessionPolicy: parseJSONArray<PolicyStatement>(simSessionPolicyJson(), 'Session policy'),
        attributes: parseJSONObject(simAttributesJson(), 'Simulation attributes'),
        servicePrincipal: simServicePrincipal().trim() || undefined,
        sessionTags: parseJSONObject(simSessionTagsJson(), 'Session tags'),
      });
      if (!result.success) throw new Error(result.message);
      setSimulation(result.data);
      setPageMessage(
        `Simulation completed: ${result.data.allowed ? 'allowed' : 'denied'} via ${result.data.decisionSource}.`
      );
    });
  };

  const handleAttachPrincipalManagedPolicy = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await attachPlatformUserPolicy(targetUserId, principalManagedPolicyName().trim());
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await attachTenantUserPolicy(
          principalTenantId().trim(),
          targetUserId,
          principalManagedPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Attached managed policy ${principalManagedPolicyName().trim()}.`);
      setPrincipalManagedPolicyName('');
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleDetachPrincipalManagedPolicy = async (policyName: string) => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await detachPlatformUserPolicy(targetUserId, policyName);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await detachTenantUserPolicy(principalTenantId().trim(), targetUserId, policyName);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Detached managed policy ${policyName}.`);
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleSavePrincipalInlinePolicy = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(principalInlinePolicyJson(), 'Inline policy');
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await putPlatformUserInlinePolicy(
          targetUserId,
          principalInlinePolicyName().trim(),
          principalInlinePolicyDescription().trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await putTenantUserInlinePolicy(
          principalTenantId().trim(),
          targetUserId,
          principalInlinePolicyName().trim(),
          principalInlinePolicyDescription().trim(),
          statements
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Saved inline policy ${principalInlinePolicyName().trim()}.`);
      setPrincipalInlinePolicyName('');
      setPrincipalInlinePolicyDescription('');
      await loadPrincipalControls();
    });
  };

  const handleDeletePrincipalInlinePolicy = async (name: string) => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await deletePlatformUserInlinePolicy(targetUserId, name);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await deleteTenantUserInlinePolicy(principalTenantId().trim(), targetUserId, name);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Deleted inline policy ${name}.`);
      await loadPrincipalControls();
    });
  };

  const handleSavePrincipalBoundary = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await putPlatformUserPermissionBoundary(
          targetUserId,
          principalBoundaryPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await putTenantUserPermissionBoundary(
          principalTenantId().trim(),
          targetUserId,
          principalBoundaryPolicyName().trim()
        );
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage(`Saved principal boundary ${principalBoundaryPolicyName().trim()}.`);
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleDeletePrincipalBoundary = async () => {
    await runAction(async () => {
      if (principalMode() === 'platform') {
        const targetUserId = Number.parseInt(principalPlatformUserId().trim(), 10);
        const result = await deletePlatformUserPermissionBoundary(targetUserId);
        if (!result.success) throw new Error(result.message);
      } else {
        const targetUserId = Number.parseInt(principalTenantUserId().trim(), 10);
        const result = await deleteTenantUserPermissionBoundary(principalTenantId().trim(), targetUserId);
        if (!result.success) throw new Error(result.message);
      }
      setPageMessage('Deleted principal permission boundary.');
      setPrincipalBoundaryPolicyName('');
      await loadPrincipalControls();
      await loadSelectedPolicy();
    });
  };

  const handleAssignPlatformRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutPlatformUserId().trim(), 10);
      const result = await upsertPlatformRole({
        targetUserId,
        roleName: shortcutPlatformRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Assigned platform role ${shortcutPlatformRoleName()} to user ${targetUserId}.`);
    });
  };

  const handleRemovePlatformRoleShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutPlatformUserId().trim(), 10);
      const result = await removePlatformRole(targetUserId, shortcutPlatformRoleName());
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed platform role ${shortcutPlatformRoleName()} from user ${targetUserId}.`);
    });
  };

  const handleAssignTenantRole = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutTenantUserId().trim(), 10);
      const result = await upsertTenantMember({
        tenantId: shortcutTenantId().trim(),
        userId: targetUserId,
        roleName: shortcutTenantRoleName(),
      });
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Assigned tenant role ${shortcutTenantRoleName()} to user ${targetUserId}.`);
    });
  };

  const handleRemoveTenantMembershipShortcut = async () => {
    await runAction(async () => {
      const targetUserId = Number.parseInt(shortcutTenantUserId().trim(), 10);
      const result = await removeTenantMember(shortcutTenantId().trim(), targetUserId);
      if (!result.success) throw new Error(result.message);
      setPageMessage(`Removed tenant membership for user ${targetUserId}.`);
    });
  };

  const policyContextValue = {
    policyScopeOptions,
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
    selectedPolicyName,
    setSelectedPolicyName,
    policyOptions,
    policyDetail,
    policyVersions,
    policyAttachments,
    attachmentColor,
    submitCreatePolicy,
    handleCreatePolicyVersion,
    handleDeletePolicy,
    handleSetDefaultVersion,
    handleDeleteVersion,
  };

  const groupContextValue = {
    groupScopeOptions,
    groupScope,
    setGroupScope,
    groupTenantId,
    setGroupTenantId,
    tenantOptions,
    groupName,
    setGroupName,
    groupDescription,
    setGroupDescription,
    submitCreateGroup,
    groupOptions,
    selectedGroupId,
    setSelectedGroupId,
    groupMemberUserId,
    setGroupMemberUserId,
    groupPolicyName,
    setGroupPolicyName,
    handleAddGroupMember,
    handleAttachGroupPolicy,
    handleDeleteGroup,
    groupMembers,
    handleRemoveGroupMember,
    groupPolicies,
    handleDetachGroupPolicy,
    groupInlinePolicyName,
    setGroupInlinePolicyName,
    groupInlinePolicyDescription,
    setGroupInlinePolicyDescription,
    groupInlinePolicyJson,
    setGroupInlinePolicyJson,
    handleSaveGroupInlinePolicy,
    groupInlinePolicies,
    handleDeleteGroupInlinePolicy,
  };

  const principalContextValue = {
    principalMode,
    setPrincipalMode,
    principalPlatformUserId,
    setPrincipalPlatformUserId,
    principalTenantId,
    setPrincipalTenantId,
    principalTenantUserId,
    setPrincipalTenantUserId,
    tenantOptions,
    principalManagedPolicyName,
    setPrincipalManagedPolicyName,
    handleAttachPrincipalManagedPolicy,
    currentManagedPolicies: () =>
      principalMode() === 'platform'
        ? platformUserPolicies()
        : tenantUserPolicies(),
    handleDetachPrincipalManagedPolicy,
    principalBoundaryPolicyName,
    setPrincipalBoundaryPolicyName,
    handleSavePrincipalBoundary,
    handleDeletePrincipalBoundary,
    currentBoundary: () =>
      principalMode() === 'platform'
        ? platformUserBoundary()
        : tenantUserBoundary(),
    principalInlinePolicyName,
    setPrincipalInlinePolicyName,
    principalInlinePolicyDescription,
    setPrincipalInlinePolicyDescription,
    principalInlinePolicyJson,
    setPrincipalInlinePolicyJson,
    handleSavePrincipalInlinePolicy,
    currentInlinePolicies: () =>
      principalMode() === 'platform'
        ? platformUserInlinePolicies()
        : tenantUserInlinePolicies(),
    handleDeletePrincipalInlinePolicy,
  };

  const trustSimContextValue = {
    trustRoleName,
    setTrustRoleName,
    loadTrustPolicy,
    handleSaveTrustPolicy,
    trustBoundaryPolicyName,
    setTrustBoundaryPolicyName,
    handleSaveRoleBoundary,
    handleDeleteRoleBoundary,
    roleBoundary,
    trustJson,
    setTrustJson,
    simScope,
    setSimScope,
    policyScopeOptions,
    simTargetUserId,
    setSimTargetUserId,
    tenantOptions,
    simTenantId,
    setSimTenantId,
    simAction,
    setSimAction,
    simResource,
    setSimResource,
    simServicePrincipal,
    setSimServicePrincipal,
    applyServiceAssumePreset,
    applyTenantAssumePreset,
    applyScopeDownDenyPreset,
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
    handleSimulate,
    simulation,
    simulationSourceColor,
    statementSourceLabel,
    simulationLayerTone,
  };

  return (
    <AdminIamView
      model={{
        pageError,
        pageMessage,
        loading,
        allowed,
        sectionLinks,
        policyContextValue,
        groupContextValue,
        principalContextValue,
        trustSimContextValue,
        organizationOptions,
        selectedOrgId,
        setSelectedOrgId,
        submitCreateOrganization,
        orgName,
        setOrgName,
        orgSlug,
        setOrgSlug,
        orgTenantId,
        setOrgTenantId,
        orgPolicyName,
        setOrgPolicyName,
        tenantOptions,
        handleAttachTenantToOrg,
        handleAttachScp,
        organizations,
        orgPolicies,
        handleDetachTenantFromOrg,
        handleDetachScp,
        shortcutPlatformUserId,
        setShortcutPlatformUserId,
        shortcutPlatformRoleName,
        setShortcutPlatformRoleName,
        platformRoleOptions,
        handleAssignPlatformRole,
        handleRemovePlatformRoleShortcut,
        shortcutTenantId,
        setShortcutTenantId,
        shortcutTenantUserId,
        setShortcutTenantUserId,
        shortcutTenantRoleName,
        setShortcutTenantRoleName,
        tenantRoleOptions,
        handleAssignTenantRole,
        handleRemoveTenantMembershipShortcut,
      }}
    />
  );
}
