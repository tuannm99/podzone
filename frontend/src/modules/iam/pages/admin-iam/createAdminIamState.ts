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
    const organizations = createOrganizationsState(shell.allowed, shell.setCanManagePlatform)
    const assignments = createAssignmentsState(userID)
    const policies = createPoliciesState(shell.allowed, organizations.selectedOrgId)
    const groups = createGroupsState(shell.allowed, organizations.selectedOrgId)
    const principals = createPrincipalsState(userID, platformEnabled)
    const trustSimulation = createTrustSimulationState(userID)
    const directory = createDirectoryViewModel({
        allowed: shell.allowed,
        canManagePlatform: shell.canManagePlatform,
        selectedOrgId: organizations.selectedOrgId,
        selectedTenantId: assignments.shortcutTenantId,
        visibleUserIds: () => [
            ...organizations.organizationMembers().map((member) => member.userId),
            ...groups.groupMembers(),
            ...policies.policyAttachments().map((attachment) => attachment.userId),
        ],
    })
    return {
        ...shell,
        ...organizations,
        ...policies,
        ...groups,
        ...assignments,
        ...directory,
        ...principals,
        ...trustSimulation,
    }
}

export type AdminIamState = ReturnType<typeof createAdminIamState>
