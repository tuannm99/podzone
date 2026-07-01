import { createSignal, type Accessor } from 'solid-js'
import {
  listOrganizations,
  type OrganizationInfo,
  type PolicyInfo,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'

export function createOrganizationsState(enabled: Accessor<boolean>) {
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
      return result.data
    },
    { enabled }
  )
  const [selectedOrgId, setSelectedOrgId] = createSignal('')
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
