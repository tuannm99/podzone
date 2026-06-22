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

type ShipmentPanelProps = {
  order: RoutedOrder;
  actions: OrderCardActions;
  ui: OrderCardUi;
};

export function ShipmentPanel(props: ShipmentPanelProps) {
  const { order, actions, ui } = props;

  return (
    <div class="mt-3 rounded-md border border-slate-200 bg-slate-50 p-3">
      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
        Manual shipment control
      </p>
      <div class="mt-3 grid gap-4 md:grid-cols-2">
        <SelectField
          label="Shipment status"
          value={actions.shipmentDraftFor(order).shipmentStatus}
          options={ui.shipmentOptions}
          onChange={(event) =>
            actions.patchShipmentDraft(order.id, {
              shipmentStatus: event.currentTarget.value,
            })
          }
        />
        <InputField
          label="Carrier"
          value={actions.shipmentDraftFor(order).shipmentCarrier}
          placeholder="UPS"
          onInput={(event) =>
            actions.patchShipmentDraft(order.id, {
              shipmentCarrier: event.currentTarget.value,
            })
          }
        />
        <InputField
          label="Tracking number"
          value={actions.shipmentDraftFor(order).shipmentTrackingNumber}
          placeholder="1Z999..."
          onInput={(event) =>
            actions.patchShipmentDraft(order.id, {
              shipmentTrackingNumber: event.currentTarget.value,
            })
          }
        />
        <InputField
          label="Tracking URL"
          value={actions.shipmentDraftFor(order).shipmentTrackingUrl}
          placeholder="https://tracking.example/..."
          onInput={(event) =>
            actions.patchShipmentDraft(order.id, {
              shipmentTrackingUrl: event.currentTarget.value,
            })
          }
        />
      </div>
      <div class="mt-4">
        <TextareaField
          label="Shipment notes"
          value={actions.shipmentDraftFor(order).shipmentNotes}
          rows={3}
          onInput={(event) =>
            actions.patchShipmentDraft(order.id, {
              shipmentNotes: event.currentTarget.value,
            })
          }
        />
      </div>
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <Button
          type="button"
          size="xs"
          color="blue"
          onClick={() => actions.saveShipment(order)}
        >
          Save shipment state
        </Button>
        <Show when={order.shipmentCarrier || order.shipmentTrackingNumber}>
          <Badge
            content={`${order.shipmentCarrier || 'manual carrier'} ${order.shipmentTrackingNumber || ''}`.trim()}
            color="indigo"
          />
        </Show>
        <Show when={order.deliveredAt}>
          <Badge content="Delivered confirmed" color="green" />
        </Show>
      </div>
    </div>
  );
}
