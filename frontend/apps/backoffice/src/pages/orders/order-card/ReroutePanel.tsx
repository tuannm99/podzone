import { Show } from 'solid-js'
import type { RoutedOrder } from '@podzone/shared/services/orders'
import { Button, InputField } from '@podzone/shared/ui/components/common/Primitives'
import type { OrderCardActions } from './types'

type ReroutePanelProps = {
    order: RoutedOrder
    actions: OrderCardActions
}

export function ReroutePanel(props: ReroutePanelProps) {
    return (
        <>
            <Show when={props.order.exceptionType}>
                <div class="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3">
                    <p class="text-xs font-semibold uppercase tracking-[0.16em] text-amber-700">Exception</p>
                    <p class="mt-2 text-sm text-amber-900">
                        {props.order.exceptionType.replaceAll('_', ' ')} · {props.order.exceptionStatus || 'draft'}
                    </p>
                </div>
            </Show>

            <Show when={props.order.status === 'routing_blocked'}>
                <div class="mt-3 rounded-md border border-rose-200 bg-rose-50 p-3">
                    <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">Manual reroute</p>
                    <p class="mt-2 text-sm text-rose-900">
                        Pick an eligible partner to clear the routing block and move the order back into the queued
                        lane.
                    </p>
                    <div class="mt-3 flex flex-col gap-3 md:flex-row md:items-end">
                        <div class="flex-1">
                            <InputField
                                label="Preferred partner"
                                value={props.actions.rerouteDraftFor(props.order).preferredPartner}
                                placeholder="partner code or name"
                                onInput={(event) =>
                                    props.actions.patchRerouteDraft(props.order.id, {
                                        preferredPartner: event.currentTarget.value,
                                    })
                                }
                            />
                        </div>
                        <Button
                            type="button"
                            color="red"
                            onClick={() => props.actions.rerouteBlockedOrder(props.order)}
                        >
                            Force reroute
                        </Button>
                    </div>
                </div>
            </Show>
        </>
    )
}
