import { Show } from 'solid-js'
import type { RoutedOrder } from '@podzone/shared/services/orders'
import { Badge, Button, InputField, SelectField, TextareaField } from '@podzone/shared/ui/components/common/Primitives'
import type { OrderCardActions, OrderCardUi } from './types'

type IssueHandlingPanelProps = {
    order: RoutedOrder
    actions: OrderCardActions
    ui: OrderCardUi
}

export function IssueHandlingPanel(props: IssueHandlingPanelProps) {
    return (
        <Show when={props.order.exceptionType || props.order.shipmentStatus === 'delivery_issue'}>
            <div class="mt-3 rounded-md border border-rose-200 bg-rose-50 p-3">
                <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">Issue cost handling</p>
                <div class="mt-3 grid gap-4 md:grid-cols-2">
                    <InputField
                        label="Issue cost"
                        value={props.actions.issueDraftFor(props.order).issueCost}
                        placeholder="$6.00"
                        onInput={(event) =>
                            props.actions.patchIssueDraft(props.order.id, {
                                issueCost: event.currentTarget.value,
                            })
                        }
                    />
                    <SelectField
                        label="Resolution path"
                        value={props.actions.issueDraftFor(props.order).issueResolution}
                        options={props.ui.issueResolutionOptions}
                        onChange={(event) =>
                            props.actions.patchIssueDraft(props.order.id, {
                                issueResolution: event.currentTarget.value,
                            })
                        }
                    />
                </div>
                <div class="mt-4">
                    <TextareaField
                        label="Issue notes"
                        value={props.actions.issueDraftFor(props.order).issueNotes}
                        rows={3}
                        onInput={(event) =>
                            props.actions.patchIssueDraft(props.order.id, {
                                issueNotes: event.currentTarget.value,
                            })
                        }
                    />
                </div>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                    <Button
                        type="button"
                        size="xs"
                        color="red"
                        onClick={() => props.actions.saveIssueHandling(props.order)}
                    >
                        Save issue handling
                    </Button>
                    <Badge content={`cost ${props.order.issueCost}`} color="red" />
                    <Badge content={props.order.issueResolution.replaceAll('_', ' ')} color="yellow" />
                </div>
            </div>
        </Show>
    )
}
