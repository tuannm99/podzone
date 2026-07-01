import { checkPlatformPermission, listUserTenants } from '@/services/iam'
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

    const tenantResult = await listUserTenants(userID)
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
    state.setLoading(false)
  }
}
