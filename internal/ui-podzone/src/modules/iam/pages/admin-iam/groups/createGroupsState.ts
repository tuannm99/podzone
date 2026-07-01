import { createSignal } from 'solid-js'
import type { GroupInfo, GroupInlinePolicy, PolicyInfo } from '@/services/iam'
import { prettyJSON } from '../presentation'

export function createGroupsState() {
  const [groups, setGroups] = createSignal<GroupInfo[]>([])
  const [selectedGroupId, setSelectedGroupId] = createSignal('')
  const [groupMembers, setGroupMembers] = createSignal<number[]>([])
  const [groupPolicies, setGroupPolicies] = createSignal<PolicyInfo[]>([])
  const [groupInlinePolicies, setGroupInlinePolicies] = createSignal<
    GroupInlinePolicy[]
  >([])
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
  const groupOptions = () =>
    groups().map((item) => ({
      name: `${item.name}${item.tenantId ? ` · ${item.tenantId}` : ''}`,
      value: String(item.id || ''),
    }))

  return {
    groups,
    setGroups,
    selectedGroupId,
    setSelectedGroupId,
    groupMembers,
    setGroupMembers,
    groupPolicies,
    setGroupPolicies,
    groupInlinePolicies,
    setGroupInlinePolicies,
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
    groupOptions,
  }
}
