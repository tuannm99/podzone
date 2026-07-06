import type { AdminIamLoaders } from './createAdminIamLoaders'
import type { AdminIamState } from './createAdminIamState'
import { createAssignmentsActions } from './assignments/createAssignmentsActions'
import { createGroupsActions } from './groups/createGroupsActions'
import { createOrganizationsActions } from './organizations/createOrganizationsActions'
import { createPoliciesActions } from './policies/createPoliciesActions'
import { createPrincipalsActions } from './principals/createPrincipalsActions'
import { createTrustSimulationActions } from './trust-simulation/createTrustSimulationActions'

export function createAdminIamActions(state: AdminIamState, loaders: AdminIamLoaders) {
    const runAction = async (work: () => Promise<void>) => {
        state.setPageError('')
        state.setPageMessage('')
        try {
            await work()
        } catch (error) {
            state.setPageError(error instanceof Error ? error.message : 'Action failed')
        }
    }

    return {
        ...createOrganizationsActions(state, loaders, runAction),
        ...createPoliciesActions(state, loaders, runAction),
        ...createGroupsActions(state, loaders, runAction),
        ...createPrincipalsActions(state, loaders, runAction),
        ...createTrustSimulationActions(state, loaders, runAction),
        ...createAssignmentsActions(state, runAction),
    }
}

export type AdminIamActions = ReturnType<typeof createAdminIamActions>
