import {
  addGroupMember,
  attachGroupPolicy,
  createGroup,
  deleteGroup,
  deleteGroupInlinePolicy,
  detachGroupPolicy,
  putGroupInlinePolicy,
  removeGroupMember,
  type PolicyStatement,
} from '@/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import { parseJSONArray } from '../presentation'
import type { RunAction } from '../shared/actions'
import type {
  CreateGroupFormValues,
  GroupInlinePolicyFormValues,
  GroupMemberFormValues,
  GroupPolicyAttachmentFormValues,
} from './forms'

export function createGroupsActions(
  state: AdminIamState,
  loaders: AdminIamLoaders,
  runAction: RunAction
) {
  const groupID = () => Number.parseInt(state.selectedGroupId().trim(), 10)
  const createGroupFromForm = (values: CreateGroupFormValues) =>
    runAction(async () => {
      const result = await createGroup({
        scope: values.scope,
        orgId:
          values.scope === 'organization'
            ? state.selectedOrgId().trim()
            : undefined,
        tenantId:
          values.scope === 'tenant' ? values.tenantId.trim() : undefined,
        name: values.name.trim(),
        description: values.description.trim(),
      })
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Created group ${values.name.trim()}.`)
      state.setGroupScope(values.scope)
      state.setGroupTenantId(values.tenantId)
      state.setGroupName('')
      state.setGroupDescription('')
      await loaders.loadGroupsForScope()
    })

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault()
    await createGroupFromForm({
      scope: state.groupScope(),
      tenantId: state.groupTenantId(),
      name: state.groupName(),
      description: state.groupDescription(),
    })
  }

  const addGroupMemberFromForm = (values: GroupMemberFormValues) =>
    runAction(async () => {
      const userID = Number.parseInt(values.userId.trim(), 10)
      const result = await addGroupMember(groupID(), userID)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Added user ${userID} to group.`)
      state.setGroupMemberUserId('')
      await loaders.loadSelectedGroup()
    })

  const handleAddGroupMember = () =>
    addGroupMemberFromForm({ userId: state.groupMemberUserId() })

  const handleRemoveGroupMember = (userID: number) =>
    runAction(async () => {
      const result = await removeGroupMember(groupID(), userID)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Removed user ${userID} from group.`)
      await loaders.loadSelectedGroup()
    })

  const attachGroupPolicyFromForm = (values: GroupPolicyAttachmentFormValues) =>
    runAction(async () => {
      const policyName = values.policyName.trim()
      const result = await attachGroupPolicy(groupID(), policyName)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Attached policy ${policyName} to group.`)
      state.setGroupPolicyName('')
      await loaders.loadSelectedGroup()
      await loaders.loadSelectedPolicy()
    })

  const handleAttachGroupPolicy = () =>
    attachGroupPolicyFromForm({ policyName: state.groupPolicyName() })

  const handleDetachGroupPolicy = (policyName: string) =>
    runAction(async () => {
      const result = await detachGroupPolicy(groupID(), policyName)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Detached policy ${policyName} from group.`)
      await loaders.loadSelectedGroup()
      await loaders.loadSelectedPolicy()
    })

  const handleDeleteGroup = () =>
    runAction(async () => {
      const result = await deleteGroup(groupID())
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Deleted group ${state.selectedGroupId().trim()}.`)
      state.setSelectedGroupId('')
      state.clearGroupMembers()
      state.clearGroupPolicies()
      state.clearGroupInlinePolicies()
      await loaders.loadGroupsForScope()
    })

  const saveGroupInlinePolicyFromForm = (values: GroupInlinePolicyFormValues) =>
    runAction(async () => {
      const statements = parseJSONArray<PolicyStatement>(
        values.statementsJson,
        'Group inline policy'
      )
      const result = await putGroupInlinePolicy(
        groupID(),
        values.name.trim(),
        values.description.trim(),
        statements
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Saved group inline policy ${values.name.trim()}.`)
      state.setGroupInlinePolicyName('')
      state.setGroupInlinePolicyDescription('')
      state.setGroupInlinePolicyJson(values.statementsJson)
      await loaders.loadSelectedGroup()
    })

  const handleSaveGroupInlinePolicy = () =>
    saveGroupInlinePolicyFromForm({
      name: state.groupInlinePolicyName(),
      description: state.groupInlinePolicyDescription(),
      statementsJson: state.groupInlinePolicyJson(),
    })

  const handleDeleteGroupInlinePolicy = (name: string) =>
    runAction(async () => {
      const result = await deleteGroupInlinePolicy(groupID(), name)
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Deleted group inline policy ${name}.`)
      await loaders.loadSelectedGroup()
    })

  return {
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
  }
}
