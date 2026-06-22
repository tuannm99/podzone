import { useParams, useSearch } from '@tanstack/solid-router';
import { createEffect, createSignal } from 'solid-js';
import {
  getRoutedOrderRecommendation,
  getRoutedOrders,
  type RoutedOrder,
  type RoutedOrderRecommendation,
} from '@/services/orders';
import {
  getProductSetupSnapshot,
  type CatalogCandidate,
} from '@/services/productSetup';
import { tenantStorage } from '@/services/tenantStorage';
import { useTenantWorkspace } from '@/solid/workspace/context';
import type {
  QueueSort,
  QueueView,
  TenantOrdersBoardContextValue,
} from './orders/board-context';
import type { TenantOrdersComposerContextValue } from './orders/composer-context';
import type { TenantOrdersInsightsContextValue } from './orders/context';
import { TenantOrdersView } from './orders/TenantOrdersView';
import type { ActivityFilter } from './orders/order-card/types';
import {
  isQueueSort,
  isQueueView,
} from './orders/presentation';
import { useOrderActions } from './orders/useOrderActions';
import { useOrderDrafts } from './orders/useOrderDrafts';
import { useOrderInsights } from './orders/useOrderInsights';
import { useOrderStorage } from './orders/useOrderStorage';

export default function TenantOrdersPage() {
  const params = useParams({ from: '/t/$tenantId/orders' });
  const search = useSearch({ strict: false }) as () => Record<string, unknown>;
  const workspace = useTenantWorkspace();

  const [availableCandidates, setAvailableCandidates] = createSignal<
    CatalogCandidate[]
  >([]);
  const [orders, setOrders] = createSignal<RoutedOrder[]>([]);
  const [selectedCandidateId, setSelectedCandidateId] = createSignal('');
  const [customerName, setCustomerName] = createSignal('');
  const [quantity, setQuantity] = createSignal('1');
  const [selectedProductType, setSelectedProductType] = createSignal('tshirt');
  const [selectedShipRegion, setSelectedShipRegion] = createSignal('us');
  const [preferredPartner, setPreferredPartner] = createSignal('');
  const [manualPartnerOverride, setManualPartnerOverride] =
    createSignal(false);
  const [routingRecommendation, setRoutingRecommendation] =
    createSignal<RoutedOrderRecommendation | null>(null);
  const [selectedExceptionType, setSelectedExceptionType] =
    createSignal('artwork_issue');
  const [activeQueueView, setActiveQueueView] = createSignal<QueueView>('all');
  const [activeQueueSort, setActiveQueueSort] =
    createSignal<QueueSort>('priority');
  const [activityFilter, setActivityFilter] =
    createSignal<ActivityFilter>('notes');
  const [hideSystemActivity, setHideSystemActivity] = createSignal(true);
  const [operatorLens, setOperatorLens] = createSignal('');
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const currentStoreId = () => workspace?.currentStoreId() || '';
  const currentStore = () => workspace?.currentStore();
  const workspaceReady = () => !workspace || currentStoreId().trim().length > 0;
  const storeLabel = () =>
    currentStore()?.name || currentStoreId() || 'selected store';
  const effectivePreferredPartner = () =>
    manualPartnerOverride() ? preferredPartner().trim() : '';

  const applyPreferredPartnerOverride = (partnerName: string) => {
    setManualPartnerOverride(true);
    setPreferredPartner(partnerName);
  };

  const resetPreferredPartnerOverride = () => {
    setManualPartnerOverride(false);
    setPreferredPartner('');
  };

  const drafts = useOrderDrafts();
  const insights = useOrderInsights({
    orders,
    storeLabel,
    activeQueueView,
    activeQueueSort,
    operatorLens,
    activityFilter,
    hideSystemActivity,
    setMessage,
    setError,
  });
  const storage = useOrderStorage({
    tenantId: () => params().tenantId,
    storeId: currentStoreId,
    activeQueueView,
    setActiveQueueView,
    activeQueueSort,
    setActiveQueueSort,
    operatorLens,
    setOperatorLens,
    setMessage,
  });
  const actions = useOrderActions({
    orders,
    setOrders,
    availableCandidates,
    selectedCandidateId,
    customerName,
    setCustomerName,
    quantity,
    setQuantity,
    selectedProductType,
    selectedShipRegion,
    effectivePreferredPartner,
    resetPreferredPartnerOverride,
    selectedExceptionType,
    selectedOrderIDs: storage.selectedOrderIDs,
    setSelectedOrderIDs: storage.setSelectedOrderIDs,
    bulkDraft: storage.bulkDraft,
    setBulkDraft: storage.setBulkDraft,
    drafts,
    setMessage,
    setError,
  });

  const loadCandidates = async () => {
    const result = await getProductSetupSnapshot();
    if (!result.success) {
      setError(result.message);
      setAvailableCandidates([]);
      return;
    }
    const published = result.data.candidates.filter(
      (candidate) => candidate.status === 'published_mock'
    );
    setAvailableCandidates(published);
    if (!selectedCandidateId() && published.length > 0) {
      setSelectedCandidateId(published[0].id);
    }
  };

  const loadOrders = async () => {
    const result = await getRoutedOrders();
    if (!result.success) {
      setError(result.message);
      setOrders([]);
      return;
    }
    setOrders(result.data.orders);
    if (!operatorLens().trim()) {
      const firstAssigned = result.data.orders.find(
        (order) =>
          order.operatorAssignee && order.operatorAssignee !== 'unassigned'
      );
      if (firstAssigned) {
        setOperatorLens(firstAssigned.operatorAssignee);
      }
    }
  };

  const loadRoutingRecommendation = async () => {
    const candidateID = selectedCandidateId().trim();
    if (!candidateID) {
      setRoutingRecommendation(null);
      return;
    }
    const result = await getRoutedOrderRecommendation({
      candidateId: candidateID,
      productType: selectedProductType(),
      shipRegion: selectedShipRegion(),
      preferredPartner: preferredPartner().trim() || undefined,
    });
    if (!result.success) {
      setError(result.message);
      setRoutingRecommendation(null);
      return;
    }
    setRoutingRecommendation(result.data);
  };

  const composerContextValue: TenantOrdersComposerContextValue = {
    availableCandidates,
    selectedCandidateId,
    setSelectedCandidateId,
    customerName,
    setCustomerName,
    quantity,
    setQuantity,
    selectedProductType,
    setSelectedProductType,
    selectedShipRegion,
    setSelectedShipRegion,
    preferredPartner,
    setPreferredPartner,
    manualPartnerOverride,
    setManualPartnerOverride,
    routingRecommendation,
    selectedExceptionType,
    setSelectedExceptionType,
    applyPreferredPartnerOverride,
    resetPreferredPartnerOverride,
    createMockOrder: actions.createMockOrder,
  };

  const boardContextValue: TenantOrdersBoardContextValue = {
    activeQueueView,
    setActiveQueueView,
    activeQueueSort,
    setActiveQueueSort,
    operatorLens,
    setOperatorLens,
    queueViewCount: insights.queueViewCount,
    savedPresets: storage.savedPresets,
    presetName: storage.presetName,
    setPresetName: storage.setPresetName,
    saveQueuePreset: storage.saveQueuePreset,
    applyQueuePreset: storage.applyQueuePreset,
    deleteQueuePreset: storage.deleteQueuePreset,
    selectedOrderIDs: storage.selectedOrderIDs,
    selectVisibleOrders: () =>
      storage.setSelectedOrderIDs(insights.sortedOrders().map((order) => order.id)),
    clearSelectedOrders: storage.clearSelectedOrders,
    bulkDraft: storage.bulkDraft,
    setBulkDraft: storage.setBulkDraft,
    applyRelativeShipmentSla: storage.applyRelativeShipmentSla,
    savedBulkTemplates: storage.savedBulkTemplates,
    bulkTemplateName: storage.bulkTemplateName,
    setBulkTemplateName: storage.setBulkTemplateName,
    saveBulkTemplate: storage.saveBulkTemplate,
    applyBulkTemplate: storage.applyBulkTemplate,
    deleteBulkTemplate: storage.deleteBulkTemplate,
    applyBulkUpdate: actions.applyBulkUpdate,
  };

  const insightsContextValue: TenantOrdersInsightsContextValue = {
    tenantId: params().tenantId,
    storeId: currentStoreId,
    storeLabel,
    blockedOrders: insights.blockedOrders,
    blockedReasonSummary: insights.blockedReasonSummary,
    forcedRerouteSummary: insights.forcedRerouteSummary,
    reconciliationOrders: insights.reconciliationOrders,
    partnerFinanceSummary: insights.partnerFinanceSummary,
    storeActivityFeed: insights.storeActivityFeed,
    copyStoreActivityFeed: insights.copyStoreActivityFeed,
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    if (!workspaceReady()) {
      return;
    }
    storage.loadSavedPresets();
    storage.loadSavedBulkTemplates();
    void loadCandidates();
    void loadOrders();
  });

  createEffect(() => {
    const current = search();
    const queueView = String(current.queueView || '');
    const queueSort = String(current.queueSort || '');
    const lens = String(current.operatorLens || '');

    if (isQueueView(queueView)) {
      setActiveQueueView(queueView);
    }
    if (isQueueSort(queueSort)) {
      setActiveQueueSort(queueSort);
    }
    if (lens) {
      setOperatorLens(lens);
    }
  });

  createEffect(() => {
    if (!workspaceReady()) {
      setRoutingRecommendation(null);
      return;
    }
    void loadRoutingRecommendation();
  });

  return (
    <TenantOrdersView
      storeLabel={storeLabel}
      workspaceReady={workspaceReady}
      message={message}
      error={error}
      orders={orders}
      activityFilter={activityFilter}
      setActivityFilter={setActivityFilter}
      hideSystemActivity={hideSystemActivity}
      setHideSystemActivity={setHideSystemActivity}
      composerContextValue={composerContextValue}
      boardContextValue={boardContextValue}
      insightsContextValue={insightsContextValue}
      storage={storage}
      actions={actions}
      drafts={drafts}
      insights={insights}
    />
  );
}
