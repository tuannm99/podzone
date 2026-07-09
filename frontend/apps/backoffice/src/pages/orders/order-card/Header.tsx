import { Show } from 'solid-js'
import type { RoutedOrder } from '@/services/orders'
import { Badge, Button } from '@/solid/components/common/Primitives'
import type { OrderCardActions, OrderCardHelpers } from './types'

type HeaderProps = {
    order: RoutedOrder
    selected: boolean
    actions: OrderCardActions
    helpers: OrderCardHelpers
}

export function Header(props: HeaderProps) {
    return (
        <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-start gap-3">
                <label class="mt-1">
                    <input
                        type="checkbox"
                        checked={props.selected}
                        onChange={(event) => props.actions.toggleSelected(event.currentTarget.checked)}
                    />
                </label>
                <div>
                    <p class="text-base font-semibold text-gray-900">{props.order.id}</p>
                    <p class="mt-1 text-sm text-gray-500">
                        {props.order.productTitle} · {props.order.partner || 'partner pending'}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                        customer {props.order.customerName} · qty {props.order.quantity} · total {props.order.total}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">owner {props.order.operatorAssignee || 'unassigned'}</p>
                </div>
            </div>
            <div class="flex flex-wrap items-center gap-2">
                <Show when={props.helpers.queueSort === 'priority'}>
                    <Badge
                        content={`priority ${props.helpers.priorityScoreFor(props.order) + 1}`}
                        color={props.helpers.priorityScoreFor(props.order) < 3 ? 'red' : 'dark'}
                    />
                </Show>
                <Badge
                    content={props.order.status.replaceAll('_', ' ')}
                    color={props.helpers.statusColor(props.order.status)}
                />
                <Show when={props.order.exceptionStatus}>
                    <Badge
                        content={`${props.order.exceptionStatus} issue`}
                        color={props.helpers.exceptionColor(props.order.exceptionStatus)}
                    />
                </Show>
                <Show when={props.order.routingBlockCode}>
                    <Badge content={`blocked ${props.order.routingBlockCode.replaceAll('_', ' ')}`} color="red" />
                </Show>
                <Badge
                    content={props.order.shipmentStatus.replaceAll('_', ' ')}
                    color={props.helpers.shipmentColor(props.order.shipmentStatus)}
                />
                <Badge
                    content={props.order.settlementStatus.replaceAll('_', ' ')}
                    color={props.helpers.settlementColor(props.order.settlementStatus)}
                />
                <Button
                    type="button"
                    size="xs"
                    color="green"
                    disabled={
                        props.order.status === 'routing_blocked' ||
                        props.order.status === 'shipped' ||
                        props.order.exceptionStatus === 'open' ||
                        props.order.exceptionStatus === 'escalated'
                    }
                    onClick={() => props.actions.advanceOrder(props.order.id)}
                >
                    Advance route
                </Button>
                <Button
                    type="button"
                    size="xs"
                    color="alternative"
                    disabled={props.order.exceptionStatus === 'open' || props.order.exceptionStatus === 'resolved'}
                    onClick={() => props.actions.raiseException(props.order.id)}
                >
                    Raise issue
                </Button>
                <Show when={props.order.routingBlockReason}>
                    <p class="w-full text-sm text-rose-700">Routing blocked: {props.order.routingBlockReason}</p>
                </Show>
                <Show when={props.order.exceptionStatus === 'open'}>
                    <Button
                        type="button"
                        size="xs"
                        color="primary"
                        onClick={() => props.actions.updateExceptionStatus(props.order.id, 'escalated')}
                    >
                        Escalate
                    </Button>
                    <Button
                        type="button"
                        size="xs"
                        color="light"
                        onClick={() => props.actions.updateExceptionStatus(props.order.id, 'resolved')}
                    >
                        Resolve
                    </Button>
                </Show>
            </div>
        </div>
    )
}
