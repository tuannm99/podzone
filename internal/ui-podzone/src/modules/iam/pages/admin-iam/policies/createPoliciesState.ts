import { createSignal } from 'solid-js'
import type {
  PolicyAttachmentInfo,
  PolicyInfo,
  PolicyVersionInfo,
} from '@/services/iam'
import { prettyJSON } from '../presentation'

export function createPoliciesState() {
  const [policies, setPolicies] = createSignal<PolicyInfo[]>([])
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
    policies().map((item) => ({
      name: `${item.name} · ${item.scope}`,
      value: item.name,
    }))

  return {
    policies,
    setPolicies,
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
