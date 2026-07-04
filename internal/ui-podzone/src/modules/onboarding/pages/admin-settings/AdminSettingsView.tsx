import { Match, Switch } from 'solid-js'
import { PageShell } from '@/solid/components/common/PageShell'
import { HeaderRuntime } from './HeaderRuntime'
import { AuditPanel } from './audit/AuditPanel'
import { InvitesPanel } from './invites/InvitesPanel'
import { PlatformAdminPanel } from './platform-roles/PlatformAdminPanel'
import { SessionsPanel } from './sessions/SessionsPanel'
import { SettingsTabs } from './SettingsTabs'
import { TeamAccessPanel } from './team-access/TeamAccessPanel'
import { useAdminSettings } from './context'

export function AdminSettingsView() {
  const settings = useAdminSettings()

  return (
    <PageShell>
      <header class="flex flex-col gap-2 border-b border-gray-200 pb-5">
        <div class="text-xs font-semibold uppercase text-gray-500">
          Settings
        </div>
        <h1 class="text-2xl font-semibold text-gray-950">
          Workspace and platform controls
        </h1>
      </header>

      <SettingsTabs
        value={settings.navigation.activeTab()}
        onChange={settings.navigation.setActiveTab}
      />

      <div role="tabpanel" class="min-w-0 pt-1">
        <Switch>
          <Match when={settings.navigation.activeTab() === 'overview'}>
            <HeaderRuntime />
          </Match>
          <Match when={settings.navigation.activeTab() === 'sessions'}>
            <SessionsPanel />
          </Match>
          <Match when={settings.navigation.activeTab() === 'team'}>
            <TeamAccessPanel />
          </Match>
          <Match when={settings.navigation.activeTab() === 'invites'}>
            <InvitesPanel />
          </Match>
          <Match when={settings.navigation.activeTab() === 'audit'}>
            <AuditPanel />
          </Match>
          <Match when={settings.navigation.activeTab() === 'platform'}>
            <PlatformAdminPanel />
          </Match>
        </Switch>
      </div>
    </PageShell>
  )
}
