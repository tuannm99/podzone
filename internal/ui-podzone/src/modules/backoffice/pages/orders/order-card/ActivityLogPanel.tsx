import { For, Show } from 'solid-js';
import type { RoutedOrder } from '@/services/orders';
import { Badge, Button, SelectField } from '@/solid/components/common/Primitives';
import { formatActivityActor, formatActivityTime } from '../utils';
import type { ActivityFilter, OrderCardActions, OrderCardHelpers, OrderCardUi } from './types';

type ActivityLogPanelProps = {
  order: RoutedOrder;
  actions: OrderCardActions;
  helpers: OrderCardHelpers;
  ui: OrderCardUi;
};

export function ActivityLogPanel(props: ActivityLogPanelProps) {
  const { order, actions, helpers, ui } = props;

  return (
    <div class="mt-3 rounded-md border border-slate-200 bg-white p-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
          Activity log
        </p>
        <div class="flex flex-wrap items-center gap-2">
          <div class="min-w-[11rem]">
            <SelectField
              label=""
              value={ui.activityFilter}
              options={ui.activityFilterOptions}
              onChange={(event) =>
                ui.setActivityFilter(event.currentTarget.value as ActivityFilter)
              }
            />
          </div>
          <Show when={ui.activityFilter === 'all'}>
            <Button
              type="button"
              size="xs"
              color={ui.hideSystemActivity ? 'dark' : 'light'}
              onClick={ui.toggleHideSystemActivity}
            >
              {ui.hideSystemActivity ? 'Show system' : 'Hide system'}
            </Button>
          </Show>
          <Button
            type="button"
            size="xs"
            color="light"
            onClick={() => {
              void actions.copyActivitySummary(order);
            }}
          >
            Copy summary
          </Button>
        </div>
      </div>
      <div class="mt-3 space-y-3">
        <Show
          when={helpers.filteredActivityLogFor(order).length > 0}
          fallback={
            <div class="rounded-md border border-dashed border-slate-200 bg-slate-50 p-3 text-sm text-slate-500">
              <Show
                when={helpers.hiddenSystemActivityCountFor(order) > 0}
                fallback={'No activity matches the current filter.'}
              >
                {helpers.hiddenSystemActivityCountFor(order)} system updates
                are hidden.
              </Show>
            </div>
          }
        >
          <For each={helpers.filteredActivityLogFor(order)}>
            {(activity) => (
              <div class="rounded-md border border-slate-200 bg-slate-50 p-3">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge
                      content={activity.type.replaceAll('_', ' ')}
                      color={helpers.activityColor(activity.type)}
                    />
                    <p class="text-xs font-medium text-slate-500">
                      {formatActivityActor(activity.actor)}
                    </p>
                  </div>
                  <p class="text-xs text-slate-500">
                    {formatActivityTime(activity.createdAt)}
                  </p>
                </div>
                <p class="mt-2 text-sm text-slate-700">{activity.message}</p>
                <Show when={activity.details.length}>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <For each={activity.details}>
                      {(detail) => (
                        <span class="rounded-full bg-white px-2 py-1 text-[11px] font-medium text-slate-600 ring-1 ring-slate-200">
                          {detail.key.replaceAll('_', ' ')}: {detail.value}
                        </span>
                      )}
                    </For>
                  </div>
                </Show>
              </div>
            )}
          </For>
        </Show>
      </div>
    </div>
  );
}
