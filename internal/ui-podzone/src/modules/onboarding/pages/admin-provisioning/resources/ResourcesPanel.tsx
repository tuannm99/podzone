import { createSignal, Match, Show, Switch } from 'solid-js'
import { ErrorAlert } from '@/solid/components/common/Feedback'
import { Tabs, type TabItem } from '@/solid/components/common/Tabs'
import { useAdminProvisioning } from '../context'
import type { ResourceEditor } from './createResourcesViewModel'
import { ResourceEditor as ResourceEditorDrawer } from './ResourceEditor'
import { ResourceTable } from './ResourceTable'

type ResourceKind = ResourceEditor['kind']

const tabs: Array<TabItem<ResourceKind>> = [
    { value: 'database-clusters', label: 'Database clusters' },
    { value: 'kubernetes-clusters', label: 'Kubernetes clusters' },
    { value: 'runtime-pools', label: 'Runtime pools' },
]

export function ResourcesPanel() {
    const { resources } = useAdminProvisioning()
    const [activeKind, setActiveKind] = createSignal<ResourceKind>('database-clusters')

    return (
        <section class="space-y-5 rounded-lg border border-gray-200 bg-white p-5">
            <div>
                <h2 class="text-base font-semibold text-gray-950">Placement resource inventory</h2>
                <p class="mt-1 text-sm text-gray-500">
                    Global capacity available to onboarding allocation and provisioning.
                </p>
            </div>
            <Tabs
                value={activeKind()}
                items={tabs}
                onChange={setActiveKind}
                variant="underline"
                ariaLabel="Resource inventory kinds"
            />
            <Show when={resources.error()}>
                <ErrorAlert>{resources.error()}</ErrorAlert>
            </Show>
            <Switch>
                <Match when={activeKind() === 'database-clusters'}>
                    <ResourceTable kind="database-clusters" page={resources.databaseClusters} vm={resources} />
                </Match>
                <Match when={activeKind() === 'kubernetes-clusters'}>
                    <ResourceTable kind="kubernetes-clusters" page={resources.kubernetesClusters} vm={resources} />
                </Match>
                <Match when={activeKind() === 'runtime-pools'}>
                    <ResourceTable kind="runtime-pools" page={resources.runtimePools} vm={resources} />
                </Match>
            </Switch>
            <ResourceEditorDrawer editor={resources.editor()} vm={resources} />
        </section>
    )
}
