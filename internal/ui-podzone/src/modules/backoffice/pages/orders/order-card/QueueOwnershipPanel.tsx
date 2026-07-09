import { Show } from 'solid-js'
import type { RoutedOrder } from '@/services/orders'
import { Badge, Button, InputField } from '@/solid/components/common/Primitives'
import type { OrderCardActions, OrderCardHelpers } from './types'

type QueueOwnershipPanelProps = {
    order: RoutedOrder
    actions: OrderCardActions
    helpers: OrderCardHelpers
}

export function QueueOwnershipPanel(props: QueueOwnershipPanelProps) {
    return (
        <div class="mt-3 rounded-md border border-sky-200 bg-sky-50 p-3">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-sky-700">Queue ownership</p>
            <div class="mt-3 grid gap-4 md:grid-cols-2">
                <InputField
                    label="Operator assignee"
                    value={props.actions.queueDraftFor(props.order).operatorAssignee}
                    placeholder="linh.nguyen"
                    onInput={(event) =>
                        props.actions.patchQueueDraft(props.order.id, {
                            operatorAssignee: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Shipment SLA due"
                    type="datetime-local"
                    value={props.actions.queueDraftFor(props.order).shipmentSlaDueAt}
                    onInput={(event) =>
                        props.actions.patchQueueDraft(props.order.id, {
                            shipmentSlaDueAt: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Issue SLA due"
                    type="datetime-local"
                    value={props.actions.queueDraftFor(props.order).issueSlaDueAt}
                    onInput={(event) =>
                        props.actions.patchQueueDraft(props.order.id, {
                            issueSlaDueAt: event.currentTarget.value,
                        })
                    }
                />
            </div>
            <div class="mt-3 flex flex-wrap items-center gap-2">
                <Button
                    type="button"
                    size="xs"
                    color="primary"
                    onClick={() => props.actions.saveQueueControl(props.order)}
                >
                    Save queue control
                </Button>
                <Badge content={`owner ${props.order.operatorAssignee || 'unassigned'}`} color="indigo" />
                <Show when={props.order.shipmentSlaDueAt}>
                    <Badge
                        content={`shipment SLA ${props.helpers.isOverdue(props.order.shipmentSlaDueAt) ? 'overdue' : 'set'}`}
                        color={props.helpers.isOverdue(props.order.shipmentSlaDueAt) ? 'red' : 'blue'}
                    />
                </Show>
                <Show when={props.order.issueSlaDueAt}>
                    <Badge
                        content={`issue SLA ${props.helpers.isOverdue(props.order.issueSlaDueAt) ? 'overdue' : 'set'}`}
                        color={props.helpers.isOverdue(props.order.issueSlaDueAt) ? 'red' : 'blue'}
                    />
                </Show>
            </div>
        </div>
    )
}
