import { For, Show } from 'solid-js';
import { Badge, Button } from '../../../components/common/Primitives';
import { useTenantOrdersInsights } from './context';
import { formatActivityActor, formatActivityTime } from './utils';

function activityColor(type: string) {
  switch (type) {
    case 'shipment_note':
      return 'indigo';
    case 'settlement_note':
      return 'green';
    case 'issue_note':
      return 'red';
    default:
      return 'dark';
  }
}

export function StoreActivityFeedPanel() {
  const insights = useTenantOrdersInsights();
  const workspaceURL = (path: string) => {
    const params = new URLSearchParams();
    const storeId = insights.storeId().trim();
    if (storeId) params.set('storeId', storeId);
    const query = params.toString();
    return `/t/${insights.tenantId}${path}${query ? `?${query}` : ''}`;
  };

  return (
    <div class="rounded-lg border border-slate-200 bg-slate-50 p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p class="text-sm font-semibold text-slate-900">
            Store activity feed
          </p>
          <p class="text-sm text-slate-500">
            Latest activity across the current queue slice for store{' '}
            {insights.storeLabel()}.
          </p>
        </div>
        <Button
          type="button"
          size="xs"
          color="light"
          onClick={() => {
            void insights.copyStoreActivityFeed();
          }}
        >
          Copy feed
        </Button>
        <Button
          type="button"
          size="xs"
          color="alternative"
          href={workspaceURL('/orders/audit')}
        >
          Open full audit
        </Button>
        <Button
          type="button"
          size="xs"
          color="alternative"
          href={workspaceURL('/orders/finance')}
        >
          Open finance view
        </Button>
      </div>
      <div class="mt-4 space-y-3">
        <Show
          when={insights.storeActivityFeed().length > 0}
          fallback={
            <div class="rounded-md border border-dashed border-slate-200 bg-white p-3 text-sm text-slate-500">
              No store activity matches the current queue and activity filters.
            </div>
          }
        >
          <For each={insights.storeActivityFeed()}>
            {(entry) => (
              <div class="rounded-md border border-slate-200 bg-white p-3">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge
                      content={entry.activity.type.replaceAll('_', ' ')}
                      color={activityColor(entry.activity.type)}
                    />
                    <p class="text-xs font-semibold text-slate-700">
                      {entry.orderId}
                    </p>
                    <p class="text-xs text-slate-500">{entry.productTitle}</p>
                  </div>
                  <p class="text-xs text-slate-500">
                    {formatActivityTime(entry.activity.createdAt)}
                  </p>
                </div>
                <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                  <span>{formatActivityActor(entry.activity.actor)}</span>
                  <span>owner {entry.operatorAssignee || 'unassigned'}</span>
                </div>
                <p class="mt-2 text-sm text-slate-700">
                  {entry.activity.message}
                </p>
                <Show when={entry.activity.details.length}>
                  <div class="mt-2 flex flex-wrap gap-2">
                    <For each={entry.activity.details}>
                      {(detail) => (
                        <span class="rounded-full bg-slate-50 px-2 py-1 text-[11px] font-medium text-slate-600 ring-1 ring-slate-200">
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
