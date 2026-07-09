import type { RoutedOrder } from '@/services/orders'
import { Badge, Button, InputField, SelectField, TextareaField } from '@/solid/components/common/Primitives'
import type { OrderCardActions, OrderCardHelpers, OrderCardUi } from './types'

type SettlementPanelProps = {
    order: RoutedOrder
    actions: OrderCardActions
    helpers: OrderCardHelpers
    ui: OrderCardUi
}

export function SettlementPanel(props: SettlementPanelProps) {
    return (
        <div class="mt-3 rounded-md border border-emerald-200 bg-emerald-50 p-3">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">Settlement readiness</p>
            <div class="mt-3 grid gap-3 md:grid-cols-2">
                <div class="rounded-md border border-emerald-200 bg-white p-3">
                    <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">Base cost snapshot</p>
                    <p class="mt-2 text-sm font-semibold text-gray-900">{props.order.baseCostSnapshot}</p>
                </div>
                <div class="rounded-md border border-emerald-200 bg-white p-3">
                    <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">Realized margin</p>
                    <p class="mt-2 text-sm font-semibold text-gray-900">{props.order.realizedMargin}</p>
                </div>
            </div>
            <div class="mt-3 grid gap-4 md:grid-cols-2">
                <InputField
                    label="Fulfillment cost"
                    value={props.actions.settlementDraftFor(props.order).fulfillmentCost}
                    placeholder="$9.50"
                    onInput={(event) =>
                        props.actions.patchSettlementDraft(props.order.id, {
                            fulfillmentCost: event.currentTarget.value,
                        })
                    }
                />
                <InputField
                    label="Shipping cost"
                    value={props.actions.settlementDraftFor(props.order).shippingCost}
                    placeholder="$4.25"
                    onInput={(event) =>
                        props.actions.patchSettlementDraft(props.order.id, {
                            shippingCost: event.currentTarget.value,
                        })
                    }
                />
                <SelectField
                    label="Settlement status"
                    value={props.actions.settlementDraftFor(props.order).settlementStatus}
                    options={props.ui.settlementOptions}
                    onChange={(event) =>
                        props.actions.patchSettlementDraft(props.order.id, {
                            settlementStatus: event.currentTarget.value,
                        })
                    }
                />
            </div>
            <div class="mt-4">
                <TextareaField
                    label="Settlement notes"
                    value={props.actions.settlementDraftFor(props.order).settlementNotes}
                    rows={3}
                    onInput={(event) =>
                        props.actions.patchSettlementDraft(props.order.id, {
                            settlementNotes: event.currentTarget.value,
                        })
                    }
                />
            </div>
            <div class="mt-3 flex flex-wrap items-center gap-2">
                <Button type="button" size="xs" color="green" onClick={() => props.actions.saveSettlement(props.order)}>
                    Save settlement state
                </Button>
                <Badge content={`margin ${props.order.realizedMargin}`} color="green" />
                <Badge
                    content={`settlement ${props.order.settlementStatus.replaceAll('_', ' ')}`}
                    color={props.helpers.settlementColor(props.order.settlementStatus)}
                />
            </div>
        </div>
    )
}
