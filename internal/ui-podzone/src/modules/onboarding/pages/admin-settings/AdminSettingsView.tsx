import { PageShell } from '@/solid/components/common/PageShell'
import { HeaderRuntime } from './HeaderRuntime'
import { SessionsAudit } from './SessionsAudit'
import { InvitesPanel } from './invites/InvitesPanel'
import { PlatformAdminPanel } from './platform-roles/PlatformAdminPanel'
import { TeamAccessPanel } from './team-access/TeamAccessPanel'

export function AdminSettingsView() {
  return (
    <PageShell>
      <HeaderRuntime />
      <SessionsAudit />
      <TeamAccessPanel />
      <InvitesPanel />
      <PlatformAdminPanel />
    </PageShell>
  )
}
