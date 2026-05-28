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
import {
  activateStore,
  createStore,
  listStores,
  type StoreInfo,
} from '../../../services/store';
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
  stores: StoreInfo[];
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
  const [creatingStoreTenantId, setCreatingStoreTenantId] = createSignal('');
  const [storeNameByTenant, setStoreNameByTenant] = createSignal<
    Record<string, string>
  >({});
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
  const [canManagePlatformIAM, setCanManagePlatformIAM] = createSignal(false);
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active');
  const canBootstrapFirstWorkspace = () => !!userID && memberships().length === 0;

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
          stores,
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
      if (result.data.length === 0) {
        setCanCreateTenant(true);
      }
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
    if (canBootstrapFirstWorkspace()) {
      setCanCreateTenant(true);
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
      const iamResult = await checkPlatformPermission('platform:manage_roles');
      setCanManagePlatformIAM(iamResult.success ? iamResult.data : false);
    } finally {
      setCheckingPlatformAccess(false);
    }
  };

  const openStore = async (nextTenantID: string, nextStoreID: string) => {
    const normalizedTenantID = nextTenantID.trim();
    const normalizedStoreID = nextStoreID.trim();
    if (!normalizedTenantID || !normalizedStoreID) return;

    setSwitchingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const { success, data } = await ensureActiveTenant(normalizedTenantID);
      if (!success) {
        setTenantError(data.message || 'Failed to open store');
        return;
      }

      tenantStorage.setTenantID(normalizedTenantID);
      storeStorage.setStoreID(normalizedTenantID, normalizedStoreID);
      setTenantId(normalizedTenantID);
      window.location.href = `/t/${normalizedTenantID}?storeId=${encodeURIComponent(normalizedStoreID)}`;
    } finally {
      setSwitchingTenant(false);
    }
  };

  const prepareTenant = async (nextTenantID = tenantId().trim()) => {
    if (!nextTenantID) return;

    setSwitchingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const { success, data } = await ensureActiveTenant(nextTenantID);
      if (!success) {
        setTenantError(data.message || 'Failed to load workspace');
        return;
      }
      tenantStorage.setTenantID(nextTenantID);
      setTenantId(nextTenantID);
      setTenantMessage(`Loaded workspace ${nextTenantID}. Choose a store below.`);
      await loadMemberships();
    } finally {
      setSwitchingTenant(false);
    }
  };

  const setDraftStoreName = (nextTenantID: string, value: string) => {
    setStoreNameByTenant((current) => ({
      ...current,
      [nextTenantID]: value,
    }));
  };

  const submitCreateStore = async (nextTenantID: string) => {
    const normalizedTenantID = nextTenantID.trim();
    const normalizedStoreName = (storeNameByTenant()[normalizedTenantID] || '').trim();
    if (!normalizedTenantID || !normalizedStoreName) return;

    setCreatingStoreTenantId(normalizedTenantID);
    setTenantError('');
    setTenantMessage('');
    try {
      const switched = await ensureActiveTenant(normalizedTenantID);
      if (!switched.success) {
        setTenantError(switched.data.message || 'Failed to load workspace');
        return;
      }
      const created = await createStore({ name: normalizedStoreName });
      if (!created.success) {
        setTenantError(created.message);
        return;
      }
      const activated = await activateStore(created.data.id);
      if (!activated.success) {
        setTenantError(activated.message);
        return;
      }
      setDraftStoreName(normalizedTenantID, '');
      await loadMemberships();
      await openStore(normalizedTenantID, activated.data.id);
    } finally {
      setCreatingStoreTenantId('');
    }
  };

  const submitCreateTenant = async (event: SubmitEvent) => {
    event.preventDefault();

    if (!userID) {
      setTenantError('No signed-in account found.');
      return;
    }
    if (!canCreateTenant() && !canBootstrapFirstWorkspace()) {
      setTenantError('Your account cannot create another workspace yet.');
      return;
    }

    const normalizedName = tenantName().trim();
    const normalizedSlug = slugify(tenantSlug() || normalizedName);
    if (!normalizedName || !normalizedSlug) {
      setTenantError('Workspace name and slug are required.');
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
          ? `Created workspace ${createdSlug} (${createdTenantID}).`
          : `Created workspace ${createdSlug}.`
      );
      await loadMemberships();
      await loadPlatformAccess();
    } finally {
      setCreatingTenant(false);
    }
  };

  onMount(() => {
    void (async () => {
      await loadMemberships();
      await loadPlatformAccess();
    })();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Seller Backoffice"
          title="Choose the store you want to operate."
          copy="Start from a tenant workspace, create or select a store, then enter the store-scoped backoffice."
        />
        <div class="flex flex-wrap gap-3">
          <Show when={canManagePlatformIAM()}>
            <Button href="/admin/iam" color="dark" size="sm">
              Open IAM console
            </Button>
          </Show>
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
          value={
            tokenStorage.getActiveTenantID()
              ? storeStorage.getStoreID(tokenStorage.getActiveTenantID()) ||
                'Not selected'
              : 'Not selected'
          }
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
        <LoadingInline label="Checking workspace creation access..." />
      </Show>

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create workspace"
            subtitle="Create the tenant workspace that will own your stores and IAM memberships."
          />
          <Show when={!canCreateTenant() && !checkingPlatformAccess()}>
            <InfoAlert>
              This account already owns a workspace. Creating additional
              workspaces requires platform approval.
            </InfoAlert>
          </Show>
          <form class="space-y-4" onSubmit={submitCreateTenant}>
            <InputField
              label="Workspace name"
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
              label="Workspace slug"
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
                disabled={
                  !tenantName().trim() ||
                  (!canCreateTenant() && !canBootstrapFirstWorkspace())
                }
              >
                Create workspace
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
            title="Load workspace"
            subtitle="Use direct tenant loading only when you know the ID. Store selection still happens below."
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
                void prepareTenant();
              }}
            >
              Load workspace
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
            subtitle="Pick one store to enter. To switch later, return here and choose another store."
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
                    <div class="mt-4 space-y-3">
                      <Show
                        when={workspace.stores.length > 0}
                        fallback={
                          <EmptyBlock
                            title="No stores in this workspace"
                            copy="Create the first store here, then the backoffice will open in that store scope."
                          />
                        }
                      >
                        <For each={workspace.stores}>
                          {(store) => (
                            <div class="flex flex-col gap-3 rounded-md border border-gray-200 bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
                              <div class="min-w-0">
                                <div class="truncate text-sm font-semibold text-gray-950">
                                  {store.name}
                                </div>
                                <div class="mt-1 flex flex-wrap gap-2">
                                  <Badge
                                    content={store.status || 'unknown'}
                                    color={store.isActive ? 'green' : 'dark'}
                                  />
                                  <Badge content={store.id} color="dark" />
                                </div>
                              </div>
                              <Button
                                size="sm"
                                onClick={() => {
                                  void openStore(workspace.tenantId, store.id);
                                }}
                              >
                                Open store
                              </Button>
                            </div>
                          )}
                        </For>
                      </Show>
                      <div class="flex flex-col gap-3 rounded-md border border-dashed border-gray-300 bg-white p-3 sm:flex-row">
                        <input
                          class="block h-9 min-w-0 flex-1 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
                          value={storeNameByTenant()[workspace.tenantId] || ''}
                          placeholder="New store name"
                          onInput={(event) =>
                            setDraftStoreName(
                              workspace.tenantId,
                              event.currentTarget.value
                            )
                          }
                        />
                        <Button
                          size="sm"
                          color="alternative"
                          loading={creatingStoreTenantId() === workspace.tenantId}
                          disabled={
                            creatingStoreTenantId() === workspace.tenantId ||
                            !(storeNameByTenant()[workspace.tenantId] || '').trim()
                          }
                          onClick={() => {
                            void submitCreateStore(workspace.tenantId);
                          }}
                        >
                          Create store
                        </Button>
                      </div>
                      <Button
                        size="sm"
                        color="light"
                        onClick={() => {
                          setTenantId(workspace.tenantId);
                          setTenantMessage(
                            `Prepared workspace ${workspace.tenantId} as the next quick-jump target.`
                          );
                        }}
                      >
                        Use for direct load
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
                Use the settings page to manage team access, workspace invites,
                sessions, and platform administration.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </PageShell>
  );
}
