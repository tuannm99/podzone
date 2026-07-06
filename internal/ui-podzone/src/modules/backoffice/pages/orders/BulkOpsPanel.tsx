import { For, Show } from 'solid-js'
import { Button, InputField, SelectField } from '@/solid/components/common/Primitives'
import { useTenantOrdersBoard, type ShipmentSlaMode } from './board-context'

const settlementOptions = [
    { name: 'Pending', value: 'pending' },
    { name: 'Reconciled', value: 'reconciled' },
    { name: 'Paid', value: 'paid' },
    { name: 'Disputed', value: 'disputed' },
]

export function BulkOpsPanel() {
    const board = useTenantOrdersBoard()

    return (
        <div class="mt-4 rounded-lg border border-gray-200 bg-white p-4">
            <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                    <p class="text-sm font-semibold text-gray-900">Bulk ops</p>
                    <p class="text-sm text-gray-500">
                        Selected {board.selectedOrderIDs().length} order(s) in the current queue workflow.
                    </p>
                </div>
                <div class="flex flex-wrap gap-2">
                    <Button type="button" size="xs" color="alternative" onClick={board.selectVisibleOrders}>
                        Select visible
                    </Button>
                    <Button type="button" size="xs" color="light" onClick={board.clearSelectedOrders}>
                        Clear
                    </Button>
                </div>
            </div>
            <div class="mt-4 grid gap-4 md:grid-cols-3">
                <InputField
                    label="Bulk owner"
                    value={board.bulkDraft().operatorAssignee}
                    placeholder="linh.nguyen"
                    onInput={(event) =>
                        board.setBulkDraft((current) => ({
                            ...current,
                            operatorAssignee: event.currentTarget.value,
                        }))
                    }
                />
                <InputField
                    label="Bulk shipment SLA"
                    type="datetime-local"
                    value={board.bulkDraft().shipmentSlaDueAt}
                    onInput={(event) =>
                        board.setBulkDraft((current) => ({
                            ...current,
                            shipmentSlaDueAt: event.currentTarget.value,
                            shipmentSlaMode: '',
                        }))
                    }
                />
                <SelectField
                    label="Relative SLA"
                    value={board.bulkDraft().shipmentSlaMode}
                    options={[
                        { name: 'No preset', value: '' },
                        { name: '+2h', value: 'plus_2h' },
                        { name: '+4h', value: 'plus_4h' },
                        { name: 'End of day', value: 'end_of_day' },
                    ]}
                    onChange={(event) => board.applyRelativeShipmentSla(event.currentTarget.value as ShipmentSlaMode)}
                />
                <SelectField
                    label="Bulk settlement status"
                    value={board.bulkDraft().settlementStatus}
                    options={[{ name: 'No change', value: '' }, ...settlementOptions]}
                    onChange={(event) =>
                        board.setBulkDraft((current) => ({
                            ...current,
                            settlementStatus: event.currentTarget.value,
                        }))
                    }
                />
            </div>
            <div class="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
                <div class="grid gap-4 md:grid-cols-[0.8fr_1.2fr]">
                    <InputField
                        label="Save bulk template"
                        value={board.bulkTemplateName()}
                        placeholder="Carrier claim follow-up"
                        onInput={(event) => board.setBulkTemplateName(event.currentTarget.value)}
                    />
                    <div class="space-y-2">
                        <p class="text-sm font-medium text-gray-700">Saved bulk templates</p>
                        <div class="flex flex-wrap gap-2">
                            <Show
                                when={board.savedBulkTemplates().length > 0}
                                fallback={
                                    <p class="text-sm text-gray-500">No saved bulk templates for this store yet.</p>
                                }
                            >
                                <For each={board.savedBulkTemplates()}>
                                    {(template) => (
                                        <div class="flex items-center gap-2 rounded-full border border-gray-200 bg-white px-2 py-1">
                                            <button
                                                type="button"
                                                class="text-sm font-medium text-gray-700"
                                                onClick={() => board.applyBulkTemplate(template)}
                                            >
                                                {template.name}
                                            </button>
                                            <button
                                                type="button"
                                                class="text-xs font-semibold text-red-600"
                                                onClick={() => board.deleteBulkTemplate(template.name)}
                                            >
                                                remove
                                            </button>
                                        </div>
                                    )}
                                </For>
                            </Show>
                        </div>
                    </div>
                </div>
                <div class="mt-4">
                    <Button type="button" size="xs" color="green" onClick={board.saveBulkTemplate}>
                        Save current bulk template
                    </Button>
                </div>
            </div>
            <div class="mt-4">
                <Button
                    type="button"
                    size="sm"
                    color="dark"
                    onClick={() => {
                        void board.applyBulkUpdate()
                    }}
                >
                    Apply bulk update
                </Button>
            </div>
        </div>
    )
}
