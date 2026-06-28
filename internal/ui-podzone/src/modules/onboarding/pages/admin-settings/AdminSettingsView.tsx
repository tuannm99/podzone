import { PageShell } from '@/solid/components/common/PageShell'
import { HeaderRuntime } from './HeaderRuntime'
import { InvitesPanel } from './InvitesPanel'
import { PlatformAdminPanel } from './PlatformAdminPanel'
import { SessionsAudit } from './SessionsAudit'
import { TeamAccess } from './TeamAccess'

export function AdminSettingsView() {
  return (
    <PageShell>
      <HeaderRuntime />
      <SessionsAudit />
      <TeamAccess />
      <InvitesPanel />
      <PlatformAdminPanel />
    </PageShell>
  )
}
