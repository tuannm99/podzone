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
import type {
  CreateGroupFormValues,
  GroupInlinePolicyFormValues,
  GroupMemberFormValues,
  GroupPolicyAttachmentFormValues,
} from './group-forms';
import type {
  CreatePolicyFormValues,
  CreatePolicyVersionFormValues,
} from './policy-forms';
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

  const createPolicyFromForm = async (values: CreatePolicyFormValues) => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Policy statements'
      );
      const result = await createPolicy({
        scope: values.scope,
        name: values.name.trim(),
        description: values.description.trim(),
        statements,
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Created policy ${values.name.trim()}.`);
      state.setPolicyName('');
      state.setPolicyDescription('');
      state.setPolicyScope(values.scope);
      state.setPolicyStatementsJson(values.statementsJson);
      await loaders.loadBootstrap();
    });
  };

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault();
    await createPolicyFromForm({
      scope: state.policyScope(),
      name: state.policyName(),
      description: state.policyDescription(),
      statementsJson: state.policyStatementsJson(),
    });
  };

  const createPolicyVersionFromForm = async (
    values: CreatePolicyVersionFormValues
  ) => {
    await runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Policy version statements'
      );
      const result = await createPolicyVersion({
        name: state.selectedPolicyName().trim(),
        statements,
        setAsDefault: false,
      });
      if (!result.success) throw new Error(result.message);
      state.setPolicyVersionJson(values.statementsJson);
      state.setPageMessage(
        `Created a new version for ${state.selectedPolicyName().trim()}.`
      );
      await loaders.loadSelectedPolicy();
    });
  };

  const handleCreatePolicyVersion = async () => {
    await createPolicyVersionFromForm({
      statementsJson: state.policyVersionJson(),
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

  const createGroupFromForm = async (values: CreateGroupFormValues) => {
    await runAction(async () => {
      const result = await createGroup({
        scope: values.scope,
        tenantId:
          values.scope === 'tenant' ? values.tenantId.trim() : undefined,
        name: values.name.trim(),
        description: values.description.trim(),
      });
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Created group ${values.name.trim()}.`);
      state.setGroupScope(values.scope);
      state.setGroupTenantId(values.tenantId);
      state.setGroupName('');
      state.setGroupDescription('');
      await loaders.loadGroupsForScope();
    });
  };

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault();
    await createGroupFromForm({
      scope: state.groupScope(),
      tenantId: state.groupTenantId(),
      name: state.groupName(),
      description: state.groupDescription(),
    });
  };

  const addGroupMemberFromForm = async (values: GroupMemberFormValues) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const userId = Number.parseInt(values.userId.trim(), 10);
      const result = await addGroupMember(groupId, userId);
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Added user ${userId} to group.`);
      state.setGroupMemberUserId('');
      await loaders.loadSelectedGroup();
    });
  };

  const handleAddGroupMember = async () => {
    await addGroupMemberFromForm({
      userId: state.groupMemberUserId(),
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

  const attachGroupPolicyFromForm = async (
    values: GroupPolicyAttachmentFormValues
  ) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const result = await attachGroupPolicy(groupId, values.policyName.trim());
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(`Attached policy ${values.policyName.trim()} to group.`);
      state.setGroupPolicyName('');
      await loaders.loadSelectedGroup();
      await loaders.loadSelectedPolicy();
    });
  };

  const handleAttachGroupPolicy = async () => {
    await attachGroupPolicyFromForm({
      policyName: state.groupPolicyName(),
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

  const saveGroupInlinePolicyFromForm = async (
    values: GroupInlinePolicyFormValues
  ) => {
    await runAction(async () => {
      const groupId = Number.parseInt(state.selectedGroupId().trim(), 10);
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Group inline policy'
      );
      const result = await putGroupInlinePolicy(
        groupId,
        values.name.trim(),
        values.description.trim(),
        statements
      );
      if (!result.success) throw new Error(result.message);
      state.setPageMessage(
        `Saved group inline policy ${values.name.trim()}.`
      );
      state.setGroupInlinePolicyName('');
      state.setGroupInlinePolicyDescription('');
      state.setGroupInlinePolicyJson(values.statementsJson);
      await loaders.loadSelectedGroup();
    });
  };

  const handleSaveGroupInlinePolicy = async () => {
    await saveGroupInlinePolicyFromForm({
      name: state.groupInlinePolicyName(),
      description: state.groupInlinePolicyDescription(),
      statementsJson: state.groupInlinePolicyJson(),
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
    createPolicyFromForm,
    createPolicyVersionFromForm,
    handleCreatePolicyVersion,
    handleDeletePolicy,
    handleSetDefaultVersion,
    handleDeleteVersion,
    submitCreateGroup,
    createGroupFromForm,
    addGroupMemberFromForm,
    handleAddGroupMember,
    handleRemoveGroupMember,
    attachGroupPolicyFromForm,
    handleAttachGroupPolicy,
    handleDetachGroupPolicy,
    handleDeleteGroup,
    saveGroupInlinePolicyFromForm,
    handleSaveGroupInlinePolicy,
    handleDeleteGroupInlinePolicy,
  };
}
