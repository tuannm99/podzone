import { createSignal, type Accessor } from 'solid-js'
import {
  listGroups,
  type GroupInfo,
  type GroupInlinePolicy,
  type PolicyInfo,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'
import { prettyJSON } from '../presentation'

export function createGroupsState(enabled: Accessor<boolean>) {
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
  const collection = createPaginatedResource<GroupInfo>(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listGroups(
        groupScope(),
        groupScope() === 'tenant' ? groupTenantId().trim() : undefined,
        query
      )
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    { enabled }
  )
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
    collection.items().map((item) => ({
      name: `${item.name}${item.tenantId ? ` · ${item.tenantId}` : ''}`,
      value: String(item.id || ''),
    }))

  return {
    groups: collection.items,
    groupsQuery: collection.query,
    groupsPageInfo: collection.pageInfo,
    groupsLoading: collection.loading,
    groupsError: collection.error,
    updateGroupsQuery: collection.updateQuery,
    reloadGroups: collection.reload,
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
