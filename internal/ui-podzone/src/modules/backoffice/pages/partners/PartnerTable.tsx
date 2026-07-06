import { For, Show, type Accessor } from 'solid-js'
import type { PageInfo } from '@/services/collection'
import type { PartnerInfo } from '@/services/partner'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock, LoadingInline } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { badgeColorForStatus, joinCapabilityList, partnerTypeLabel } from './presentation'

type PartnerTableProps = {
    tenantID: string
    partners: Accessor<PartnerInfo[]>
    pageInfo: Accessor<PageInfo>
    page: Accessor<number>
    loading: Accessor<boolean>
    onPageChange: (page: number) => void
    onEdit: (partner: PartnerInfo) => void
    onToggleStatus: (partner: PartnerInfo) => void
}

export function PartnerTable(props: PartnerTableProps) {
    return (
        <>
            <Show when={props.loading()}>
                <LoadingInline label="Loading partners..." />
            </Show>
            <Show
                when={!props.loading() && props.partners().length > 0}
                fallback={
                    !props.loading() ? (
                        <EmptyBlock title="No partners found" copy="No partner records match the current filters." />
                    ) : null
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Partner</TableHeaderCell>
                            <TableHeaderCell>Capabilities</TableHeaderCell>
                            <TableHeaderCell>Routing</TableHeaderCell>
                            <TableHeaderCell>Status</TableHeaderCell>
                            <TableHeaderCell class="text-right">Actions</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={props.partners()}>
                            {(partner) => (
                                <TableRow>
                                    <TableCell>
                                        <p class="font-semibold text-gray-900">{partner.name}</p>
                                        <p class="mt-1 text-xs text-gray-500">
                                            {partner.code} · {partnerTypeLabel(partner.partnerType)}
                                        </p>
                                        <Show when={partner.contactEmail}>
                                            <p class="mt-1 text-xs text-gray-500">{partner.contactEmail}</p>
                                        </Show>
                                    </TableCell>
                                    <TableCell class="max-w-56 text-gray-600">
                                        <p>{joinCapabilityList(partner.supportedProductTypes) || 'Any product'}</p>
                                        <p class="mt-1 text-xs text-gray-500">
                                            {joinCapabilityList(partner.supportedRegions) || 'Any region'}
                                        </p>
                                    </TableCell>
                                    <TableCell class="whitespace-nowrap text-gray-600">
                                        <p>Priority {partner.routingPriority || 0}</p>
                                        <p class="mt-1 text-xs text-gray-500">SLA {partner.slaDays || 0} days</p>
                                    </TableCell>
                                    <TableCell>
                                        <Badge content={partner.status} color={badgeColorForStatus(partner.status)} />
                                    </TableCell>
                                    <TableCell>
                                        <div class="flex flex-wrap justify-end gap-2">
                                            <Button
                                                size="xs"
                                                color="alternative"
                                                href={`/t/${props.tenantID}/partners/${partner.id}`}
                                            >
                                                View
                                            </Button>
                                            <Button size="xs" color="light" onClick={() => props.onEdit(partner)}>
                                                Edit
                                            </Button>
                                            <Button
                                                size="xs"
                                                color={partner.status === 'active' ? 'alternative' : 'green'}
                                                onClick={() => props.onToggleStatus(partner)}
                                            >
                                                {partner.status === 'active' ? 'Deactivate' : 'Activate'}
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={props.page()}
                    pageSize={props.pageInfo().pageSize}
                    total={props.pageInfo().total}
                    onPageChange={props.onPageChange}
                />
            </Show>
        </>
    )
}
