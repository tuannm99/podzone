import { createSignal, type Accessor } from 'solid-js'
import {
    listGroupInlinePolicies,
    listGroupMembers,
    listGroupPolicies,
    listGroups,
    type GroupInfo,
    type GroupInlinePolicy,
    type PolicyInfo,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'
import { prettyJSON } from '../presentation'

export function createGroupsState(enabled: Accessor<boolean>, selectedOrgId: Accessor<string>) {
    const [selectedGroupId, setSelectedGroupId] = createSignal('')
    const [groupScope, setGroupScope] = createSignal('organization')
    const groupOrgId = () => (groupScope() === 'organization' ? selectedOrgId().trim() : '')
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
                groupOrgId() || undefined,
                groupScope() === 'tenant' ? groupTenantId().trim() : undefined,
                query
            )
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: () => enabled() && (groupScope() !== 'organization' || groupOrgId().length > 0),
            dependency: () => `${groupScope()}:${groupOrgId()}:${groupTenantId().trim()}`,
        }
    )
    const selectedGroupID = () => Number.parseInt(selectedGroupId().trim(), 10)
    const childEnabled = () => enabled() && Number.isFinite(selectedGroupID()) && selectedGroupID() > 0
    const members = createPaginatedResource<number>(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listGroupMembers(selectedGroupID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled: childEnabled }
    )
    const policies = createPaginatedResource<PolicyInfo>(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listGroupPolicies(selectedGroupID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled: childEnabled }
    )
    const inlinePolicies = createPaginatedResource<GroupInlinePolicy>(
        {
            page: 1,
            pageSize: 10,
            sortBy: 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listGroupInlinePolicies(selectedGroupID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        { enabled: childEnabled }
    )
    const [groupMemberUserId, setGroupMemberUserId] = createSignal('')
    const [groupPolicyName, setGroupPolicyName] = createSignal('')
    const [groupInlinePolicyName, setGroupInlinePolicyName] = createSignal('')
    const [groupInlinePolicyDescription, setGroupInlinePolicyDescription] = createSignal('')
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
        groupMembers: members.items,
        groupMembersQuery: members.query,
        groupMembersPageInfo: members.pageInfo,
        groupMembersLoading: members.loading,
        groupMembersError: members.error,
        updateGroupMembersQuery: members.updateQuery,
        reloadGroupMembers: members.reload,
        clearGroupMembers: members.clear,
        groupPolicies: policies.items,
        groupPoliciesQuery: policies.query,
        groupPoliciesPageInfo: policies.pageInfo,
        groupPoliciesLoading: policies.loading,
        groupPoliciesError: policies.error,
        updateGroupPoliciesQuery: policies.updateQuery,
        reloadGroupPolicies: policies.reload,
        clearGroupPolicies: policies.clear,
        groupInlinePolicies: inlinePolicies.items,
        groupInlinePoliciesQuery: inlinePolicies.query,
        groupInlinePoliciesPageInfo: inlinePolicies.pageInfo,
        groupInlinePoliciesLoading: inlinePolicies.loading,
        groupInlinePoliciesError: inlinePolicies.error,
        updateGroupInlinePoliciesQuery: inlinePolicies.updateQuery,
        reloadGroupInlinePolicies: inlinePolicies.reload,
        clearGroupInlinePolicies: inlinePolicies.clear,
        groupScope,
        setGroupScope,
        groupOrgId,
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
