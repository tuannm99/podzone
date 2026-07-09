import { Match, Show, Switch } from 'solid-js'
import { ErrorAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { SelectField } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { Tabs, type TabItem } from '@/solid/components/common/Tabs'
import { ConnectionsPanel } from './connections/ConnectionsPanel'
import { useAdminProvisioning } from './context'
import { PipelinePanel } from './pipeline/PipelinePanel'
import { ResourcesPanel } from './resources/ResourcesPanel'
import type { ProvisioningTab } from './createProvisioningShellViewModel'

const tabs: Array<TabItem<ProvisioningTab>> = [
    { value: 'pipeline', label: 'Pipeline' },
    { value: 'resources', label: 'Resource inventory' },
    { value: 'connections', label: 'Connections' },
]

export function AdminProvisioningView() {
    const vm = useAdminProvisioning()
    const workspaceOptions = () =>
        vm.shell.memberships().map((membership) => ({
            name: `${membership.tenantId} · ${membership.roleName}`,
            value: membership.tenantId,
        }))

    return (
        <PageShell>
            <SectionLead
                eyebrow="Onboarding"
                title="Store provisioning"
                copy="Operate placement inventory, tenant connections, and store delivery pipelines."
            />
            <div class="flex flex-col gap-4 border-b border-gray-200 pb-4 xl:flex-row xl:items-end xl:justify-between">
                <Tabs
                    value={vm.shell.activeTab()}
                    items={tabs}
                    onChange={vm.shell.setActiveTab}
                    variant="underline"
                    ariaLabel="Provisioning console sections"
                />
                <div class="w-full xl:w-96">
                    <SelectField
                        label="Workspace"
                        value={vm.shell.selectedTenantId()}
                        options={workspaceOptions()}
                        onChange={(event) => vm.shell.setSelectedTenantId(event.currentTarget.value)}
                    />
                </div>
            </div>
            <Show when={vm.shell.error()}>
                <ErrorAlert>{vm.shell.error()}</ErrorAlert>
            </Show>
            <Show when={vm.shell.loading()}>
                <LoadingInline label="Loading provisioning access..." />
            </Show>
            <Switch>
                <Match when={vm.shell.activeTab() === 'pipeline'}>
                    <PipelinePanel />
                </Match>
                <Match when={vm.shell.activeTab() === 'resources'}>
                    <ResourcesPanel />
                </Match>
                <Match when={vm.shell.activeTab() === 'connections'}>
                    <ConnectionsPanel />
                </Match>
            </Switch>
        </PageShell>
    )
}
