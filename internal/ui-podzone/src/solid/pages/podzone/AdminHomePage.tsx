import {
  For,
  Show,
  createEffect,
  createSignal,
  onCleanup,
  onMount,
} from 'solid-js';
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
  createStoreRequest,
  listStoreRequests,
  retryStoreRequest,
  type StoreRequest,
  type StoreRequestStatus,
} from '../../../services/onboarding';
import { listStores, type StoreInfo } from '../../../services/store';
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
  SelectField,
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

const provisioningSteps: StoreRequestStatus[] = [
  'requested',
  'queued',
  'provisioning',
  'ready',
];

function provisioningStepIndex(status: StoreRequestStatus) {
  if (status === 'pending_approval') return 0;
  if (status === 'failed') return 2;
  return provisioningSteps.indexOf(status);
}

function provisioningStatusLabel(status: StoreRequestStatus) {
  switch (status) {
    case 'pending_approval':
      return 'Pending approval';
    case 'queued':
      return 'Queued';
    case 'provisioning':
      return 'Provisioning infrastructure';
    case 'ready':
      return 'Ready';
    case 'failed':
      return 'Provisioning failed';
    default:
      return status.charAt(0).toUpperCase() + status.slice(1);
  }
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
  storeRequests: StoreRequest[];
  storeCount: number;
  activeStoreCount: number;
};

export default function AdminHomePage() {
  const user = tokenStorage.getUser();
  const userID = parseUserID(user?.id);

  const [tenantName, setTenantName] = createSignal('');
  const [tenantSlug, setTenantSlug] = createSignal('');
  const [tenantError, setTenantError] = createSignal('');
  const [tenantMessage, setTenantMessage] = createSignal('');
  const [switchingTenant, setSwitchingTenant] = createSignal(false);
  const [creatingTenant, setCreatingTenant] = createSignal(false);
  const [creatingStoreTenantId, setCreatingStoreTenantId] = createSignal('');
  const [retryingStoreRequestId, setRetryingStoreRequestId] = createSignal('');
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
  const [selectedWorkspaceId, setSelectedWorkspaceId] = createSignal('');
  const [selectedStoreId, setSelectedStoreId] = createSignal('');
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active');
  const canBootstrapFirstWorkspace = () => !!userID && memberships().length === 0;
  const activeWorkspaceSummaries = () =>
    workspaceSummaries().filter(
      (workspace) => workspace.status === 'active'
    );
  const selectedWorkspace = () =>
    activeWorkspaceSummaries().find(
      (workspace) => workspace.tenantId === selectedWorkspaceId()
    );
  const selectedWorkspaceOptions = () =>
    activeWorkspaceSummaries().map((workspace) => ({
      name: `${workspace.tenantId} · ${workspace.roleName}`,
      value: workspace.tenantId,
    }));
  const selectedStoreOptions = () =>
    (selectedWorkspace()?.stores || []).map((store) => ({
      name: `${store.name}${store.isActive ? ' · active' : ''}`,
      value: store.id,
    }));
  const currentSelectionLabel = () => {
    const workspace = selectedWorkspace();
    const store = workspace?.stores.find((item) => item.id === selectedStoreId());
    if (!workspace) return 'No workspace selected';
    if (!store) return `${workspace.tenantId} selected, no store chosen`;
    return `${workspace.tenantId} / ${store.name}`;
  };

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
        const requestsResult = await listStoreRequests(membership.tenantId);
        const storeRequests = requestsResult.success ? requestsResult.data : [];
        const storesResult = await listStores();
        const stores = storesResult.success ? storesResult.data : [];
        summaries.push({
          tenantId: membership.tenantId,
          roleName: membership.roleName,
          status: membership.status,
          userId: membership.userId,
          stores,
          storeRequests,
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

  createEffect(() => {
    const nextWorkspace =
      selectedWorkspaceId() ||
      activeWorkspaceSummaries()[0]?.tenantId ||
      activeMemberships()[0]?.tenantId ||
      '';
    if (!nextWorkspace) return;
    if (nextWorkspace !== selectedWorkspaceId()) {
      setSelectedWorkspaceId(nextWorkspace);
    }
  });

  createEffect(() => {
    const stores = selectedWorkspace()?.stores || [];
    if (stores.length === 0) {
      setSelectedStoreId('');
      return;
    }
    const current = selectedStoreId();
    if (stores.some((store) => store.id === current)) {
      return;
    }
    const preferred = stores.find((store) => store.isActive) || stores[0];
    setSelectedStoreId(preferred?.id || '');
  });

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
      window.location.href = `/t/${normalizedTenantID}?storeId=${encodeURIComponent(normalizedStoreID)}`;
    } finally {
      setSwitchingTenant(false);
    }
  };

  const prepareTenant = async (nextTenantID: string) => {
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
      const created = await createStoreRequest({
        tenantId: normalizedTenantID,
        name: normalizedStoreName,
        subdomain: slugify(normalizedStoreName),
      });
      if (!created.success) {
        setTenantError(created.message);
        return;
      }
      setDraftStoreName(normalizedTenantID, '');
      setTenantMessage(
        `Store request ${created.data.name} is ${created.data.status}. It will become selectable after provisioning completes.`
      );
      await loadMemberships();
    } finally {
      setCreatingStoreTenantId('');
    }
  };

  const retryStore = async (tenantID: string, requestID: string) => {
    setRetryingStoreRequestId(requestID);
    setTenantError('');
    setTenantMessage('');
    try {
      const switched = await ensureActiveTenant(tenantID);
      if (!switched.success) {
        setTenantError(switched.data.message || 'Failed to load workspace');
        return;
      }
      const result = await retryStoreRequest({
        tenantId: tenantID,
        requestId: requestID,
      });
      if (!result.success) {
        setTenantError(result.message);
        return;
      }
      setTenantMessage('Store provisioning has been queued again.');
      await loadMemberships();
    } finally {
      setRetryingStoreRequestId('');
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
      setSelectedWorkspaceId(createdTenantID);
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

    const refreshTimer = window.setInterval(() => {
      const hasActiveProvisioning = workspaceSummaries().some((workspace) =>
        workspace.storeRequests.some((request) =>
          ['requested', 'pending_approval', 'queued', 'provisioning'].includes(
            request.status
          )
        )
      );
      if (hasActiveProvisioning && !loadingTenants()) {
        void loadMemberships();
      }
    }, 10000);
    onCleanup(() => window.clearInterval(refreshTimer));
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
          <Button href="/admin/iam" color="dark" size="sm">
            Open IAM console
          </Button>
        </div>
        <Show when={!canManagePlatformIAM()}>
          <InfoAlert>
            Platform IAM is available only for platform admins. This session can still manage workspace and store access.
          </InfoAlert>
        </Show>
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

      <Card class="space-y-4">
        <SectionTitle
          title="Choose store"
          subtitle="Pick a workspace, then choose a store inside it. If the workspace has no store yet, create one right here."
        />
        <div class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
          <div class="space-y-3">
            <SelectField
              label="Workspace"
              value={selectedWorkspaceId()}
              options={selectedWorkspaceOptions()}
              onChange={(event) => {
                setSelectedWorkspaceId(event.currentTarget.value);
              }}
            />
            <p class="text-sm text-gray-600">{currentSelectionLabel()}</p>
          </div>

          <div class="space-y-3">
            <Show
              when={selectedWorkspace() && selectedStoreOptions().length > 0}
              fallback={
                <div class="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4">
                  <p class="text-sm font-semibold text-gray-950">
                    No store in this workspace yet
                  </p>
                  <p class="text-sm text-gray-600">
                    Create the first store below, then select it from this workspace.
                  </p>
                  <div class="flex flex-col gap-3 sm:flex-row">
                    <input
                      class="block h-10 min-w-0 flex-1 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
                      value={storeNameByTenant()[selectedWorkspaceId()] || ''}
                      placeholder="New store name"
                      onInput={(event) =>
                        setDraftStoreName(
                          selectedWorkspaceId(),
                          event.currentTarget.value
                        )
                      }
                    />
                    <Button
                      size="sm"
                      color="dark"
                      loading={creatingStoreTenantId() === selectedWorkspaceId()}
                      disabled={
                        !selectedWorkspaceId() ||
                        creatingStoreTenantId() === selectedWorkspaceId() ||
                        !(storeNameByTenant()[selectedWorkspaceId()] || '').trim()
                      }
                      onClick={() => {
                        void submitCreateStore(selectedWorkspaceId());
                      }}
                    >
                      Create store
                    </Button>
                  </div>
                </div>
              }
            >
              <SelectField
                label="Store"
                value={selectedStoreId()}
                options={selectedStoreOptions()}
                onChange={(event) => setSelectedStoreId(event.currentTarget.value)}
              />
              <div class="flex flex-wrap gap-3">
                <Button
                  disabled={!selectedWorkspaceId() || !selectedStoreId()}
                  loading={switchingTenant()}
                  onClick={() => {
                    void openStore(selectedWorkspaceId(), selectedStoreId());
                  }}
                >
                  Open selected store
                </Button>
                <Button
                  color="light"
                  disabled={!selectedWorkspaceId() || !selectedStoreId()}
                  onClick={() => {
                    const current = selectedWorkspace();
                    const store = current?.stores.find(
                      (item) => item.id === selectedStoreId()
                    );
                    if (!current || !store) return;
                    setTenantMessage(
                      `Selected ${store.name} in ${current.tenantId}.`
                    );
                  }}
                >
                  Confirm selection
                </Button>
              </div>
            </Show>
          </div>
        </div>
      </Card>

      <Show when={canCreateTenant() || canBootstrapFirstWorkspace()}>
        <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
          <Card class="space-y-4">
            <SectionTitle
              title="Create workspace"
              subtitle="Create the tenant workspace that will own your stores and IAM memberships."
            />
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
              title="Current workspace"
              subtitle="The workspace you are preparing to enter."
            />
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
              <p class="text-sm font-semibold text-gray-950">
                {selectedWorkspace()?.tenantId || 'No workspace selected'}
              </p>
              <p class="mt-1 text-sm text-gray-600">
                {selectedWorkspace()?.roleName || 'Select a workspace above'}
              </p>
            </div>
            <Button
              color="alternative"
              disabled={!selectedWorkspaceId()}
              onClick={() => {
                void prepareTenant(selectedWorkspaceId());
              }}
            >
              Reload workspace
            </Button>
          </Card>
        </div>
      </Show>

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
                      <Show when={workspace.storeRequests.length > 0}>
                        <For each={workspace.storeRequests}>
                          {(request) => (
                            <div class="rounded-md border border-gray-200 bg-white p-3">
                              <div class="flex flex-wrap items-center justify-between gap-2">
                                <div class="min-w-0">
                                  <div class="truncate text-sm font-semibold text-gray-950">
                                    {request.name}
                                  </div>
                                  <div class="mt-1 text-xs text-gray-500">
                                    {request.subdomain}
                                  </div>
                                </div>
                                <Badge
                                  content={provisioningStatusLabel(request.status)}
                                  color={request.status === 'ready' ? 'green' : request.status === 'failed' ? 'red' : 'yellow'}
                                />
                              </div>
                              <div
                                class="mt-3 grid grid-cols-4 gap-1"
                                aria-label={`Provisioning progress: ${provisioningStatusLabel(request.status)}`}
                              >
                                <For each={provisioningSteps}>
                                  {(step, index) => (
                                    <div>
                                      <div
                                        class={`h-1.5 rounded-full ${
                                          index() <= provisioningStepIndex(request.status)
                                            ? request.status === 'failed'
                                              ? 'bg-red-500'
                                              : 'bg-gray-950'
                                            : 'bg-gray-200'
                                        }`}
                                      />
                                      <p class="mt-1 truncate text-[11px] text-gray-500">
                                        {provisioningStatusLabel(step)}
                                      </p>
                                    </div>
                                  )}
                                </For>
                              </div>
                              <Show when={request.last_error}>
                                <p class="mt-2 text-sm text-red-700">
                                  {request.last_error}
                                </p>
                              </Show>
                              <Show when={request.status === 'failed'}>
                                <div class="mt-3">
                                  <Button
                                    size="xs"
                                    color="alternative"
                                    loading={retryingStoreRequestId() === request.id}
                                    disabled={retryingStoreRequestId() === request.id}
                                    onClick={() => {
                                      void retryStore(
                                        workspace.tenantId,
                                        request.id
                                      );
                                    }}
                                  >
                                    Retry provisioning
                                  </Button>
                                </div>
                              </Show>
                            </div>
                          )}
                        </For>
                      </Show>
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
