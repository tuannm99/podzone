import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps, Setter } from 'solid-js'
import type { GroupInfo, GroupInlinePolicy, PolicyInfo } from '@/services/iam'
import type { CollectionQuery, PageInfo } from '@/services/collection'
import type {
  CreateGroupFormValues,
  GroupInlinePolicyFormValues,
  GroupMemberFormValues,
  GroupPolicyAttachmentFormValues,
} from './forms'

export type ScopeOption = {
  name: string
  value: string
}

export type TenantOption = {
  name: string
  value: string
}

export type GroupOption = {
  name: string
  value: string
}

export type AdminIamGroupContextValue = {
  groupScopeOptions: Accessor<ScopeOption[]>
  groupScope: Accessor<string>
  setGroupScope: Setter<string>
  groupTenantId: Accessor<string>
  setGroupTenantId: Setter<string>
  tenantOptions: Accessor<TenantOption[]>
  groupName: Accessor<string>
  setGroupName: Setter<string>
  groupDescription: Accessor<string>
  setGroupDescription: Setter<string>
  submitCreateGroup: (event: SubmitEvent) => Promise<void>
  createGroupFromForm: (values: CreateGroupFormValues) => Promise<void>
  groupOptions: Accessor<GroupOption[]>
  groups: Accessor<GroupInfo[]>
  query: CollectionQuery
  pageInfo: Accessor<PageInfo>
  loading: Accessor<boolean>
  error: Accessor<string>
  updateQuery: (patch: Partial<CollectionQuery>) => void
  selectedGroupId: Accessor<string>
  setSelectedGroupId: Setter<string>
  groupMemberUserId: Accessor<string>
  setGroupMemberUserId: Setter<string>
  groupPolicyName: Accessor<string>
  setGroupPolicyName: Setter<string>
  addGroupMemberFromForm: (values: GroupMemberFormValues) => Promise<void>
  handleAddGroupMember: () => Promise<void>
  attachGroupPolicyFromForm: (
    values: GroupPolicyAttachmentFormValues
  ) => Promise<void>
  handleAttachGroupPolicy: () => Promise<void>
  handleDeleteGroup: () => Promise<void>
  groupMembers: Accessor<number[]>
  groupMembersQuery: CollectionQuery
  groupMembersPageInfo: Accessor<PageInfo>
  groupMembersLoading: Accessor<boolean>
  groupMembersError: Accessor<string>
  updateGroupMembersQuery: (patch: Partial<CollectionQuery>) => void
  handleRemoveGroupMember: (userId: number) => Promise<void>
  groupPolicies: Accessor<PolicyInfo[]>
  groupPoliciesQuery: CollectionQuery
  groupPoliciesPageInfo: Accessor<PageInfo>
  groupPoliciesLoading: Accessor<boolean>
  groupPoliciesError: Accessor<string>
  updateGroupPoliciesQuery: (patch: Partial<CollectionQuery>) => void
  handleDetachGroupPolicy: (policyName: string) => Promise<void>
  groupInlinePolicyName: Accessor<string>
  setGroupInlinePolicyName: Setter<string>
  groupInlinePolicyDescription: Accessor<string>
  setGroupInlinePolicyDescription: Setter<string>
  groupInlinePolicyJson: Accessor<string>
  setGroupInlinePolicyJson: Setter<string>
  saveGroupInlinePolicyFromForm: (
    values: GroupInlinePolicyFormValues
  ) => Promise<void>
  handleSaveGroupInlinePolicy: () => Promise<void>
  groupInlinePolicies: Accessor<GroupInlinePolicy[]>
  groupInlinePoliciesQuery: CollectionQuery
  groupInlinePoliciesPageInfo: Accessor<PageInfo>
  groupInlinePoliciesLoading: Accessor<boolean>
  groupInlinePoliciesError: Accessor<string>
  updateGroupInlinePoliciesQuery: (patch: Partial<CollectionQuery>) => void
  handleDeleteGroupInlinePolicy: (name: string) => Promise<void>
}

const AdminIamGroupContext = createContext<AdminIamGroupContextValue>()

export function AdminIamGroupProvider(
  props: ParentProps<{ value: AdminIamGroupContextValue }>
) {
  return (
    <AdminIamGroupContext.Provider value={props.value}>
      {props.children}
    </AdminIamGroupContext.Provider>
  )
}

export function useAdminIamGroup() {
  const ctx = useContext(AdminIamGroupContext)
  if (!ctx) {
    throw new Error('AdminIamGroupContext is missing')
  }
  return ctx
}
