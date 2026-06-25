import {
  addGroupMember,
  attachGroupPolicy,
  attachServiceControlPolicy,
  attachTenantToOrganization,
  createGroup,
  createOrganization,
  createPolicy,
  createPolicyVersion,
  deleteGroup,
  deleteGroupInlinePolicy,
  deletePolicy,
  deletePolicyVersion,
  detachGroupPolicy,
  detachServiceControlPolicy,
  detachTenantFromOrganization,
  putGroupInlinePolicy,
  removeGroupMember,
  setDefaultPolicyVersion,
  type PolicyStatement,
} from '@/services/iam';
import { parseJSONArray } from './presentation';
import type { AdminIamLoaders } from './useAdminIamLoaders';
import type { AdminIamState } from './useAdminIamState';

type RunAction = (work: () => Promise<void>) => Promise<void>;

export function useAdminIamOrgPolicyGroupActions(
  state: AdminIamState,
  loaders: AdminIamLoaders,
  runAction: RunAction
) {
  const submitCreateOrganization = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createOrganization({
        name: state.orgName().trim(),
        slug: state.orgSlug().trim(),
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Created organization ${result.data.organization?.slug || state.orgName()}.`
      );
      state.setOrgName('');
      state.setOrgSlug('');
      await loaders.loadBootstrap();
    });
  };

  const handleAttachTenantToOrg = async () => {
    await runAction(async () => {
      const result = await attachTenantToOrganization(
        state.selectedOrgId().trim(),
        state.orgTenantId().trim()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Attached tenant ${state.orgTenantId().trim()} to organization.`
      );
      await loaders.loadBootstrap();
    });
  };

  const handleDetachTenantFromOrg = async (tenantId: string) => {
    await runAction(async () => {
      const result = await detachTenantFromOrganization(
        state.selectedOrgId().trim(),
        tenantId
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Detached tenant ${tenantId} from organization.`);
      await loaders.loadBootstrap();
    });
  };

  const handleAttachScp = async () => {
    await runAction(async () => {
      const result = await attachServiceControlPolicy(
        state.selectedOrgId().trim(),
        state.orgPolicyName().trim()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Attached SCP ${state.orgPolicyName().trim()}.`);
      state.setOrgPolicyName('');
      await loaders.loadSelectedOrganization();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleDetachScp = async (policyName: string) => {
    await runAction(async () => {
      const result = await detachServiceControlPolicy(
        state.selectedOrgId().trim(),
        policyName
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Detached SCP ${policyName}.`);
      await loaders.loadSelectedOrganization();
      await loaders.loadSelectedPolicy();
    });
  };

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        state.policyStatementsJson(),
        'Policy statements'
      );
      const result = await createPolicy({
        scope: state.policyScope(),
        name: state.policyName().trim(),
        description: state.policyDescription().trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Created policy ${state.policyName().trim()}.`);
      state.setPolicyName('');
      state.setPolicyDescription('');
      await loaders.loadBootstrap();
    });
  };

  const handleCreatePolicyVersion = async () => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        state.policyVersionJson(),
        'Policy version statements'
      );
      const result = await createPolicyVersion({
        name: state.selectedPolicyName().trim(),
        statements,
        setAsDefault: false,
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Created a new version for ${state.selectedPolicyName().trim()}.`
      );
      await loaders.loadSelectedPolicy();
    });
  };

  const handleDeletePolicy = async () => {
    await runAction(async () => {
      const name = state.selectedPolicyName().trim();
      const result = await deletePolicy(name);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Deleted policy ${name}.`);
      state.setSelectedPolicyName('');
      state.setPolicyDetail(undefined);
      state.setPolicyVersions([]);
      state.setPolicyAttachments([]);
      await loaders.loadBootstrap();
    });
  };

  const handleSetDefaultVersion = async (version: string) => {
    await runAction(async () => {
      const result = await setDefaultPolicyVersion(
        state.selectedPolicyName().trim(),
        version
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Set ${version} as default for ${state.selectedPolicyName().trim()}.`
      );
      await loaders.loadSelectedPolicy();
    });
  };

  const handleDeleteVersion = async (version: string) => {
    await runAction(async () => {
      const result = await deletePolicyVersion(
        state.selectedPolicyName().trim(),
        version
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Deleted policy version ${version}.`);
      await loaders.loadSelectedPolicy();
    });
  };

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault();
    await runAction(async () => {
      const result = await createGroup({
        scope: state.groupScope(),
        tenantId:
          state.groupScope() === 'tenant'
            ? state.groupTenantId().trim()
            : undefined,
        name: state.groupName().trim(),
        description: state.groupDescription().trim(),
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Created group ${state.groupName().trim()}.`);
      state.setGroupName('');
      state.setGroupDescription('');
      await loaders.loadGroupsForScope();
    });
  };

  const handleAddGroupMember = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const userId = Number.parseInt(state.groupMemberUserId().trim(), 10);
      const result = await addGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Added user ${userId} to group.`);
      state.setGroupMemberUserId('');
      await loaders.loadSelectedGroup();
    });
  };

  const handleRemoveGroupMember = async (userId: number) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await removeGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Removed user ${userId} from group.`);
      await loaders.loadSelectedGroup();
    });
  };

  const handleAttachGroupPolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await attachGroupPolicy(
        groupId,
        state.groupPolicyName().trim()
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Attached policy ${state.groupPolicyName().trim()} to group.`
      );
      state.setGroupPolicyName('');
      await loaders.loadSelectedGroup();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleDetachGroupPolicy = async (policyName: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await detachGroupPolicy(groupId, policyName);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Detached policy ${policyName} from group.`);
      await loaders.loadSelectedGroup();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleDeleteGroup = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await deleteGroup(groupId);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Deleted group ${state.selectedGroupId().trim()}.`);
      state.setSelectedGroupId('');
      state.setGroupMembers([]);
      state.setGroupPolicies([]);
      state.setGroupInlinePolicies([]);
      await loaders.loadGroupsForScope();
    });
  };

  const handleSaveGroupInlinePolicy = async () => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const statements = parseJSONArray<PolicyStatement>(
        state.groupInlinePolicyJson(),
        'Group inline policy'
      );
      const result = await putGroupInlinePolicy(
        groupId,
        state.groupInlinePolicyName().trim(),
        state.groupInlinePolicyDescription().trim(),
        statements
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Saved group inline policy ${state.groupInlinePolicyName().trim()}.`
      );
      state.setGroupInlinePolicyName('');
      state.setGroupInlinePolicyDescription('');
      await loaders.loadSelectedGroup();
    });
  };

  const handleDeleteGroupInlinePolicy = async (name: string) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await deleteGroupInlinePolicy(groupId, name);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Deleted group inline policy ${name}.`);
      await loaders.loadSelectedGroup();
    });
  };

  return {
    submitCreateOrganization,
    handleAttachTenantToOrg,
    handleDetachTenantFromOrg,
    handleAttachScp,
    handleDetachScp,
    submitCreatePolicy,
    handleCreatePolicyVersion,
    handleDeletePolicy,
    handleSetDefaultVersion,
    handleDeleteVersion,
    submitCreateGroup,
    handleAddGroupMember,
    handleRemoveGroupMember,
    handleAttachGroupPolicy,
    handleDetachGroupPolicy,
    handleDeleteGroup,
    handleSaveGroupInlinePolicy,
    handleDeleteGroupInlinePolicy,
  };
}
