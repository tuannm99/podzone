import { Show } from 'solid-js';
import type { RoutedOrder } from '@/services/orders';
import {
  Badge,
  Button,
  InputField,
} from '@/solid/components/common/Primitives';
import type { OrderCardActions, OrderCardHelpers } from './types';

type QueueOwnershipPanelProps = {
  order: RoutedOrder;
  actions: OrderCardActions;
  helpers: OrderCardHelpers;
};

export function QueueOwnershipPanel(props: QueueOwnershipPanelProps) {
  const { order, actions, helpers } = props;

  return (
    <div class="mt-3 rounded-md border border-sky-200 bg-sky-50 p-3">
      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-sky-700">
        Queue ownership
      </p>
      <div class="mt-3 grid gap-4 md:grid-cols-2">
        <InputField
          label="Operator assignee"
          value={actions.queueDraftFor(order).operatorAssignee}
          placeholder="linh.nguyen"
          onInput={(event) =>
            actions.patchQueueDraft(order.id, {
              operatorAssignee: event.currentTarget.value,
            })
          }
        />
        <InputField
          label="Shipment SLA due"
          type="datetime-local"
          value={actions.queueDraftFor(order).shipmentSlaDueAt}
          onInput={(event) =>
            actions.patchQueueDraft(order.id, {
              shipmentSlaDueAt: event.currentTarget.value,
            })
          }
        />
        <InputField
          label="Issue SLA due"
          type="datetime-local"
          value={actions.queueDraftFor(order).issueSlaDueAt}
          onInput={(event) =>
            actions.patchQueueDraft(order.id, {
              issueSlaDueAt: event.currentTarget.value,
            })
          }
        />
      </div>
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <Button
          type="button"
          size="xs"
          color="blue"
          onClick={() => actions.saveQueueControl(order)}
        >
          Save queue control
        </Button>
        <Badge
          content={`owner ${order.operatorAssignee || 'unassigned'}`}
          color="indigo"
        />
        <Show when={order.shipmentSlaDueAt}>
          <Badge
            content={`shipment SLA ${helpers.isOverdue(order.shipmentSlaDueAt) ? 'overdue' : 'set'}`}
            color={helpers.isOverdue(order.shipmentSlaDueAt) ? 'red' : 'blue'}
          />
        </Show>
        <Show when={order.issueSlaDueAt}>
          <Badge
            content={`issue SLA ${helpers.isOverdue(order.issueSlaDueAt) ? 'overdue' : 'set'}`}
            color={helpers.isOverdue(order.issueSlaDueAt) ? 'red' : 'blue'}
          />
        </Show>
      </div>
    </div>
  );
}
