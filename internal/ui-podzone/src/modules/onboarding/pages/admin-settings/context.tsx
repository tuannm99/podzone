import { createContext, useContext, type Accessor, type Setter } from 'solid-js'
import type { AuditViewModel } from './audit/createAuditViewModel'
import type { InvitesViewModel } from './invites/createInvitesViewModel'
import type { PlatformRolesViewModel } from './platform-roles/createPlatformRolesViewModel'
import type { SessionsViewModel } from './sessions/createSessionsViewModel'
import type { TeamAccessViewModel } from './team-access/createTeamAccessViewModel'
import type { WorkspaceAccessViewModel } from './team-access/createWorkspaceAccessViewModel'

export type AdminSettingsViewModel = {
  navigation: {
    activeTab: Accessor<AdminSettingsTab>
    setActiveTab: (tab: AdminSettingsTab) => void
  }
  user: {
    userID: number
    hasToken: Accessor<boolean>
    activeTenantID: Accessor<string>
    sessionID: Accessor<string>
    routeTenantID: Accessor<string>
    setRouteTenantID: Setter<string>
  }
  workspaceAccess: WorkspaceAccessViewModel
  sessions: SessionsViewModel
  audit: AuditViewModel
  teamAccess: TeamAccessViewModel
  invites: InvitesViewModel
  platformRoles: PlatformRolesViewModel
}

export type AdminSettingsTab =
  'overview' | 'sessions' | 'team' | 'invites' | 'audit' | 'platform'

export const AdminSettingsContext = createContext<AdminSettingsViewModel>()

export function useAdminSettings() {
  const value = useContext(AdminSettingsContext)
  if (!value) {
    throw new Error(
      'useAdminSettings must be used inside AdminSettingsContext.Provider'
    )
  }
  return value
}
