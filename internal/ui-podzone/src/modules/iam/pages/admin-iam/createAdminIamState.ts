import { createAssignmentsState } from './assignments/createAssignmentsState'
import { createGroupsState } from './groups/createGroupsState'
import { createOrganizationsState } from './organizations/createOrganizationsState'
import { createPoliciesState } from './policies/createPoliciesState'
import { createPrincipalsState } from './principals/createPrincipalsState'
import { createShellState } from './shared/createShellState'
import { createTrustSimulationState } from './trust-simulation/createTrustSimulationState'

export function createAdminIamState(userID: number) {
  const shell = createShellState()
  const platformEnabled = () => shell.allowed() && shell.canManagePlatform()
  return {
    ...shell,
    ...createOrganizationsState(shell.allowed, shell.setCanManagePlatform),
    ...createPoliciesState(platformEnabled),
    ...createGroupsState(platformEnabled),
    ...createAssignmentsState(userID),
    ...createPrincipalsState(userID, platformEnabled),
    ...createTrustSimulationState(userID),
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
