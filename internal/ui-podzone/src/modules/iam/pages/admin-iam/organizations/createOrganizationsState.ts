import { createSignal, type Accessor, type Setter } from 'solid-js'
import {
  listOrganizations,
  listOrganizationMembers,
  type OrganizationInfo,
  type OrganizationMembership,
  type PolicyInfo,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'

export function createOrganizationsState(
  enabled: Accessor<boolean>,
  setCanManagePlatform: Setter<boolean>
) {
  const [selectedOrgId, setSelectedOrgId] = createSignal('')
  const collection = createPaginatedResource<OrganizationInfo>(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listOrganizations(query)
      if (!result.success) throw new Error(result.message)
      setCanManagePlatform(result.data.canManagePlatform)
      return result.data
    },
    { enabled }
  )
  const members = createPaginatedResource<OrganizationMembership>(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_ASC',
    },
    async (query) => {
      const result = await listOrganizationMembers(
        selectedOrgId().trim(),
        query
      )
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () => enabled() && selectedOrgId().trim().length > 0,
      dependency: selectedOrgId,
    }
  )
  const [orgPolicies, setOrgPolicies] = createSignal<PolicyInfo[]>([])
  const [orgName, setOrgName] = createSignal('')
  const [orgSlug, setOrgSlug] = createSignal('')
  const [orgTenantId, setOrgTenantId] = createSignal('')
  const [orgPolicyName, setOrgPolicyName] = createSignal('')
  const organizationOptions = () =>
    collection.items().map((item) => ({
      name: `${item.slug} · ${item.name}`,
      value: item.id,
    }))

  return {
    organizations: collection.items,
    organizationsQuery: collection.query,
    organizationsPageInfo: collection.pageInfo,
    organizationsLoading: collection.loading,
    organizationsError: collection.error,
    updateOrganizationsQuery: collection.updateQuery,
    reloadOrganizations: collection.reload,
    selectedOrgId,
    setSelectedOrgId,
    organizationMembers: members.items,
    organizationMembersQuery: members.query,
    organizationMembersPageInfo: members.pageInfo,
    organizationMembersLoading: members.loading,
    organizationMembersError: members.error,
    updateOrganizationMembersQuery: members.updateQuery,
    reloadOrganizationMembers: members.reload,
    orgPolicies,
    setOrgPolicies,
    orgName,
    setOrgName,
    orgSlug,
    setOrgSlug,
    orgTenantId,
    setOrgTenantId,
    orgPolicyName,
    setOrgPolicyName,
    organizationOptions,
  }
}
