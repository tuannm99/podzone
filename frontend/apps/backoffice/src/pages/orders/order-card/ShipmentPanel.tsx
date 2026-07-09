import { Show } from 'solid-js'
import type { RoutedOrder } from '@podzone/shared/services/orders'
import { Badge, Button, InputField, SelectField, TextareaField } from '@podzone/shared/ui/components/common/Primitives'
import type { OrderCardActions, OrderCardUi } from './types'

type ShipmentPanelProps = {
    order: RoutedOrder
    actions: OrderCardActions
    ui: OrderCardUi
}

export function ShipmentPanel(props: ShipmentPanelProps) {
    return (
        <div class="mt-3 rounded-md border border-slate-200 bg-slate-50 p-3">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">Manual shipment control</p>
            <div class="mt-3 grid gap-4 md:grid-cols-2">
                <SelectField
                    label="Shipment status"
                    value={props.actions.shipmentDraftFor(props.order).shipmentStatus}
                    options={props.ui.shipmentOptions}
                    onChange={(event) =>
                        props.actions.patchShipmentDraft(props.order.id, {
                            shipmentStatus: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Carrier"
                    value={props.actions.shipmentDraftFor(props.order).shipmentCarrier}
                    placeholder="UPS"
                    onInput={(event) =>
                        props.actions.patchShipmentDraft(props.order.id, {
                            shipmentCarrier: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Tracking number"
                    value={props.actions.shipmentDraftFor(props.order).shipmentTrackingNumber}
                    placeholder="1Z999..."
                    onInput={(event) =>
                        props.actions.patchShipmentDraft(props.order.id, {
                            shipmentTrackingNumber: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Tracking URL"
                    value={props.actions.shipmentDraftFor(props.order).shipmentTrackingUrl}
                    placeholder="https://tracking.example/..."
                    onInput={(event) =>
                        props.actions.patchShipmentDraft(props.order.id, {
                            shipmentTrackingUrl: event.currentTarget.value,
                        })
                    }
                />
            </div>
            <div class="mt-4">
                <TextareaField
                    label="Shipment notes"
                    value={props.actions.shipmentDraftFor(props.order).shipmentNotes}
                    rows={3}
                    onInput={(event) =>
                        props.actions.patchShipmentDraft(props.order.id, {
                            shipmentNotes: event.currentTarget.value,
                        })
                    }
                />
            </div>
            <div class="mt-3 flex flex-wrap items-center gap-2">
                <Button type="button" size="xs" color="primary" onClick={() => props.actions.saveShipment(props.order)}>
                    Save shipment state
                </Button>
                <Show when={props.order.shipmentCarrier || props.order.shipmentTrackingNumber}>
                    <Badge
                        content={`${props.order.shipmentCarrier || 'manual carrier'} ${props.order.shipmentTrackingNumber || ''}`.trim()}
                        color="indigo"
                    />
                </Show>
                <Show when={props.order.deliveredAt}>
                    <Badge content="Delivered confirmed" color="green" />
                </Show>
            </div>
        </div>
    )
}
