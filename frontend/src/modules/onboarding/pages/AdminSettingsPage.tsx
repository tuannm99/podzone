import { useNavigate, useSearch } from '@tanstack/solid-router'
import { createSignal } from 'solid-js'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'
import { AdminSettingsView } from './admin-settings/AdminSettingsView'
import { createAuditViewModel } from './admin-settings/audit/createAuditViewModel'
import { AdminSettingsContext, type AdminSettingsTab, type AdminSettingsViewModel } from './admin-settings/context'
import { createInvitesViewModel } from './admin-settings/invites/createInvitesViewModel'
import { createPlatformRolesViewModel } from './admin-settings/platform-roles/createPlatformRolesViewModel'
import { parseUserID } from './admin-settings/presentation'
import { createSessionsViewModel } from './admin-settings/sessions/createSessionsViewModel'
import { createTeamAccessViewModel } from './admin-settings/team-access/createTeamAccessViewModel'
import { createWorkspaceAccessViewModel } from './admin-settings/team-access/createWorkspaceAccessViewModel'

export function createAdminSettingsViewModel(): AdminSettingsViewModel {
    const navigate = useNavigate()
    const search = useSearch({ from: '/admin/settings' })
    const validTabs = new Set<AdminSettingsTab>(['overview', 'sessions', 'team', 'invites', 'audit', 'platform'])
    const activeTab = (): AdminSettingsTab => {
        const tab = search().tab
        return validTabs.has(tab as AdminSettingsTab) ? (tab as AdminSettingsTab) : 'overview'
    }
    const setActiveTab = (tab: AdminSettingsTab) => {
        void navigate({ to: '/admin/settings', search: (prev) => ({ ...prev, tab }) })
    }
    const userID = parseUserID(tokenStorage.getUser()?.id)
    const [routeTenantID, setRouteTenantID] = createSignal(tenantStorage.getTenantID())
    const user = {
        userID,
        hasToken: () => Boolean(tokenStorage.getToken()),
        activeTenantID: () => tokenStorage.getActiveTenantID(),
        sessionID: () => tokenStorage.getSessionID(),
        routeTenantID,
        setRouteTenantID,
    }
    const workspaceAccess = createWorkspaceAccessViewModel(
        userID,
        user.activeTenantID,
        () => activeTab() === 'overview' || activeTab() === 'team' || activeTab() === 'invites'
    )

    return {
        navigation: { activeTab, setActiveTab },
        user,
        workspaceAccess,
        sessions: createSessionsViewModel(user.sessionID, () => activeTab() === 'sessions'),
        audit: createAuditViewModel(() => activeTab() === 'audit'),
        teamAccess: createTeamAccessViewModel(userID, workspaceAccess),
        invites: createInvitesViewModel(workspaceAccess),
        platformRoles: createPlatformRolesViewModel(
            userID,
            () => activeTab() === 'overview' || activeTab() === 'platform'
        ),
    }
}

export default function AdminSettingsPage() {
    const viewModel = createAdminSettingsViewModel()
    return (
        <AdminSettingsContext.Provider value={viewModel}>
            <AdminSettingsView />
        </AdminSettingsContext.Provider>
    )
}
