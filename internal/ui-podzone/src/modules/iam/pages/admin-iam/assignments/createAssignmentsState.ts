import { createSignal } from 'solid-js'
import { platformRoleOptions, tenantRoleOptions } from '../presentation'

export function createAssignmentsState(userID: number) {
    const [shortcutPlatformUserId, setShortcutPlatformUserId] = createSignal(userID ? String(userID) : '')
    const [shortcutPlatformRoleName, setShortcutPlatformRoleName] = createSignal(platformRoleOptions[1].value)
    const [shortcutTenantId, setShortcutTenantId] = createSignal('')
    const [shortcutTenantUserId, setShortcutTenantUserId] = createSignal('')
    const [shortcutTenantRoleName, setShortcutTenantRoleName] = createSignal(tenantRoleOptions[1].value)

    return {
        shortcutPlatformUserId,
        setShortcutPlatformUserId,
        shortcutPlatformRoleName,
        setShortcutPlatformRoleName,
        shortcutTenantId,
        setShortcutTenantId,
        shortcutTenantUserId,
        setShortcutTenantUserId,
        shortcutTenantRoleName,
        setShortcutTenantRoleName,
    }
}
