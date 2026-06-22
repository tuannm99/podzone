import { Show } from 'solid-js';
import type { RoutedOrder } from '@/services/orders';
import {
  Badge,
  Button,
  InputField,
  SelectField,
  TextareaField,
} from '@/solid/components/common/Primitives';
import type { OrderCardActions, OrderCardUi } from './types';

type IssueHandlingPanelProps = {
  order: RoutedOrder;
  actions: OrderCardActions;
  ui: OrderCardUi;
};

export function IssueHandlingPanel(props: IssueHandlingPanelProps) {
  const { order, actions, ui } = props;

  return (
    <Show
      when={order.exceptionType || order.shipmentStatus === 'delivery_issue'}
    >
      <div class="mt-3 rounded-md border border-rose-200 bg-rose-50 p-3">
        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">
          Issue cost handling
        </p>
        <div class="mt-3 grid gap-4 md:grid-cols-2">
          <InputField
            label="Issue cost"
            value={actions.issueDraftFor(order).issueCost}
            placeholder="$6.00"
            onInput={(event) =>
              actions.patchIssueDraft(order.id, {
                issueCost: event.currentTarget.value,
              })
            }
          />
          <SelectField
            label="Resolution path"
            value={actions.issueDraftFor(order).issueResolution}
            options={ui.issueResolutionOptions}
            onChange={(event) =>
              actions.patchIssueDraft(order.id, {
                issueResolution: event.currentTarget.value,
              })
            }
          />
        </div>
        <div class="mt-4">
          <TextareaField
            label="Issue notes"
            value={actions.issueDraftFor(order).issueNotes}
            rows={3}
            onInput={(event) =>
              actions.patchIssueDraft(order.id, {
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
            onClick={() => actions.saveIssueHandling(order)}
          >
            Save issue handling
          </Button>
          <Badge content={`cost ${order.issueCost}`} color="red" />
          <Badge
            content={order.issueResolution.replaceAll('_', ' ')}
            color="yellow"
          />
        </div>
      </div>
    </Show>
  );
}
