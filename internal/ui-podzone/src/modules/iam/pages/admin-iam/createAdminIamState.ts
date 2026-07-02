import { createAssignmentsState } from './assignments/createAssignmentsState'
import { createGroupsState } from './groups/createGroupsState'
import { createOrganizationsState } from './organizations/createOrganizationsState'
import { createPoliciesState } from './policies/createPoliciesState'
import { createPrincipalsState } from './principals/createPrincipalsState'
import { createShellState } from './shared/createShellState'
import { createTrustSimulationState } from './trust-simulation/createTrustSimulationState'

export function createAdminIamState(userID: number) {
  const shell = createShellState()
  return {
    ...shell,
    ...createOrganizationsState(shell.allowed),
    ...createPoliciesState(shell.allowed),
    ...createGroupsState(shell.allowed),
    ...createAssignmentsState(userID),
    ...createPrincipalsState(userID, shell.allowed),
    ...createTrustSimulationState(userID),
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
