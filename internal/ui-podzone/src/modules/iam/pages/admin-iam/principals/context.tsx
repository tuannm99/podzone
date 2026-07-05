import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps, Setter } from 'solid-js'
import type {
  PermissionBoundary,
  PolicyInfo,
  UserInlinePolicy,
} from '@/services/iam'
import type { CollectionQuery, PageInfo } from '@/services/collection'
import type { SelectOption } from '@/solid/components/common/Primitives'
import type { SearchSelectOption } from '@/solid/components/common/SearchSelectField'
import type {
  PrincipalBoundaryFormValues,
  PrincipalInlinePolicyFormValues,
  PrincipalManagedPolicyFormValues,
} from './forms'

export type TenantOption = {
  name: string
  value: string
}

export type PrincipalMode = 'platform' | 'tenant'

export type AdminIamPrincipalContextValue = {
  permissionOptions: Accessor<SelectOption[]>
  managedPolicyOptions: Accessor<SelectOption[]>
  platformUserOptions: Accessor<SearchSelectOption[]>
  platformUsersLoading: Accessor<boolean>
  platformUsersError: Accessor<string>
  searchPlatformUsers: (search: string) => void
  tenantUserOptions: Accessor<SearchSelectOption[]>
  tenantUsersLoading: Accessor<boolean>
  tenantUsersError: Accessor<string>
  searchTenantUsers: (search: string) => void
  principalMode: Accessor<PrincipalMode>
  setPrincipalMode: Setter<PrincipalMode>
  principalPlatformUserId: Accessor<string>
  setPrincipalPlatformUserId: Setter<string>
  principalTenantId: Accessor<string>
  setPrincipalTenantId: Setter<string>
  principalTenantUserId: Accessor<string>
  setPrincipalTenantUserId: Setter<string>
  loadPrincipalControls: () => Promise<void>
  tenantOptions: Accessor<TenantOption[]>
  principalManagedPolicyName: Accessor<string>
  setPrincipalManagedPolicyName: Setter<string>
  attachPrincipalManagedPolicyFromForm: (
    values: PrincipalManagedPolicyFormValues
  ) => Promise<void>
  handleAttachPrincipalManagedPolicy: () => Promise<void>
  currentManagedPolicies: Accessor<PolicyInfo[]>
  managedPoliciesQuery: CollectionQuery
  managedPoliciesPageInfo: Accessor<PageInfo>
  managedPoliciesLoading: Accessor<boolean>
  managedPoliciesError: Accessor<string>
  updateManagedPoliciesQuery: (patch: Partial<CollectionQuery>) => void
  handleDetachPrincipalManagedPolicy: (policyName: string) => Promise<void>
  principalBoundaryPolicyName: Accessor<string>
  setPrincipalBoundaryPolicyName: Setter<string>
  savePrincipalBoundaryFromForm: (
    values: PrincipalBoundaryFormValues
  ) => Promise<void>
  handleSavePrincipalBoundary: () => Promise<void>
  handleDeletePrincipalBoundary: () => Promise<void>
  currentBoundary: Accessor<PermissionBoundary | null>
  principalInlinePolicyName: Accessor<string>
  setPrincipalInlinePolicyName: Setter<string>
  principalInlinePolicyDescription: Accessor<string>
  setPrincipalInlinePolicyDescription: Setter<string>
  principalInlinePolicyJson: Accessor<string>
  setPrincipalInlinePolicyJson: Setter<string>
  savePrincipalInlinePolicyFromForm: (
    values: PrincipalInlinePolicyFormValues
  ) => Promise<void>
  handleSavePrincipalInlinePolicy: () => Promise<void>
  currentInlinePolicies: Accessor<UserInlinePolicy[]>
  inlinePoliciesQuery: CollectionQuery
  inlinePoliciesPageInfo: Accessor<PageInfo>
  inlinePoliciesLoading: Accessor<boolean>
  inlinePoliciesError: Accessor<string>
  updateInlinePoliciesQuery: (patch: Partial<CollectionQuery>) => void
  handleDeletePrincipalInlinePolicy: (name: string) => Promise<void>
}

const AdminIamPrincipalContext = createContext<AdminIamPrincipalContextValue>()

export function AdminIamPrincipalProvider(
  props: ParentProps<{ value: AdminIamPrincipalContextValue }>
) {
  return (
    <AdminIamPrincipalContext.Provider value={props.value}>
      {props.children}
    </AdminIamPrincipalContext.Provider>
  )
}

export function useAdminIamPrincipal() {
  const ctx = useContext(AdminIamPrincipalContext)
  if (!ctx) {
    throw new Error('AdminIamPrincipalContext is missing')
  }
  return ctx
}
