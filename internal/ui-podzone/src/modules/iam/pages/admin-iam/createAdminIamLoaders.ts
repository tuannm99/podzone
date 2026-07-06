import { createGroupsLoaders } from './groups/createGroupsLoaders'
import { createOrganizationsLoaders } from './organizations/createOrganizationsLoaders'
import { createPoliciesLoaders } from './policies/createPoliciesLoaders'
import { createPrincipalsLoaders } from './principals/createPrincipalsLoaders'
import { createBootstrapLoader } from './shared/createBootstrapLoader'
import { createTrustSimulationLoaders } from './trust-simulation/createTrustSimulationLoaders'
import type { AdminIamState } from './createAdminIamState'

export function createAdminIamLoaders(state: AdminIamState, userID: number) {
    return {
        loadBootstrap: createBootstrapLoader(state, userID),
        ...createOrganizationsLoaders(state),
        ...createPoliciesLoaders(state),
        ...createGroupsLoaders(state),
        ...createPrincipalsLoaders(state),
        ...createTrustSimulationLoaders(state),
    }
}

export type AdminIamLoaders = ReturnType<typeof createAdminIamLoaders>
