import { listServiceControlPolicies } from '@podzone/shared/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createOrganizationsLoaders(state: AdminIamState) {
    let requestID = 0

    const loadSelectedOrganization = async () => {
        const currentRequest = ++requestID
        const organizationID = state.selectedOrgId().trim()
        if (!organizationID) {
            state.setOrgPolicies([])
            return
        }
        const result = await listServiceControlPolicies(organizationID)
        if (currentRequest !== requestID || organizationID !== state.selectedOrgId().trim()) return
        if (!result.success) {
            state.setPageError(result.message)
            return
        }
        state.setOrgPolicies(result.data)
    }

    return { loadSelectedOrganization }
}
