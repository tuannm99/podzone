import { getRolePermissionBoundary, getRoleTrustPolicy } from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'
import { prettyJSON } from '../presentation'

export function createTrustSimulationLoaders(state: AdminIamState) {
  let requestID = 0

  const loadTrustPolicy = async () => {
    const currentRequest = ++requestID
    const roleName = state.trustRoleName().trim()
    if (!roleName) return
    const result = await getRoleTrustPolicy(roleName)
    if (
      currentRequest !== requestID ||
      roleName !== state.trustRoleName().trim()
    )
      return
    if (!result.success) {
      state.setPageError(result.message)
      return
    }
    state.setTrustJson(prettyJSON(result.data))
    const boundaryResult = await getRolePermissionBoundary(roleName)
    if (
      currentRequest !== requestID ||
      roleName !== state.trustRoleName().trim()
    )
      return
    if (boundaryResult.success) {
      state.setRoleBoundary(boundaryResult.data)
      state.setTrustBoundaryPolicyName(boundaryResult.data?.policyName || '')
    } else {
      state.setPageError(boundaryResult.message)
    }
  }

  return { loadTrustPolicy }
}
