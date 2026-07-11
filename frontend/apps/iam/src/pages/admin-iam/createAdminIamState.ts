import { type Accessor } from 'solid-js'
import { createAssignmentsState } from './assignments/createAssignmentsState'
import { createDirectoryViewModel } from './directory/createDirectoryViewModel'
import { createGroupsState } from './groups/createGroupsState'
import { createOrganizationsState } from './organizations/createOrganizationsState'
import { createPoliciesState } from './policies/createPoliciesState'
import { createPrincipalsState } from './principals/createPrincipalsState'
import { createShellState } from './shared/createShellState'
import { createTrustSimulationState } from './trust-simulation/createTrustSimulationState'
import type { IamSectionID } from './presentation'

export function createAdminIamState(userID: number, activeSection: Accessor<IamSectionID>) {
    const shell = createShellState()
    const onSection = (id: IamSectionID) => () => shell.allowed() && activeSection() === id
    const platformEnabled = () => shell.allowed() && shell.canManagePlatform()

    const organizations = createOrganizationsState(onSection('iam-orgs'), shell.setCanManagePlatform)
    const assignments = createAssignmentsState(userID)
    const policies = createPoliciesState(onSection('iam-policies'), organizations.selectedOrgId)
    const groups = createGroupsState(onSection('iam-groups'), organizations.selectedOrgId)
    const principals = createPrincipalsState(userID, () => platformEnabled() && activeSection() === 'iam-principals')
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
