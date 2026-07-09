import { For } from 'solid-js'
import type { RoutedOrder } from '@/services/orders'
import { ActivityLogPanel } from './order-card/ActivityLogPanel'
import { Header } from './order-card/Header'
import { IssueHandlingPanel } from './order-card/IssueHandlingPanel'
import { QueueOwnershipPanel } from './order-card/QueueOwnershipPanel'
import { ReroutePanel } from './order-card/ReroutePanel'
import { SettlementPanel } from './order-card/SettlementPanel'
import { ShipmentPanel } from './order-card/ShipmentPanel'
import type { OrderCardActions, OrderCardHelpers, OrderCardUi } from './order-card/types'

type OrderCardProps = {
    order: RoutedOrder
    selected: boolean
    actions: OrderCardActions
    helpers: OrderCardHelpers
    ui: OrderCardUi
}

export function OrderCard(props: OrderCardProps) {
    return (
        <div class="rounded-lg border border-gray-200 bg-white p-4">
            <Header order={props.order} selected={props.selected} actions={props.actions} helpers={props.helpers} />
            <ReroutePanel order={props.order} actions={props.actions} />
            <QueueOwnershipPanel order={props.order} actions={props.actions} helpers={props.helpers} />
            <IssueHandlingPanel order={props.order} actions={props.actions} ui={props.ui} />
            <SettlementPanel order={props.order} actions={props.actions} helpers={props.helpers} ui={props.ui} />
            <ShipmentPanel order={props.order} actions={props.actions} ui={props.ui} />

            <div class="mt-3 rounded-md bg-gray-50 p-3">
                <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">Timeline</p>
                <div class="mt-2 space-y-1 text-sm text-gray-600">
                    <For each={props.order.timeline}>{(entry) => <p>{entry}</p>}</For>
                </div>
            </div>

            <ActivityLogPanel order={props.order} actions={props.actions} helpers={props.helpers} ui={props.ui} />
        </div>
    )
}
