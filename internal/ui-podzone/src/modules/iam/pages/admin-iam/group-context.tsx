import { createContext, useContext } from 'solid-js';
import type { Accessor, ParentProps, Setter } from 'solid-js';
import type { GroupInlinePolicy, PolicyInfo } from '@/services/iam';

export type ScopeOption = {
  name: string;
  value: string;
};

export type TenantOption = {
  name: string;
  value: string;
};

export type GroupOption = {
  name: string;
  value: string;
};

export type AdminIamGroupContextValue = {
  groupScopeOptions: ScopeOption[];
  groupScope: Accessor<string>;
  setGroupScope: Setter<string>;
  groupTenantId: Accessor<string>;
  setGroupTenantId: Setter<string>;
  tenantOptions: Accessor<TenantOption[]>;
  groupName: Accessor<string>;
  setGroupName: Setter<string>;
  groupDescription: Accessor<string>;
  setGroupDescription: Setter<string>;
  submitCreateGroup: (event: SubmitEvent) => Promise<void>;
  groupOptions: Accessor<GroupOption[]>;
  selectedGroupId: Accessor<string>;
  setSelectedGroupId: Setter<string>;
  groupMemberUserId: Accessor<string>;
  setGroupMemberUserId: Setter<string>;
  groupPolicyName: Accessor<string>;
  setGroupPolicyName: Setter<string>;
  handleAddGroupMember: () => Promise<void>;
  handleAttachGroupPolicy: () => Promise<void>;
  handleDeleteGroup: () => Promise<void>;
  groupMembers: Accessor<number[]>;
  handleRemoveGroupMember: (userId: number) => Promise<void>;
  groupPolicies: Accessor<PolicyInfo[]>;
  handleDetachGroupPolicy: (policyName: string) => Promise<void>;
  groupInlinePolicyName: Accessor<string>;
  setGroupInlinePolicyName: Setter<string>;
  groupInlinePolicyDescription: Accessor<string>;
  setGroupInlinePolicyDescription: Setter<string>;
  groupInlinePolicyJson: Accessor<string>;
  setGroupInlinePolicyJson: Setter<string>;
  handleSaveGroupInlinePolicy: () => Promise<void>;
  groupInlinePolicies: Accessor<GroupInlinePolicy[]>;
  handleDeleteGroupInlinePolicy: (name: string) => Promise<void>;
};

const AdminIamGroupContext = createContext<AdminIamGroupContextValue>();

export function AdminIamGroupProvider(
  props: ParentProps<{ value: AdminIamGroupContextValue }>
) {
  return (
    <AdminIamGroupContext.Provider value={props.value}>
      {props.children}
    </AdminIamGroupContext.Provider>
  );
}

export function useAdminIamGroup() {
  const ctx = useContext(AdminIamGroupContext);
  if (!ctx) {
    throw new Error('AdminIamGroupContext is missing');
  }
  return ctx;
}
