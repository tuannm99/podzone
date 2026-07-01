import {
  getPlatformUserPermissionBoundary,
  getTenantUserPermissionBoundary,
  listPlatformUserInlinePolicies,
  listPlatformUserPolicies,
  listTenantUserInlinePolicies,
  listTenantUserPolicies,
} from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createPrincipalsLoaders(state: AdminIamState) {
  let requestID = 0

  const loadPrincipalControls = async () => {
    if (!state.allowed()) return
    const currentRequest = ++requestID
    if (state.principalMode() === 'platform') {
      const userID = Number.parseInt(state.principalPlatformUserId().trim(), 10)
      if (!Number.isFinite(userID) || userID <= 0) {
        state.setPlatformUserPolicies([])
        state.setPlatformUserInlinePolicies([])
        state.setPlatformUserBoundary(null)
        return
      }
      const [policiesResult, inlineResult, boundaryResult] = await Promise.all([
        listPlatformUserPolicies(userID),
        listPlatformUserInlinePolicies(userID),
        getPlatformUserPermissionBoundary(userID),
      ])
      if (
        currentRequest !== requestID ||
        state.principalMode() !== 'platform' ||
        userID !== Number.parseInt(state.principalPlatformUserId().trim(), 10)
      )
        return
      if (policiesResult.success)
        state.setPlatformUserPolicies(policiesResult.data)
      else state.setPageError(policiesResult.message)
      if (inlineResult.success)
        state.setPlatformUserInlinePolicies(inlineResult.data)
      else state.setPageError(inlineResult.message)
      if (boundaryResult.success) {
        state.setPlatformUserBoundary(boundaryResult.data)
        state.setPrincipalBoundaryPolicyName(
          boundaryResult.data?.policyName || ''
        )
      } else state.setPageError(boundaryResult.message)
      return
    }

    const userID = Number.parseInt(state.principalTenantUserId().trim(), 10)
    const tenantID = state.principalTenantId().trim()
    if (!tenantID || !Number.isFinite(userID) || userID <= 0) {
      state.setTenantUserPolicies([])
      state.setTenantUserInlinePolicies([])
      state.setTenantUserBoundary(null)
      return
    }
    const [policiesResult, inlineResult, boundaryResult] = await Promise.all([
      listTenantUserPolicies(tenantID, userID),
      listTenantUserInlinePolicies(tenantID, userID),
      getTenantUserPermissionBoundary(tenantID, userID),
    ])
    if (
      currentRequest !== requestID ||
      state.principalMode() !== 'tenant' ||
      tenantID !== state.principalTenantId().trim() ||
      userID !== Number.parseInt(state.principalTenantUserId().trim(), 10)
    )
      return
    if (policiesResult.success) state.setTenantUserPolicies(policiesResult.data)
    else state.setPageError(policiesResult.message)
    if (inlineResult.success)
      state.setTenantUserInlinePolicies(inlineResult.data)
    else state.setPageError(inlineResult.message)
    if (boundaryResult.success) {
      state.setTenantUserBoundary(boundaryResult.data)
      state.setPrincipalBoundaryPolicyName(
        boundaryResult.data?.policyName || ''
      )
    } else state.setPageError(boundaryResult.message)
  }

  return { loadPrincipalControls }
}
