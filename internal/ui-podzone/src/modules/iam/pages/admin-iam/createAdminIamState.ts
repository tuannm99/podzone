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
  const organizations = createOrganizationsState(
    shell.allowed,
    shell.setCanManagePlatform
  )
  return {
    ...shell,
    ...organizations,
    ...createPoliciesState(shell.allowed, organizations.selectedOrgId),
    ...createGroupsState(shell.allowed, organizations.selectedOrgId),
    ...createAssignmentsState(userID),
    ...createPrincipalsState(userID, platformEnabled),
    ...createTrustSimulationState(userID),
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
