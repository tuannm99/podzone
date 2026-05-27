import { For, Show, createSignal, onMount } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import { ensureActiveTenant } from '../../../services/auth';
import {
  checkPlatformPermission,
  createTenant,
  listUserTenants,
  type TenantMembership,
} from '../../../services/iam';
import { getRoutedOrders } from '../../../services/orders';
import { listStores } from '../../../services/store';
import { storeStorage } from '../../../services/storeStorage';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

function membershipStatusColor(status: string) {
  return status === 'active' ? 'green' : 'dark';
}

function isOverdue(value?: string) {
  if (!value) {
    return false;
  }
  return new Date(value).getTime() < Date.now();
}

function buildOrdersHref(tenantID: string, storeID: string, queueView: string) {
  const params = new URLSearchParams({
    storeId: storeID,
    queueView,
    queueSort: 'priority',
  });
  return `/t/${tenantID}/orders?${params.toString()}`;
}

type StoreAttention = {
  tenantId: string;
  storeId: string;
  storeName: string;
  overdueCount: number;
  disputedCount: number;
  unassignedCount: number;
};

type WorkspaceSummary = {
  tenantId: string;
  roleName: string;
  status: string;
  userId: number;
  storeCount: number;
  activeStoreCount: number;
};

export default function AdminHomePage() {
  const user = tokenStorage.getUser();
  const userID = parseUserID(user?.id);

  const [tenantId, setTenantId] = createSignal(
    tokenStorage.getActiveTenantID() || tenantStorage.getTenantID()
  );
  const [tenantName, setTenantName] = createSignal('');
  const [tenantSlug, setTenantSlug] = createSignal('');
  const [tenantError, setTenantError] = createSignal('');
  const [tenantMessage, setTenantMessage] = createSignal('');
  const [switchingTenant, setSwitchingTenant] = createSignal(false);
  const [creatingTenant, setCreatingTenant] = createSignal(false);
  const [loadingTenants, setLoadingTenants] = createSignal(false);
  const [loadingAttention, setLoadingAttention] = createSignal(false);
  const [checkingPlatformAccess, setCheckingPlatformAccess] =
    createSignal(false);
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);
  const [workspaceSummaries, setWorkspaceSummaries] = createSignal<
    WorkspaceSummary[]
  >([]);
  const [storeAttention, setStoreAttention] = createSignal<StoreAttention[]>([]);
  const [canCreateTenant, setCanCreateTenant] = createSignal(false);
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active');

  const loadWorkspaceData = async (membershipsToInspect: TenantMembership[]) => {
    const activeWorkspaces = membershipsToInspect.filter(
      (membership) => membership.status === 'active'
    );
    if (activeWorkspaces.length === 0) {
      setWorkspaceSummaries([]);
      setStoreAttention([]);
      return;
    }

    setLoadingAttention(true);
    const originalTenantID = tokenStorage.getActiveTenantID();
    const originalStoreID = originalTenantID
      ? storeStorage.getStoreID(originalTenantID)
      : '';
    const previousStoreByTenant = new Map<string, string>();
    const summaries: WorkspaceSummary[] = [];
    const snapshot: StoreAttention[] = [];
    try {
      for (const membership of activeWorkspaces) {
        const switched = await ensureActiveTenant(membership.tenantId);
        if (!switched.success) {
          continue;
        }
        const storesResult = await listStores();
        if (!storesResult.success) {
          continue;
        }
        const stores = storesResult.data;
        summaries.push({
          tenantId: membership.tenantId,
          roleName: membership.roleName,
          status: membership.status,
          userId: membership.userId,
          storeCount: stores.length,
          activeStoreCount: stores.filter((store) => store.isActive).length,
        });
        for (const store of stores) {
          if (!previousStoreByTenant.has(membership.tenantId)) {
            previousStoreByTenant.set(
              membership.tenantId,
              storeStorage.getStoreID(membership.tenantId)
            );
          }
          storeStorage.setStoreID(membership.tenantId, store.id);
          const ordersResult = await getRoutedOrders();
          if (!ordersResult.success) {
            continue;
          }
          const orders = ordersResult.data.orders;
          snapshot.push({
            tenantId: membership.tenantId,
            storeId: store.id,
            storeName: store.name,
            overdueCount: orders.filter(
              (order) =>
                (!!order.shipmentSlaDueAt &&
                  isOverdue(order.shipmentSlaDueAt) &&
                  order.shipmentStatus !== 'delivered') ||
                (!!order.issueSlaDueAt &&
                  isOverdue(order.issueSlaDueAt) &&
                  (order.exceptionStatus === 'open' ||
                    order.exceptionStatus === 'escalated' ||
                    order.shipmentStatus === 'delivery_issue'))
            ).length,
            disputedCount: orders.filter(
              (order) => order.settlementStatus === 'disputed'
            ).length,
            unassignedCount: orders.filter(
              (order) =>
                !order.operatorAssignee ||
                order.operatorAssignee === 'unassigned'
            ).length,
          });
        }
      }
    } finally {
      previousStoreByTenant.forEach((storeID, tenantID) => {
        if (storeID) {
          storeStorage.setStoreID(tenantID, storeID);
        } else {
          storeStorage.clearStoreID(tenantID);
        }
      });
      if (originalTenantID) {
        await ensureActiveTenant(originalTenantID);
        tenantStorage.setTenantID(originalTenantID);
        if (originalStoreID) {
          storeStorage.setStoreID(originalTenantID, originalStoreID);
        } else {
          storeStorage.clearStoreID(originalTenantID);
        }
      }
      setLoadingAttention(false);
    }
    setWorkspaceSummaries(summaries);
    setStoreAttention(snapshot);
  };

  const loadMemberships = async () => {
    if (!userID) return;
    setLoadingTenants(true);
    try {
      const result = await listUserTenants(userID);
      if (!result.success) {
        setTenantError(result.message);
        return;
      }
      setMemberships(result.data);
      await loadWorkspaceData(result.data);
    } finally {
      setLoadingTenants(false);
    }
  };

  const loadPlatformAccess = async () => {
    if (!userID) {
      setCanCreateTenant(false);
      return;
    }
    setCheckingPlatformAccess(true);
    try {
      const result = await checkPlatformPermission('tenant:create');
      if (!result.success) {
        setTenantError(result.message);
        setCanCreateTenant(false);
        return;
      }
      setCanCreateTenant(result.data);
    } finally {
      setCheckingPlatformAccess(false);
    }
  };

  const openTenant = async (nextTenantID = tenantId().trim()) => {
    if (!nextTenantID) return;

    setSwitchingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const { success, data } = await ensureActiveTenant(nextTenantID);
      if (!success) {
      setTenantError(data.message || 'Failed to open store');
        return;
      }

      tenantStorage.setTenantID(nextTenantID);
      setTenantId(nextTenantID);
      window.location.href = `/t/${nextTenantID}`;
    } finally {
      setSwitchingTenant(false);
    }
  };

  const submitCreateTenant = async (event: SubmitEvent) => {
    event.preventDefault();

    if (!userID) {
      setTenantError('No signed-in account found.');
      return;
    }
    if (!canCreateTenant()) {
      setTenantError('Your account cannot create stores yet.');
      return;
    }

    const normalizedName = tenantName().trim();
    const normalizedSlug = slugify(tenantSlug() || normalizedName);
    if (!normalizedName || !normalizedSlug) {
      setTenantError('Store name and store slug are required.');
      return;
    }

    setCreatingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const result = await createTenant({
        name: normalizedName,
        slug: normalizedSlug,
      });
      if (!result.success) {
        setTenantError(result.message);
        return;
      }

      const createdTenantID = result.data.tenant?.id || '';
      const createdSlug = result.data.tenant?.slug || normalizedSlug;
      setTenantName('');
      setTenantSlug('');
      setTenantId(createdTenantID);
      setTenantMessage(
        createdTenantID
          ? `Created store ${createdSlug} (${createdTenantID}).`
          : `Created store ${createdSlug}.`
      );
      await loadMemberships();
    } finally {
      setCreatingTenant(false);
    }
  };

  onMount(() => {
    void loadPlatformAccess();
    void loadMemberships();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Seller Backoffice"
          title="Manage your stores from one control room."
          copy="Create a new store, review where your team has access, and open the right workspace without relying on technical IDs."
        />
        <div class="flex flex-wrap gap-3">
          <Button href="/admin/iam" color="dark" size="sm">
            Open IAM console
          </Button>
        </div>
      </Card>

      <Show when={tenantError()}>
        <ErrorAlert>{tenantError()}</ErrorAlert>
      </Show>

      <Show when={tenantMessage()}>
        <InfoAlert>{tenantMessage()}</InfoAlert>
      </Show>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Session"
          value={tokenStorage.getToken() ? 'Active' : 'Missing'}
        />
        <StatCard
          label="User"
          value={user?.username || user?.email || 'Unknown'}
        />
        <StatCard
          label="My workspaces"
          value={`${activeMemberships().length}/${memberships().length}`}
        />
        <StatCard
          label="Current store"
          value={tokenStorage.getActiveTenantID() || 'Not selected'}
        />
        <StatCard
          label="Stores with attention"
          value={String(
            storeAttention().filter(
              (store) =>
                store.overdueCount > 0 ||
                store.disputedCount > 0 ||
                store.unassignedCount > 0
            ).length
          )}
        />
      </div>

      <Show when={checkingPlatformAccess()}>
        <LoadingInline label="Checking store creation access..." />
      </Show>

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create store"
            subtitle="Open a new store workspace and make yourself the store owner."
          />
          <Show when={!canCreateTenant() && !checkingPlatformAccess()}>
            <InfoAlert>
              Creating stores requires platform approval. The first platform
              owner can grant this access to additional operators.
            </InfoAlert>
          </Show>
          <form class="space-y-4" onSubmit={submitCreateTenant}>
            <InputField
              label="Store name"
              value={tenantName()}
              placeholder="Urban Finds"
              onInput={(event) => {
                const value = event.currentTarget.value;
                setTenantName(value);
                if (!tenantSlug().trim()) {
                  setTenantSlug(slugify(value));
                }
              }}
            />
            <InputField
              label="Store slug"
              value={tenantSlug()}
              placeholder="urban-finds"
              onInput={(event) =>
                setTenantSlug(slugify(event.currentTarget.value))
              }
            />
            <div class="flex flex-wrap gap-3">
              <Button
                type="submit"
                loading={creatingTenant()}
                disabled={!tenantName().trim() || !canCreateTenant()}
              >
                Create store
              </Button>
              <Badge
                content={
                  tenantSlug().trim()
                    ? `slug ${tenantSlug().trim()}`
                    : 'slug pending'
                }
                color={tenantSlug().trim() ? 'indigo' : 'dark'}
              />
            </div>
          </form>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Quick workspace jump"
            subtitle="Keep direct tenant opening for debugging, but prefer opening from your assigned workspaces below."
          />
          <div class="flex flex-col gap-3 sm:flex-row">
            <input
              class="block h-10 w-full rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
              value={tenantId()}
              placeholder="tenant id"
              onInput={(event) => setTenantId(event.currentTarget.value)}
            />
            <Button
              disabled={!tenantId().trim() || switchingTenant()}
              loading={switchingTenant()}
              onClick={() => {
                void openTenant();
              }}
            >
              Open workspace
            </Button>
          </div>
          {!tenantId().trim() ? (
            <EmptyBlock
              title="No workspace selected"
              copy="Create a tenant workspace or pick one from your assigned workspaces to open the right shell."
            />
          ) : null}
        </Card>
      </div>

      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="My workspaces"
            subtitle="Tenant workspaces your account can access, including how many stores each workspace currently owns."
          />
          <Show when={loadingTenants()}>
            <LoadingInline label="Loading workspaces..." />
          </Show>
          <Show
            when={!loadingTenants() && workspaceSummaries().length > 0}
            fallback={
              <EmptyBlock
                title="No workspaces yet"
                copy="Create your first tenant workspace to start managing stores, catalog, team access, and POD operations."
              />
            }
          >
            <div class="space-y-3">
              <For each={workspaceSummaries()}>
                {(workspace) => (
                  <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={workspace.roleName} color="blue" />
                      <Badge
                        content={workspace.status}
                        color={membershipStatusColor(workspace.status)}
                      />
                      <Show
                        when={
                          workspace.tenantId ===
                          tokenStorage.getActiveTenantID()
                        }
                      >
                        <Badge content="current" color="yellow" />
                      </Show>
                    </div>
                    <p class="mt-3 font-semibold text-gray-950">
                      {workspace.tenantId}
                    </p>
                    <p class="mt-1 text-sm text-gray-600">
                      access role for user {workspace.userId}
                    </p>
                    <p class="mt-1 text-sm text-gray-600">
                      {workspace.storeCount} stores · {workspace.activeStoreCount} active
                    </p>
                    <div class="mt-3 flex flex-wrap gap-3">
                      <Button
                        size="sm"
                        onClick={() => {
                          void openTenant(workspace.tenantId);
                        }}
                      >
                        Open workspace
                      </Button>
                      <Button
                        size="sm"
                        color="alternative"
                        onClick={() => {
                          setTenantId(workspace.tenantId);
                          setTenantMessage(
                            `Prepared workspace ${workspace.tenantId} as the next quick-jump target.`
                          );
                        }}
                      >
                        Use for quick jump
                      </Button>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Attention view"
            subtitle="Cross-store queue pressure grouped by real stores inside each tenant workspace."
          />
          <Show when={loadingAttention()}>
            <LoadingInline label="Scanning store attention..." />
          </Show>
          <Show
            when={!loadingAttention() && storeAttention().length > 0}
            fallback={
              <EmptyBlock
                title="No store attention data"
                copy="Create or open stores first, then this panel will surface where the queue needs follow-up."
              />
            }
          >
            <div class="space-y-3">
              <For each={storeAttention()}>
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
                        href={buildOrdersHref(
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
                        href={buildOrdersHref(
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
                        href={buildOrdersHref(
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
                Use the settings page to manage team access, store invites,
                sessions, and platform administration.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </PageShell>
  );
}
