import { getPolicy } from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createPoliciesLoaders(state: AdminIamState) {
    let requestID = 0

    const loadSelectedPolicy = async () => {
        if (!state.allowed()) return
        const currentRequest = ++requestID
        const name = state.selectedPolicyName().trim()
        if (!name) {
            state.setPolicyDetail(undefined)
            state.clearPolicyVersions()
            state.clearPolicyAttachments()
            return
        }
        const policyResult = await getPolicy(state.policyRef(name))
        if (currentRequest !== requestID || name !== state.selectedPolicyName().trim()) return
        if (policyResult.success) state.setPolicyDetail(policyResult.data)
        else state.setPageError(policyResult.message)
        await Promise.all([state.reloadPolicyVersions(), state.reloadPolicyAttachments()])
    }

    return { loadSelectedPolicy }
}
