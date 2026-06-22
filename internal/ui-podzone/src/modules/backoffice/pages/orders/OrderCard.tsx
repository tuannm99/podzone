import { For } from 'solid-js';
import type { RoutedOrder } from '@/services/orders';
import { ActivityLogPanel } from './order-card/ActivityLogPanel';
import { Header } from './order-card/Header';
import { IssueHandlingPanel } from './order-card/IssueHandlingPanel';
import { QueueOwnershipPanel } from './order-card/QueueOwnershipPanel';
import { ReroutePanel } from './order-card/ReroutePanel';
import { SettlementPanel } from './order-card/SettlementPanel';
import { ShipmentPanel } from './order-card/ShipmentPanel';
import type {
  OrderCardActions,
  OrderCardHelpers,
  OrderCardUi,
} from './order-card/types';

type OrderCardProps = {
  order: RoutedOrder;
  selected: boolean;
  actions: OrderCardActions;
  helpers: OrderCardHelpers;
  ui: OrderCardUi;
};

export function OrderCard(props: OrderCardProps) {
  const { order, actions, helpers, ui } = props;

  return (
    <div class="rounded-lg border border-gray-200 bg-white p-4">
      <Header
        order={order}
        selected={props.selected}
        actions={actions}
        helpers={helpers}
      />
      <ReroutePanel order={order} actions={actions} />
      <QueueOwnershipPanel
        order={order}
        actions={actions}
        helpers={helpers}
      />
      <IssueHandlingPanel order={order} actions={actions} ui={ui} />
      <SettlementPanel
        order={order}
        actions={actions}
        helpers={helpers}
        ui={ui}
      />
      <ShipmentPanel order={order} actions={actions} ui={ui} />

      <div class="mt-3 rounded-md bg-gray-50 p-3">
        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
          Timeline
        </p>
        <div class="mt-2 space-y-1 text-sm text-gray-600">
          <For each={order.timeline}>{(entry) => <p>{entry}</p>}</For>
        </div>
      </div>

      <ActivityLogPanel
        order={order}
        actions={actions}
        helpers={helpers}
        ui={ui}
      />
    </div>
  );
}
