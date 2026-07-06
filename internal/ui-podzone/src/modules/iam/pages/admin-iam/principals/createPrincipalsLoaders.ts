import { getPlatformUserPermissionBoundary, getTenantUserPermissionBoundary } from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createPrincipalsLoaders(state: AdminIamState) {
    let requestID = 0

    const loadPrincipalControls = async () => {
        if (!state.allowed()) return
        const currentRequest = ++requestID
        if (state.principalMode() === 'platform') {
            const userID = Number.parseInt(state.principalPlatformUserId().trim(), 10)
            if (!Number.isFinite(userID) || userID <= 0) {
                state.clearPrincipalManagedPolicies()
                state.clearPrincipalInlinePolicies()
                state.setPlatformUserBoundary(null)
                return
            }
            const [boundaryResult] = await Promise.all([
                getPlatformUserPermissionBoundary(userID),
                state.reloadPrincipalManagedPolicies(),
                state.reloadPrincipalInlinePolicies(),
            ])
            if (
                currentRequest !== requestID ||
                state.principalMode() !== 'platform' ||
                userID !== Number.parseInt(state.principalPlatformUserId().trim(), 10)
            )
                return
            if (boundaryResult.success) {
                state.setPlatformUserBoundary(boundaryResult.data)
                state.setPrincipalBoundaryPolicyName(boundaryResult.data?.policyName || '')
            } else state.setPageError(boundaryResult.message)
            return
        }

        const userID = Number.parseInt(state.principalTenantUserId().trim(), 10)
        const tenantID = state.principalTenantId().trim()
        if (!tenantID || !Number.isFinite(userID) || userID <= 0) {
            state.clearPrincipalManagedPolicies()
            state.clearPrincipalInlinePolicies()
            state.setTenantUserBoundary(null)
            return
        }
        const [boundaryResult] = await Promise.all([
            getTenantUserPermissionBoundary(tenantID, userID),
            state.reloadPrincipalManagedPolicies(),
            state.reloadPrincipalInlinePolicies(),
        ])
        if (
            currentRequest !== requestID ||
            state.principalMode() !== 'tenant' ||
            tenantID !== state.principalTenantId().trim() ||
            userID !== Number.parseInt(state.principalTenantUserId().trim(), 10)
        )
            return
        if (boundaryResult.success) {
            state.setTenantUserBoundary(boundaryResult.data)
            state.setPrincipalBoundaryPolicyName(boundaryResult.data?.policyName || '')
        } else state.setPageError(boundaryResult.message)
    }

    return { loadPrincipalControls }
}
