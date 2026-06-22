import {
  createEffect,
  createSignal,
  onCleanup,
  onMount,
} from 'solid-js';
import { ensureActiveTenant } from '@/services/auth';
import {
  checkPlatformPermission,
  createTenant,
  listUserTenants,
  type TenantMembership,
} from '@/services/iam';
import { getRoutedOrders } from '@/services/orders';
import {
  createStoreRequest,
  listStoreRequests,
  retryStoreRequest,
} from '@/services/onboarding';
import { listStores } from '@/services/store';
import { storeStorage } from '@/services/storeStorage';
import { tenantStorage } from '@/services/tenantStorage';
import { tokenStorage } from '@/services/tokenStorage';
import { AdminHomeContext } from './admin-home/context';
import { AdminHomeView } from './admin-home/AdminHomeView';
import {
  buildOrdersHref,
  isOverdue,
  membershipStatusColor,
  parseUserID,
  provisioningStepIndex,
  provisioningStatusLabel,
  provisioningSteps,
  slugify,
  type StoreAttention,
  type WorkspaceSummary,
} from './admin-home/presentation';

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
  const [storeAttention, setStoreAttention] = createSignal<StoreAttention[]>(
    []
  );
  const [canCreateTenant, setCanCreateTenant] = createSignal(false);
  const [canManagePlatformIAM, setCanManagePlatformIAM] = createSignal(false);
  const [selectedWorkspaceId, setSelectedWorkspaceId] = createSignal('');
  const [selectedStoreId, setSelectedStoreId] = createSignal('');
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active');
  const canBootstrapFirstWorkspace = () =>
    !!userID && memberships().length === 0;
  const activeWorkspaceSummaries = () =>
    workspaceSummaries().filter((workspace) => workspace.status === 'active');
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
    const store = workspace?.stores.find(
      (item) => item.id === selectedStoreId()
    );
    if (!workspace) return 'No workspace selected';
    if (!store) return `${workspace.tenantId} selected, no store chosen`;
    return `${workspace.tenantId} / ${store.name}`;
  };

  const loadWorkspaceData = async (
    membershipsToInspect: TenantMembership[]
  ) => {
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
      setTenantMessage(
        `Loaded workspace ${nextTenantID}. Choose a store below.`
      );
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
    const normalizedStoreName = (
      storeNameByTenant()[normalizedTenantID] || ''
    ).trim();
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

  const viewModel = {
    user, userID, tenantName, setTenantName, tenantSlug, setTenantSlug, tenantError, tenantMessage, setTenantMessage, switchingTenant, creatingTenant, creatingStoreTenantId, retryingStoreRequestId, storeNameByTenant, loadingTenants, loadingAttention, checkingPlatformAccess, memberships, workspaceSummaries, storeAttention, canCreateTenant, canManagePlatformIAM, selectedWorkspaceId, setSelectedWorkspaceId, selectedStoreId, setSelectedStoreId, activeMemberships, canBootstrapFirstWorkspace, activeWorkspaceSummaries, selectedWorkspace, selectedWorkspaceOptions, selectedStoreOptions, currentSelectionLabel, slugify, membershipStatusColor, provisioningSteps, provisioningStepIndex, provisioningStatusLabel, buildOrdersHref, loadWorkspaceData, prepareTenant, openStore, setDraftStoreName, submitCreateStore, retryStore, submitCreateTenant,
  };

  return (
    <AdminHomeContext.Provider value={viewModel}>
      <AdminHomeView />
    </AdminHomeContext.Provider>
  );
}
