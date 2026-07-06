import { removePlatformRole, removeTenantMember, upsertPlatformRole, upsertTenantMember } from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'
import type { RunAction } from '../shared/actions'

export function createAssignmentsActions(state: AdminIamState, runAction: RunAction) {
    const platformUserID = () => Number.parseInt(state.shortcutPlatformUserId().trim(), 10)
    const tenantUserID = () => Number.parseInt(state.shortcutTenantUserId().trim(), 10)

    const handleAssignPlatformRole = () =>
        runAction(async () => {
            const userID = platformUserID()
            const roleName = state.shortcutPlatformRoleName()
            const result = await upsertPlatformRole({
                targetUserId: userID,
                roleName,
            })
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Assigned platform role ${roleName} to user ${userID}.`)
        })

    const handleRemovePlatformRoleShortcut = () =>
        runAction(async () => {
            const userID = platformUserID()
            const roleName = state.shortcutPlatformRoleName()
            const result = await removePlatformRole(userID, roleName)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Removed platform role ${roleName} from user ${userID}.`)
        })

    const handleAssignTenantRole = () =>
        runAction(async () => {
            const userID = tenantUserID()
            const roleName = state.shortcutTenantRoleName()
            const result = await upsertTenantMember({
                tenantId: state.shortcutTenantId().trim(),
                userId: userID,
                roleName,
            })
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Assigned tenant role ${roleName} to user ${userID}.`)
        })

    const handleRemoveTenantMembershipShortcut = () =>
        runAction(async () => {
            const userID = tenantUserID()
            const result = await removeTenantMember(state.shortcutTenantId().trim(), userID)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Removed tenant membership for user ${userID}.`)
        })

    return {
        handleAssignPlatformRole,
        handleRemovePlatformRoleShortcut,
        handleAssignTenantRole,
        handleRemoveTenantMembershipShortcut,
    }
}
