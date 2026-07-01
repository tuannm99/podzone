import { createSignal, type Accessor } from 'solid-js'
import type {
  PolicyAttachmentInfo,
  PolicyInfo,
  PolicyVersionInfo,
} from '@/services/iam'
import { listPolicies } from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'
import { prettyJSON } from '../presentation'

export function createPoliciesState(enabled: Accessor<boolean>) {
  const collection = createPaginatedResource<PolicyInfo>(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listPolicies(query)
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    { enabled }
  )
  const [selectedPolicyName, setSelectedPolicyName] = createSignal('')
  const [policyDetail, setPolicyDetail] = createSignal<PolicyInfo>()
  const [policyVersions, setPolicyVersions] = createSignal<PolicyVersionInfo[]>(
    []
  )
  const [policyAttachments, setPolicyAttachments] = createSignal<
    PolicyAttachmentInfo[]
  >([])
  const [policyScope, setPolicyScope] = createSignal('platform')
  const [policyName, setPolicyName] = createSignal('')
  const [policyDescription, setPolicyDescription] = createSignal('')
  const [policyStatementsJson, setPolicyStatementsJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:read',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const [policyVersionJson, setPolicyVersionJson] = createSignal(
    prettyJSON([
      {
        effect: 'allow',
        actionPattern: 'order:update',
        resourcePattern: '*',
        conditions: [],
      },
    ])
  )
  const policyOptions = () =>
    collection.items().map((item) => ({
      name: `${item.name} · ${item.scope}`,
      value: item.name,
    }))

  return {
    policies: collection.items,
    policiesQuery: collection.query,
    policiesPageInfo: collection.pageInfo,
    policiesLoading: collection.loading,
    policiesError: collection.error,
    updatePoliciesQuery: collection.updateQuery,
    reloadPolicies: collection.reload,
    selectedPolicyName,
    setSelectedPolicyName,
    policyDetail,
    setPolicyDetail,
    policyVersions,
    setPolicyVersions,
    policyAttachments,
    setPolicyAttachments,
    policyScope,
    setPolicyScope,
    policyName,
    setPolicyName,
    policyDescription,
    setPolicyDescription,
    policyStatementsJson,
    setPolicyStatementsJson,
    policyVersionJson,
    setPolicyVersionJson,
    policyOptions,
  }
}
