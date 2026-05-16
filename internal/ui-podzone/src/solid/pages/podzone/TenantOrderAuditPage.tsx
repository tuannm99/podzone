import { useParams } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal } from 'solid-js';
import {
  getRoutedOrderActivities,
  type RoutedOrderActivityFeedEntry,
} from '../../../services/orders';
import { tenantStorage } from '../../../services/tenantStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const activityFilterOptions = [
  { name: 'All', value: 'all' },
  { name: 'Notes only', value: 'notes' },
  { name: 'System', value: 'system' },
  { name: 'Shipment', value: 'shipment_note' },
  { name: 'Settlement', value: 'settlement_note' },
  { name: 'Issue', value: 'issue_note' },
] as const;

const timeWindowOptions = [
  { name: '24h', value: '24h' },
  { name: '7 days', value: '7d' },
  { name: '30 days', value: '30d' },
  { name: 'All time', value: 'all' },
] as const;

type ActivityFilter = (typeof activityFilterOptions)[number]['value'];
type TimeWindow = (typeof timeWindowOptions)[number]['value'];

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

function formatActivityTime(value: string) {
  return new Date(value).toLocaleString();
}

function formatActivityActor(actor: string) {
  const normalized = actor.trim();
  return normalized || 'system';
}

function resolveSinceIso(window: TimeWindow) {
  const now = Date.now();
  switch (window) {
    case '24h':
      return new Date(now - 24 * 60 * 60 * 1000).toISOString();
    case '7d':
      return new Date(now - 7 * 24 * 60 * 60 * 1000).toISOString();
    case '30d':
      return new Date(now - 30 * 24 * 60 * 60 * 1000).toISOString();
    default:
      return undefined;
  }
}

function formatFeedSummary(
  tenantId: string,
  entries: RoutedOrderActivityFeedEntry[]
) {
  return [
    `Store audit feed for ${tenantId}`,
    '',
    ...entries.map((entry) => {
      const details = entry.activity.details
        .map((detail) => `${detail.key}=${detail.value}`)
        .join(', ');
      return [
        `[${formatActivityTime(entry.activity.createdAt)}]`,
        entry.orderId,
        `(${entry.productTitle})`,
        `[${entry.partner}]`,
        `owner ${entry.operatorAssignee || 'unassigned'}`,
        entry.activity.type,
        `by ${formatActivityActor(entry.activity.actor)}`,
        entry.activity.message,
        details ? `(${details})` : '',
      ]
        .filter(Boolean)
        .join(' ');
    }),
  ].join('\n');
}

export default function TenantOrderAuditPage() {
  const params = useParams({ from: '/t/$tenantId/orders/audit' });

  const [entries, setEntries] = createSignal<RoutedOrderActivityFeedEntry[]>([]);
  const [nextCursor, setNextCursor] = createSignal<string>();
  const [total, setTotal] = createSignal(0);
  const [activityFilter, setActivityFilter] =
    createSignal<ActivityFilter>('notes');
  const [hideSystemActivity, setHideSystemActivity] = createSignal(true);
  const [timeWindow, setTimeWindow] = createSignal<TimeWindow>('7d');
  const [actorFilter, setActorFilter] = createSignal('');
  const [orderFilter, setOrderFilter] = createSignal('');
  const [partnerFilter, setPartnerFilter] = createSignal('');
  const [assigneeFilter, setAssigneeFilter] = createSignal('');
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const loadEntries = async (after?: string, append = false) => {
    const result = await getRoutedOrderActivities({
      activityType: activityFilter(),
      actorContains: actorFilter().trim(),
      orderId: orderFilter().trim(),
      partner: partnerFilter().trim(),
      assignee: assigneeFilter().trim(),
      since: resolveSinceIso(timeWindow()),
      limit: 50,
      after,
      includeSystem:
        activityFilter() === 'all'
          ? !hideSystemActivity()
          : activityFilter() === 'system',
    });
    if (!result.success) {
      setError(result.message);
      setEntries([]);
      setNextCursor(undefined);
      setTotal(0);
      return;
    }
    setEntries((current) =>
      append ? [...current, ...result.data.entries] : result.data.entries
    );
    setNextCursor(result.data.nextCursor);
    setTotal(result.data.total);
  };

  const auditFeed = () => entries();

  const copyFeed = async () => {
    try {
      await navigator.clipboard.writeText(
        formatFeedSummary(params().tenantId, auditFeed())
      );
      setMessage(`Copied audit feed for ${params().tenantId}.`);
    } catch {
      setError('Could not copy audit feed to clipboard.');
    }
  };

  const loadMore = async () => {
    const cursor = nextCursor();
    if (!cursor) {
      return;
    }
    await loadEntries(cursor, true);
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    void loadEntries(undefined, false);
  });

  createEffect(() => {
    activityFilter();
    hideSystemActivity();
    actorFilter();
    orderFilter();
    partnerFilter();
    assigneeFilter();
    timeWindow();
    void loadEntries(undefined, false);
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Store Audit"
          title={`Audit history for store ${params().tenantId}`}
          copy="Review store-wide routed order activity across shipment, settlement, issue handling, and queue ownership updates without opening each order card."
        />
      </Card>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <Card class="space-y-4">
        <SectionTitle
          title="Audit filters"
          subtitle="Focus the store-wide activity feed by type, actor, order, partner, assignee, and recent time window."
        />
        <div class="grid gap-4 md:grid-cols-3">
          <SelectField
            label="Activity type"
            value={activityFilter()}
            options={activityFilterOptions.map((option) => ({
              name: option.name,
              value: option.value,
            }))}
            onChange={(event) =>
              setActivityFilter(event.currentTarget.value as ActivityFilter)
            }
          />
          <SelectField
            label="Time window"
            value={timeWindow()}
            options={timeWindowOptions.map((option) => ({
              name: option.name,
              value: option.value,
            }))}
            onChange={(event) =>
              setTimeWindow(event.currentTarget.value as TimeWindow)
            }
          />
          <InputField
            label="Actor filter"
            value={actorFilter()}
            placeholder="user:12"
            onInput={(event) => setActorFilter(event.currentTarget.value)}
          />
          <InputField
            label="Order filter"
            value={orderFilter()}
            placeholder="ORD-1234ABCD"
            onInput={(event) => setOrderFilter(event.currentTarget.value)}
          />
          <InputField
            label="Partner filter"
            value={partnerFilter()}
            placeholder="Print Partner A"
            onInput={(event) => setPartnerFilter(event.currentTarget.value)}
          />
          <InputField
            label="Assignee filter"
            value={assigneeFilter()}
            placeholder="ops.lead"
            onInput={(event) => setAssigneeFilter(event.currentTarget.value)}
          />
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <Show when={activityFilter() === 'all'}>
            <Button
              type="button"
              size="xs"
              color={hideSystemActivity() ? 'dark' : 'light'}
              onClick={() => setHideSystemActivity((current) => !current)}
            >
              {hideSystemActivity() ? 'Show system' : 'Hide system'}
            </Button>
          </Show>
          <Button
            type="button"
            size="xs"
            color="light"
            href={`/t/${params().tenantId}/orders`}
          >
            Back to orders board
          </Button>
          <Button
            type="button"
            size="xs"
            color="blue"
            onClick={() => {
              void copyFeed();
            }}
          >
            Copy audit feed
          </Button>
        </div>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Audit feed"
          subtitle="Newest matching activity first across all routed orders in this store."
        />
        <Show
          when={auditFeed().length > 0}
          fallback={
            <EmptyBlock
              title="No audit activity matched"
              copy="Try widening the time window, removing the actor filter, or including system activity."
            />
          }
        >
          <div class="flex flex-wrap items-center justify-between gap-2 text-sm text-slate-500">
            <p>
              Showing {auditFeed().length} of {total()} matching activity entries.
            </p>
            <Show when={!!nextCursor()}>
              <Button
                type="button"
                size="xs"
                color="alternative"
                onClick={() => {
                  void loadMore();
                }}
              >
                Load more
              </Button>
            </Show>
          </div>
          <div class="space-y-3">
            <For each={auditFeed()}>
              {(entry) => (
                <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={entry.activity.type.replaceAll('_', ' ')}
                        color={activityColor(entry.activity.type)}
                      />
                      <p class="text-sm font-semibold text-slate-900">
                        {entry.orderId}
                      </p>
                      <p class="text-sm text-slate-500">{entry.productTitle}</p>
                      <p class="text-sm text-slate-500">{entry.partner}</p>
                    </div>
                    <p class="text-xs text-slate-500">
                      {formatActivityTime(entry.activity.createdAt)}
                    </p>
                  </div>
                  <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                    <span>{formatActivityActor(entry.activity.actor)}</span>
                    <span>owner {entry.operatorAssignee || 'unassigned'}</span>
                  </div>
                  <p class="mt-3 text-sm text-slate-700">
                    {entry.activity.message}
                  </p>
                  <Show when={entry.activity.details.length}>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <For each={entry.activity.details}>
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
          </div>
        </Show>
      </Card>
    </PageShell>
  );
}
