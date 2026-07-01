import {
  getPolicy,
  listPolicyAttachments,
  listPolicyVersions,
} from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createPoliciesLoaders(state: AdminIamState) {
  let requestID = 0

  const loadSelectedPolicy = async () => {
    const currentRequest = ++requestID
    const name = state.selectedPolicyName().trim()
    if (!name) {
      state.setPolicyDetail(undefined)
      state.setPolicyVersions([])
      state.setPolicyAttachments([])
      return
    }
    const [policyResult, versionsResult, attachmentsResult] = await Promise.all(
      [getPolicy(name), listPolicyVersions(name), listPolicyAttachments(name)]
    )
    if (
      currentRequest !== requestID ||
      name !== state.selectedPolicyName().trim()
    )
      return
    if (policyResult.success) state.setPolicyDetail(policyResult.data)
    else state.setPageError(policyResult.message)
    if (versionsResult.success) state.setPolicyVersions(versionsResult.data)
    else state.setPageError(versionsResult.message)
    if (attachmentsResult.success)
      state.setPolicyAttachments(attachmentsResult.data)
    else state.setPageError(attachmentsResult.message)
  }

  return { loadSelectedPolicy }
}
