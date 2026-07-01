import {
  checkPlatformPermission,
  listOrganizations,
  listPolicies,
  listUserTenants,
} from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createBootstrapLoader(state: AdminIamState, userID: number) {
  let requestID = 0

  return async function loadBootstrap() {
    const currentRequest = ++requestID
    state.setLoading(true)
    state.setPageError('')
    const permission = await checkPlatformPermission('platform:manage_roles')
    if (currentRequest !== requestID) return
    if (!permission.success) {
      state.setLoading(false)
      state.setPageError(permission.message)
      return
    }
    state.setAllowed(permission.data)
    if (!permission.data) {
      state.setLoading(false)
      state.setPageError('Missing permission: platform:manage_roles')
      return
    }

    const [tenantResult, orgResult, policyResult] = await Promise.all([
      listUserTenants(userID),
      listOrganizations(),
      listPolicies(),
    ])
    if (currentRequest !== requestID) return

    if (!tenantResult.success) {
      state.setPageError(tenantResult.message)
    } else {
      state.setMemberships(tenantResult.data)
      const firstTenantID = tenantResult.data[0]?.tenantId
      if (!state.orgTenantId() && firstTenantID)
        state.setOrgTenantId(firstTenantID)
      if (!state.groupTenantId() && firstTenantID)
        state.setGroupTenantId(firstTenantID)
      if (!state.simTenantId() && firstTenantID)
        state.setSimTenantId(firstTenantID)
      if (!state.principalTenantId() && firstTenantID)
        state.setPrincipalTenantId(firstTenantID)
      if (!state.shortcutTenantId() && firstTenantID)
        state.setShortcutTenantId(firstTenantID)
    }
    if (!orgResult.success) {
      state.setPageError(orgResult.message)
    } else {
      state.setOrganizations(orgResult.data)
      if (!state.selectedOrgId() && orgResult.data[0]?.id)
        state.setSelectedOrgId(orgResult.data[0].id)
    }
    if (!policyResult.success) {
      state.setPageError(policyResult.message)
    } else {
      state.setPolicies(policyResult.data)
      if (!state.selectedPolicyName() && policyResult.data[0]?.name)
        state.setSelectedPolicyName(policyResult.data[0].name)
    }
    state.setLoading(false)
  }
}
