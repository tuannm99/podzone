import { Show } from 'solid-js';
import type { RoutedOrder } from '@/services/orders';
import { Badge, Button } from '@/solid/components/common/Primitives';
import type { OrderCardActions, OrderCardHelpers } from './types';

type HeaderProps = {
  order: RoutedOrder;
  selected: boolean;
  actions: OrderCardActions;
  helpers: OrderCardHelpers;
};

export function Header(props: HeaderProps) {
  const { order, actions, helpers } = props;

  return (
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-start gap-3">
        <label class="mt-1">
          <input
            type="checkbox"
            checked={props.selected}
            onChange={(event) =>
              actions.toggleSelected(event.currentTarget.checked)
            }
          />
        </label>
        <div>
          <p class="text-base font-semibold text-gray-900">{order.id}</p>
          <p class="mt-1 text-sm text-gray-500">
            {order.productTitle} · {order.partner || 'partner pending'}
          </p>
          <p class="mt-1 text-sm text-gray-500">
            customer {order.customerName} · qty {order.quantity} · total{' '}
            {order.total}
          </p>
          <p class="mt-1 text-sm text-gray-500">
            owner {order.operatorAssignee || 'unassigned'}
          </p>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <Show when={helpers.queueSort === 'priority'}>
          <Badge
            content={`priority ${helpers.priorityScoreFor(order) + 1}`}
            color={helpers.priorityScoreFor(order) < 3 ? 'red' : 'dark'}
          />
        </Show>
        <Badge
          content={order.status.replaceAll('_', ' ')}
          color={helpers.statusColor(order.status)}
        />
        <Show when={order.exceptionStatus}>
          <Badge
            content={`${order.exceptionStatus} issue`}
            color={helpers.exceptionColor(order.exceptionStatus)}
          />
        </Show>
        <Show when={order.routingBlockCode}>
          <Badge
            content={`blocked ${order.routingBlockCode.replaceAll('_', ' ')}`}
            color="red"
          />
        </Show>
        <Badge
          content={order.shipmentStatus.replaceAll('_', ' ')}
          color={helpers.shipmentColor(order.shipmentStatus)}
        />
        <Badge
          content={order.settlementStatus.replaceAll('_', ' ')}
          color={helpers.settlementColor(order.settlementStatus)}
        />
        <Button
          type="button"
          size="xs"
          color="green"
          disabled={
            order.status === 'routing_blocked' ||
            order.status === 'shipped' ||
            order.exceptionStatus === 'open' ||
            order.exceptionStatus === 'escalated'
          }
          onClick={() => actions.advanceOrder(order.id)}
        >
          Advance route
        </Button>
        <Button
          type="button"
          size="xs"
          color="alternative"
          disabled={
            order.exceptionStatus === 'open' ||
            order.exceptionStatus === 'resolved'
          }
          onClick={() => actions.raiseException(order.id)}
        >
          Raise issue
        </Button>
        <Show when={order.routingBlockReason}>
          <p class="w-full text-sm text-rose-700">
            Routing blocked: {order.routingBlockReason}
          </p>
        </Show>
        <Show when={order.exceptionStatus === 'open'}>
          <Button
            type="button"
            size="xs"
            color="blue"
            onClick={() => actions.updateExceptionStatus(order.id, 'escalated')}
          >
            Escalate
          </Button>
          <Button
            type="button"
            size="xs"
            color="light"
            onClick={() => actions.updateExceptionStatus(order.id, 'resolved')}
          >
            Resolve
          </Button>
        </Show>
      </div>
    </div>
  );
}
