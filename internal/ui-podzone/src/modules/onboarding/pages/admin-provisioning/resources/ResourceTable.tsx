import { For, Show } from 'solid-js'
import type { CollectionQuery } from '@/services/collection'
import type {
    DatabaseClusterResource,
    KubernetesClusterResource,
    RuntimePoolResource,
} from '@/services/onboarding/provisioning'
import { CollectionControls } from '@/solid/components/common/CollectionControls'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import type { createPaginatedResource } from '@/solid/pagination'
import type { ResourceEditor, ResourcesViewModel } from './createResourcesViewModel'

type Resource = DatabaseClusterResource | KubernetesClusterResource | RuntimePoolResource
type PageModel = ReturnType<typeof createPaginatedResource<Resource>>

function capacity(resource: Resource) {
    if ('namespaces' in resource) {
        const current = resource.namespaces.reduce((sum, namespace) => sum + namespace.current_tenants, 0)
        const maximum = resource.namespaces.reduce((sum, namespace) => sum + namespace.max_tenants, 0)
        return `${current} / ${maximum} tenants`
    }
    return `${resource.current_tenants} / ${resource.max_tenants} tenants`
}

function detail(resource: Resource) {
    if ('engine' in resource) {
        return `${resource.engine} · ${resource.placement_db}`
    }
    if ('namespaces' in resource) {
        return `${resource.namespaces.length} namespaces`
    }
    return resource.kind
}

export function ResourceTable(props: { kind: ResourceEditor['kind']; page: PageModel; vm: ResourcesViewModel }) {
    return (
        <div class="space-y-4">
            <div class="flex items-center justify-between gap-4">
                <p class="text-sm text-gray-500">Inventory consumed by the placement planner.</p>
                <Button size="sm" onClick={() => props.vm.setEditor({ kind: props.kind })}>
                    Add resource
                </Button>
            </div>
            <CollectionControls
                query={props.page.query as CollectionQuery}
                loading={props.page.loading}
                error={props.page.error}
                searchPlaceholder="Name, region, engine, kind, status"
                sortOptions={[
                    { label: 'Updated', value: 'updatedAt' },
                    { label: 'Name', value: 'name' },
                    { label: 'Status', value: 'status' },
                ]}
                filterFields={[
                    {
                        label: 'Status',
                        value: 'status',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                    {
                        label: 'Health',
                        value: 'healthy',
                        operators: ['FILTER_OPERATOR_EQ'],
                    },
                ]}
                updateQuery={props.page.updateQuery}
            />
            <Show
                when={props.page.items().length}
                fallback={
                    <EmptyBlock
                        title="No resources configured"
                        copy="Add capacity before approving a store provisioning request."
                    />
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Resource</TableHeaderCell>
                            <TableHeaderCell>Capacity</TableHeaderCell>
                            <TableHeaderCell>Status</TableHeaderCell>
                            <TableHeaderCell />
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={props.page.items()}>
                            {(resource) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-medium text-gray-950">{resource.name}</p>
                                        <p class="text-xs text-gray-500">{detail(resource)}</p>
                                    </TableCell>
                                    <TableCell>{capacity(resource)}</TableCell>
                                    <TableCell>
                                        <div class="flex flex-wrap gap-2">
                                            <Badge
                                                content={resource.status}
                                                color={resource.status === 'active' ? 'green' : 'dark'}
                                            />
                                            <Badge
                                                content={resource.healthy ? 'healthy' : 'unhealthy'}
                                                color={resource.healthy ? 'blue' : 'red'}
                                            />
                                        </div>
                                    </TableCell>
                                    <TableCell class="text-right">
                                        <div class="flex justify-end gap-2">
                                            <Button
                                                size="xs"
                                                color="alternative"
                                                onClick={() =>
                                                    props.vm.setEditor({
                                                        kind: props.kind,
                                                        value: resource as never,
                                                    })
                                                }
                                            >
                                                Edit
                                            </Button>
                                            <Button
                                                size="xs"
                                                color="red"
                                                loading={props.vm.deletingName() === resource.name}
                                                onClick={() => void props.vm.remove(props.kind, resource.name)}
                                            >
                                                Archive
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={props.page.pageInfo().page}
                    pageSize={props.page.pageInfo().pageSize}
                    total={props.page.pageInfo().total}
                    loading={props.page.loading()}
                    onPageChange={(page) => props.page.updateQuery({ page })}
                />
            </Show>
        </div>
    )
}
