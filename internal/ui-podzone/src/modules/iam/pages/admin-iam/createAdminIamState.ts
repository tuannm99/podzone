import { createAssignmentsState } from './assignments/createAssignmentsState'
import { createDirectoryViewModel } from './directory/createDirectoryViewModel'
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
  const assignments = createAssignmentsState(userID)
  const directory = createDirectoryViewModel({
    allowed: shell.allowed,
    canManagePlatform: shell.canManagePlatform,
    selectedOrgId: organizations.selectedOrgId,
    selectedTenantId: assignments.shortcutTenantId,
  })
  return {
    ...shell,
    ...organizations,
    ...createPoliciesState(shell.allowed, organizations.selectedOrgId),
    ...createGroupsState(shell.allowed, organizations.selectedOrgId),
    ...assignments,
    ...directory,
    ...createPrincipalsState(userID, platformEnabled),
    ...createTrustSimulationState(userID),
  }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
