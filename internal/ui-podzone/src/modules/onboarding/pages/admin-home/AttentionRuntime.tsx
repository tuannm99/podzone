import { For, Show } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '@/services/baseurl';
import { EmptyBlock, LoadingInline } from '@/solid/components/common/Feedback';
import { Badge, Button, Card } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { useAdminHome } from './context';

export function AttentionRuntime() {
  return (
    <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
      <AttentionView />
      <RuntimeEndpoints />
    </div>
  );
}

function AttentionView() {
  const vm = useAdminHome();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Attention view"
        subtitle="Cross-store queue pressure grouped by real stores inside each tenant workspace."
      />
      <Show when={vm.loadingAttention()}>
        <LoadingInline label="Scanning store attention..." />
      </Show>
      <Show
        when={!vm.loadingAttention() && vm.storeAttention().length > 0}
        fallback={
          <EmptyBlock
            title="No store attention data"
            copy="Create or open stores first, then this panel will surface where the queue needs follow-up."
          />
        }
      >
        <div class="space-y-3">
          <For each={vm.storeAttention()}>
            {(store) => (
              <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                <div class="flex flex-wrap items-center gap-2">
                  <Badge content={store.tenantId} color="indigo" />
                  <Badge content={store.storeName} color="blue" />
                  <Badge
                    content={`${store.overdueCount} overdue`}
                    color={store.overdueCount > 0 ? 'red' : 'green'}
                  />
                  <Badge
                    content={`${store.disputedCount} disputed`}
                    color={store.disputedCount > 0 ? 'red' : 'green'}
                  />
                  <Badge
                    content={`${store.unassignedCount} unassigned`}
                    color={store.unassignedCount > 0 ? 'yellow' : 'green'}
                  />
                </div>
                <div class="mt-3 flex flex-wrap gap-3">
                  <Button
                    size="sm"
                    color="alternative"
                    href={vm.buildOrdersHref(
                      store.tenantId,
                      store.storeId,
                      'overdue'
                    )}
                  >
                    Open overdue
                  </Button>
                  <Button
                    size="sm"
                    color="alternative"
                    href={vm.buildOrdersHref(
                      store.tenantId,
                      store.storeId,
                      'settlement_pending'
                    )}
                  >
                    Open settlement follow-up
                  </Button>
                  <Button
                    size="sm"
                    color="alternative"
                    href={vm.buildOrdersHref(
                      store.tenantId,
                      store.storeId,
                      'all'
                    )}
                  >
                    Open priority queue
                  </Button>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </Card>
  );
}

function RuntimeEndpoints() {
  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Runtime endpoints"
        subtitle="Current application entrypoints used by the backoffice."
      />
      <div class="space-y-3 text-sm text-gray-600">
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-950">Gateway</p>
          <p class="mt-1 break-all">{GW_API_URL}</p>
        </div>
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-950">Store GraphQL</p>
          <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
        </div>
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-950">Next step</p>
          <p class="mt-1">
            Use the settings page to manage team access, workspace invites,
            sessions, and platform administration.
          </p>
        </div>
      </div>
    </Card>
  );
}
