import { createAssignmentsState } from './assignments/createAssignmentsState'
import { createGroupsState } from './groups/createGroupsState'
import { createOrganizationsState } from './organizations/createOrganizationsState'
import { createPoliciesState } from './policies/createPoliciesState'
import { createPrincipalsState } from './principals/createPrincipalsState'
import { createShellState } from './shared/createShellState'
import { createTrustSimulationState } from './trust-simulation/createTrustSimulationState'

export function createAdminIamState(userID: number) {
  return {
    ...createShellState(),
    ...createOrganizationsState(),
    ...createPoliciesState(),
    ...createGroupsState(),
    ...createAssignmentsState(userID),
    ...createPrincipalsState(userID),
    ...createTrustSimulationState(userID),
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
