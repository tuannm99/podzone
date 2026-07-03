import { PageShell } from '@/solid/components/common/PageShell'
import { AttentionRuntime } from './AttentionRuntime'
import { HeaderStats } from './HeaderStats'
import { ProvisioningRequestsPanel } from './ProvisioningRequestsPanel'
import { StoreChooser } from './StoreChooser'
import { WorkspaceList } from './WorkspaceList'
import { WorkspaceSetup } from './WorkspaceSetup'

export function AdminHomeView() {
  return (
    <PageShell>
      <HeaderStats />
      <StoreChooser />
      <ProvisioningRequestsPanel />
      <WorkspaceSetup />
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <WorkspaceList />
      </div>
      <AttentionRuntime />
    </PageShell>
  )
}
