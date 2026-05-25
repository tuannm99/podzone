import { useParams, useSearch } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal } from 'solid-js';
import {
  advanceRoutedOrder,
  bulkUpdateRoutedOrders,
  createRoutedOrder,
  forceRerouteBlockedOrder,
  getRoutedOrderRecommendation,
  getRoutedOrders,
  openRoutedOrderException,
  type RoutedOrderRecommendation,
  type RoutedOrder,
  updateRoutedOrderExceptionStatus,
  updateRoutedOrderIssueHandling,
  updateRoutedOrderQueueControl,
  updateRoutedOrderSettlement,
  updateRoutedOrderShipment,
} from '../../../services/orders';
import {
  getProductSetupSnapshot,
  type CatalogCandidate,
} from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Card,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { TenantOrdersInsightsProvider } from './orders/context';
import { OrdersInsightsPanel } from './orders/OrdersInsightsPanel';
import { StoreActivityFeedPanel } from './orders/StoreActivityFeedPanel';
import {
  TenantOrdersComposerProvider,
} from './orders/composer-context';
import { CreateRoutedOrderPanel } from './orders/CreateRoutedOrderPanel';
import { TenantOrdersBoardProvider } from './orders/board-context';
import { QueueToolbarPanel } from './orders/QueueToolbarPanel';
import { BulkOpsPanel } from './orders/BulkOpsPanel';
import { OrderCard } from './orders/OrderCard';
import {
  anomalyFlagsFor,
  formatActivitySummary,
  formatStoreActivitySummary,
  hasFinanceAttention,
  parseMoneyValue,
  type PartnerFinanceSummaryItem,
  type StoreActivityFeedEntry,
} from './orders/utils';

const routeStatuses = [
  { name: 'Routing blocked', value: 'routing_blocked' },
  { name: 'Queued', value: 'queued' },
  { name: 'In production', value: 'in_production' },
  { name: 'Shipped', value: 'shipped' },
];

const shipmentOptions = [
  { name: 'Awaiting label', value: 'awaiting_label' },
  { name: 'Label ready', value: 'label_ready' },
  { name: 'In transit', value: 'in_transit' },
  { name: 'Delivered', value: 'delivered' },
  { name: 'Delivery issue', value: 'delivery_issue' },
];

const settlementOptions = [
  { name: 'Pending', value: 'pending' },
  { name: 'Reconciled', value: 'reconciled' },
  { name: 'Paid', value: 'paid' },
  { name: 'Disputed', value: 'disputed' },
];

const issueResolutionOptions = [
  { name: 'Monitor', value: 'monitor' },
  { name: 'Reprint', value: 'reprint' },
  { name: 'Refund', value: 'refund' },
  { name: 'Carrier claim', value: 'carrier_claim' },
  { name: 'Address retry', value: 'address_retry' },
];

const activityFilterOptions = [
  { name: 'All', value: 'all' },
  { name: 'Notes only', value: 'notes' },
  { name: 'System', value: 'system' },
  { name: 'Shipment', value: 'shipment_note' },
  { name: 'Settlement', value: 'settlement_note' },
  { name: 'Issue', value: 'issue_note' },
] as const;

type ShipmentDraft = {
  shipmentStatus: string;
  shipmentCarrier: string;
  shipmentTrackingNumber: string;
  shipmentTrackingUrl: string;
  shipmentNotes: string;
};

type SettlementDraft = {
  fulfillmentCost: string;
  shippingCost: string;
  settlementStatus: string;
  settlementNotes: string;
};

type IssueDraft = {
  issueCost: string;
  issueResolution: string;
  issueNotes: string;
};

type QueueDraft = {
  operatorAssignee: string;
  shipmentSlaDueAt: string;
  issueSlaDueAt: string;
};

type RerouteDraft = {
  preferredPartner: string;
};

type QueueView =
  | 'all'
  | 'my_queue'
  | 'overdue'
  | 'delivery_issues'
  | 'settlement_pending'
  | 'finance_review';

type QueueSort = 'priority' | 'newest';
type ActivityFilter = (typeof activityFilterOptions)[number]['value'];

type ShipmentSlaMode = '' | 'plus_2h' | 'plus_4h' | 'end_of_day';

type BulkDraft = {
  operatorAssignee: string;
  shipmentSlaDueAt: string;
  shipmentSlaMode: ShipmentSlaMode;
  settlementStatus: string;
};

type SavedQueuePreset = {
  name: string;
  queueView: QueueView;
  queueSort: QueueSort;
  operatorLens: string;
};

type SavedBulkTemplate = {
  name: string;
  operatorAssignee: string;
  shipmentSlaMode: ShipmentSlaMode;
  settlementStatus: string;
};

function statusColor(status: string) {
  switch (status) {
    case 'routing_blocked':
      return 'red';
    case 'shipped':
      return 'green';
    case 'in_production':
      return 'blue';
    default:
      return 'yellow';
  }
}

function exceptionColor(status: string) {
  switch (status) {
    case 'resolved':
      return 'green';
    case 'escalated':
      return 'red';
    case 'open':
      return 'yellow';
    default:
      return 'dark';
  }
}

function shipmentColor(status: string) {
  switch (status) {
    case 'delivered':
      return 'green';
    case 'in_transit':
      return 'blue';
    case 'delivery_issue':
      return 'red';
    case 'label_ready':
      return 'indigo';
    default:
      return 'dark';
  }
}

function settlementColor(status: string) {
  switch (status) {
    case 'paid':
      return 'green';
    case 'reconciled':
      return 'blue';
    case 'disputed':
      return 'red';
    default:
      return 'yellow';
  }
}

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


function toLocalDateTimeValue(value?: string) {
  if (!value) {
    return '';
  }
  const date = new Date(value);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function toIsoDateTime(value: string) {
  if (!value.trim()) {
    return '';
  }
  return new Date(value).toISOString();
}

function isOverdue(value?: string) {
  if (!value) {
    return false;
  }
  return new Date(value).getTime() < Date.now();
}

function isQueueView(value: string): value is QueueView {
  return (
    value === 'all' ||
    value === 'my_queue' ||
    value === 'overdue' ||
    value === 'delivery_issues' ||
    value === 'settlement_pending' ||
    value === 'finance_review'
  );
}

function isQueueSort(value: string): value is QueueSort {
  return value === 'priority' || value === 'newest';
}

function queuePresetStorageKey(tenantID: string) {
  return `podzone:orders:queue-presets:${tenantID}`;
}

function bulkTemplateStorageKey(tenantID: string) {
  return `podzone:orders:bulk-templates:${tenantID}`;
}

function resolveShipmentSla(mode: ShipmentSlaMode) {
  if (!mode) {
    return '';
  }
  const now = new Date();
  if (mode === 'plus_2h') {
    return new Date(now.getTime() + 2 * 60 * 60 * 1000).toISOString();
  }
  if (mode === 'plus_4h') {
    return new Date(now.getTime() + 4 * 60 * 60 * 1000).toISOString();
  }
  const endOfDay = new Date(now);
  endOfDay.setHours(23, 59, 0, 0);
  return endOfDay.toISOString();
}

export default function TenantOrdersPage() {
  const params = useParams({ from: '/t/$tenantId/orders' });
  const search = useSearch({ strict: false }) as () => Record<string, unknown>;

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
  const [manualPartnerOverride, setManualPartnerOverride] = createSignal(false);
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
  const [savedPresets, setSavedPresets] = createSignal<SavedQueuePreset[]>([]);
  const [presetName, setPresetName] = createSignal('');
  const [savedBulkTemplates, setSavedBulkTemplates] = createSignal<
    SavedBulkTemplate[]
  >([]);
  const [bulkTemplateName, setBulkTemplateName] = createSignal('');
  const [selectedOrderIDs, setSelectedOrderIDs] = createSignal<string[]>([]);
  const [bulkDraft, setBulkDraft] = createSignal<BulkDraft>({
    operatorAssignee: '',
    shipmentSlaDueAt: '',
    shipmentSlaMode: '',
    settlementStatus: '',
  });
  const [shipmentDrafts, setShipmentDrafts] = createSignal<
    Record<string, ShipmentDraft>
  >({});
  const [settlementDrafts, setSettlementDrafts] = createSignal<
    Record<string, SettlementDraft>
  >({});
  const [issueDrafts, setIssueDrafts] = createSignal<
    Record<string, IssueDraft>
  >({});
  const [queueDrafts, setQueueDrafts] = createSignal<
    Record<string, QueueDraft>
  >({});
  const [rerouteDrafts, setRerouteDrafts] = createSignal<
    Record<string, RerouteDraft>
  >({});
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');

  const shipmentDraftFor = (order: RoutedOrder): ShipmentDraft =>
    shipmentDrafts()[order.id] || {
      shipmentStatus: order.shipmentStatus || 'awaiting_label',
      shipmentCarrier: order.shipmentCarrier || '',
      shipmentTrackingNumber: order.shipmentTrackingNumber || '',
      shipmentTrackingUrl: order.shipmentTrackingUrl || '',
      shipmentNotes: order.shipmentNotes || '',
    };

  const settlementDraftFor = (order: RoutedOrder): SettlementDraft =>
    settlementDrafts()[order.id] || {
      fulfillmentCost:
        order.fulfillmentCost || order.baseCostSnapshot || '$0.00',
      shippingCost: order.shippingCost || '$0.00',
      settlementStatus: order.settlementStatus || 'pending',
      settlementNotes: order.settlementNotes || '',
    };

  const issueDraftFor = (order: RoutedOrder): IssueDraft =>
    issueDrafts()[order.id] || {
      issueCost: order.issueCost || '$0.00',
      issueResolution: order.issueResolution || 'monitor',
      issueNotes: order.issueNotes || '',
    };

  const queueDraftFor = (order: RoutedOrder): QueueDraft =>
    queueDrafts()[order.id] || {
      operatorAssignee: order.operatorAssignee || 'unassigned',
      shipmentSlaDueAt: toLocalDateTimeValue(order.shipmentSlaDueAt),
      issueSlaDueAt: toLocalDateTimeValue(order.issueSlaDueAt),
    };

  const rerouteDraftFor = (order: RoutedOrder): RerouteDraft =>
    rerouteDrafts()[order.id] || {
      preferredPartner: order.partner || '',
    };

  const patchShipmentDraft = (
    orderId: string,
    patch: Partial<ShipmentDraft>
  ) => {
    setShipmentDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          shipmentStatus: 'awaiting_label',
          shipmentCarrier: '',
          shipmentTrackingNumber: '',
          shipmentTrackingUrl: '',
          shipmentNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const patchSettlementDraft = (
    orderId: string,
    patch: Partial<SettlementDraft>
  ) => {
    setSettlementDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          fulfillmentCost: '$0.00',
          shippingCost: '$0.00',
          settlementStatus: 'pending',
          settlementNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const patchIssueDraft = (orderId: string, patch: Partial<IssueDraft>) => {
    setIssueDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          issueCost: '$0.00',
          issueResolution: 'monitor',
          issueNotes: '',
        }),
        ...patch,
      },
    }));
  };

  const patchQueueDraft = (orderId: string, patch: Partial<QueueDraft>) => {
    setQueueDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          operatorAssignee: 'unassigned',
          shipmentSlaDueAt: '',
          issueSlaDueAt: '',
        }),
        ...patch,
      },
    }));
  };

  const patchRerouteDraft = (
    orderId: string,
    patch: Partial<RerouteDraft>
  ) => {
    setRerouteDrafts((current) => ({
      ...current,
      [orderId]: {
        ...(current[orderId] || {
          preferredPartner: '',
        }),
        ...patch,
      },
    }));
  };

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

  const matchesQueueView = (order: RoutedOrder, view: QueueView) => {
    switch (view) {
      case 'my_queue':
        return (
          !!operatorLens().trim() &&
          order.operatorAssignee.toLowerCase() ===
            operatorLens().trim().toLowerCase()
        );
      case 'overdue':
        return (
          (order.shipmentSlaDueAt &&
            isOverdue(order.shipmentSlaDueAt) &&
            order.shipmentStatus !== 'delivered') ||
          (order.issueSlaDueAt &&
            isOverdue(order.issueSlaDueAt) &&
            (order.exceptionStatus === 'open' ||
              order.exceptionStatus === 'escalated' ||
              order.shipmentStatus === 'delivery_issue'))
        );
      case 'delivery_issues':
        return (
          order.shipmentStatus === 'delivery_issue' ||
          order.issueResolution === 'carrier_claim'
        );
      case 'settlement_pending':
        return (
          order.settlementStatus === 'pending' ||
          order.settlementStatus === 'disputed'
        );
      case 'finance_review':
        return hasFinanceAttention(order);
      default:
        return true;
    }
  };

  const filteredOrders = () =>
    orders().filter((order) => matchesQueueView(order, activeQueueView()));

  const priorityScoreFor = (order: RoutedOrder) => {
    const shipmentOverdue =
      !!order.shipmentSlaDueAt &&
      isOverdue(order.shipmentSlaDueAt) &&
      order.shipmentStatus !== 'delivered';
    const issueOverdue =
      !!order.issueSlaDueAt &&
      isOverdue(order.issueSlaDueAt) &&
      (order.exceptionStatus === 'open' ||
        order.exceptionStatus === 'escalated' ||
        order.shipmentStatus === 'delivery_issue');

    if (shipmentOverdue || issueOverdue) {
      return 0;
    }
    if (order.status === 'routing_blocked') {
      return 1;
    }
    if (order.shipmentStatus === 'delivery_issue') {
      return 2;
    }
    if (order.settlementStatus === 'disputed') {
      return 3;
    }
    if (
      order.exceptionStatus === 'open' ||
      order.exceptionStatus === 'escalated'
    ) {
      return 4;
    }
    if (order.status === 'in_production') {
      return 5;
    }
    if (order.settlementStatus === 'pending') {
      return 6;
    }
    return 7;
  };

  const sortedOrders = () => {
    const ranked = [...filteredOrders()];
    if (activeQueueSort() === 'newest') {
      return ranked.sort(
        (a, b) =>
          new Date(b.createdAt || 0).getTime() -
          new Date(a.createdAt || 0).getTime()
      );
    }
    return ranked.sort((a, b) => {
      const priorityDelta = priorityScoreFor(a) - priorityScoreFor(b);
      if (priorityDelta !== 0) {
        return priorityDelta;
      }
      return (
        new Date(b.createdAt || 0).getTime() -
        new Date(a.createdAt || 0).getTime()
      );
    });
  };

  const queueViewCount = (view: QueueView) =>
    orders().filter((order) => matchesQueueView(order, view)).length;

  const blockedOrders = () =>
    orders().filter((order) => order.status === 'routing_blocked');

  const blockedReasonSummary = () => {
    const counts = new Map<string, number>();
    for (const order of blockedOrders()) {
      const key = order.routingBlockCode || 'unknown';
      counts.set(key, (counts.get(key) || 0) + 1);
    }
    return [...counts.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
      .map(([code, count]) => ({ code, count }));
  };

  const forcedRerouteSummary = () => {
    const counts = new Map<string, number>();
    for (const order of orders()) {
      for (const activity of order.activityLog) {
        const manualReroute = activity.details.some(
          (detail) =>
            detail.key === 'manual_reroute' && detail.value === 'true'
        );
        if (!manualReroute) {
          continue;
        }
        const partner =
          activity.details.find((detail) => detail.key === 'partner')?.value ||
          order.partner ||
          'unknown';
        counts.set(partner, (counts.get(partner) || 0) + 1);
      }
    }
    return [...counts.entries()]
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
      .map(([partner, count]) => ({ partner, count }));
  };

  const reconciliationOrders = () =>
    [...orders()]
      .filter(hasFinanceAttention)
      .sort((a, b) => {
        const aDisputed = a.settlementStatus === 'disputed' ? 0 : 1;
        const bDisputed = b.settlementStatus === 'disputed' ? 0 : 1;
        if (aDisputed !== bDisputed) {
          return aDisputed - bDisputed;
        }
        const aAnomaly = anomalyFlagsFor(a).length;
        const bAnomaly = anomalyFlagsFor(b).length;
        if (aAnomaly !== bAnomaly) {
          return bAnomaly - aAnomaly;
        }
        return (
          new Date(a.createdAt || 0).getTime() -
          new Date(b.createdAt || 0).getTime()
        );
      });

  const partnerFinanceSummary = () => {
    const summary = new Map<string, PartnerFinanceSummaryItem>();

    for (const order of orders()) {
      const partner = order.partner || 'partner pending';
      const current = summary.get(partner) || {
        partner,
        orders: 0,
        pending: 0,
        disputed: 0,
        paid: 0,
        blocked: 0,
        forcedReroutes: 0,
        realizedMargin: 0,
      };
      current.orders += 1;
      if (order.settlementStatus === 'pending') {
        current.pending += 1;
      }
      if (order.settlementStatus === 'disputed') {
        current.disputed += 1;
      }
      if (order.settlementStatus === 'paid') {
        current.paid += 1;
      }
      if (order.status === 'routing_blocked') {
        current.blocked += 1;
      }
      const margin = parseMoneyValue(order.realizedMargin);
      if (margin !== null) {
        current.realizedMargin += margin;
      }
      for (const activity of order.activityLog) {
        const manualReroute = activity.details.some(
          (detail) =>
            detail.key === 'manual_reroute' && detail.value === 'true'
        );
        if (manualReroute) {
          current.forcedReroutes += 1;
        }
      }
      summary.set(partner, current);
    }

    return [...summary.values()].sort((a, b) => {
      if (b.disputed !== a.disputed) {
        return b.disputed - a.disputed;
      }
      if (b.pending !== a.pending) {
        return b.pending - a.pending;
      }
      return a.partner.localeCompare(b.partner);
    });
  };

  const matchesActivityFilter = (
    activity: RoutedOrder['activityLog'][number]
  ) => {
    const selectedFilter = activityFilter();
    if (selectedFilter === 'notes') {
      return activity.type !== 'system';
    }
    if (selectedFilter !== 'all' && activity.type !== selectedFilter) {
      return false;
    }
    if (
      hideSystemActivity() &&
      selectedFilter === 'all' &&
      activity.type === 'system'
    ) {
      return false;
    }
    return true;
  };

  const filteredActivityLogFor = (order: RoutedOrder) => {
    return order.activityLog
      .filter(matchesActivityFilter)
      .slice()
      .reverse()
      .slice(0, 8);
  };

  const hiddenSystemActivityCountFor = (order: RoutedOrder) => {
    if (activityFilter() !== 'all' || !hideSystemActivity()) {
      return 0;
    }
    return order.activityLog.filter((activity) => activity.type === 'system')
      .length;
  };

  const storeActivityFeed = () =>
    sortedOrders()
      .flatMap((order) =>
        order.activityLog.filter(matchesActivityFilter).map((activity) => ({
          orderId: order.id,
          productTitle: order.productTitle,
          operatorAssignee: order.operatorAssignee,
          activity,
        }) satisfies StoreActivityFeedEntry)
      )
      .sort(
        (a, b) =>
          new Date(b.activity.createdAt).getTime() -
          new Date(a.activity.createdAt).getTime()
      )
      .slice(0, 14);

  const copyActivitySummary = async (order: RoutedOrder) => {
    const summary = formatActivitySummary(order, filteredActivityLogFor(order));
    try {
      await navigator.clipboard.writeText(summary);
      setMessage(`Copied activity summary for ${order.id}.`);
    } catch {
      setError('Could not copy activity summary to clipboard.');
    }
  };

  const copyStoreActivityFeed = async () => {
    const summary = formatStoreActivitySummary(
      params().tenantId,
      storeActivityFeed()
    );
    try {
      await navigator.clipboard.writeText(summary);
      setMessage(`Copied store activity feed for ${params().tenantId}.`);
    } catch {
      setError('Could not copy store activity feed to clipboard.');
    }
  };

  const insightsContextValue = {
    tenantId: params().tenantId,
    blockedOrders,
    blockedReasonSummary,
    forcedRerouteSummary,
    reconciliationOrders,
    partnerFinanceSummary,
    storeActivityFeed,
    copyStoreActivityFeed,
  };

  const loadSavedPresets = () => {
    const raw = window.localStorage.getItem(
      queuePresetStorageKey(params().tenantId)
    );
    if (!raw) {
      setSavedPresets([]);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as SavedQueuePreset[];
      setSavedPresets(Array.isArray(parsed) ? parsed : []);
    } catch {
      setSavedPresets([]);
    }
  };

  const persistSavedPresets = (next: SavedQueuePreset[]) => {
    window.localStorage.setItem(
      queuePresetStorageKey(params().tenantId),
      JSON.stringify(next)
    );
    setSavedPresets(next);
  };

  const loadSavedBulkTemplates = () => {
    const raw = window.localStorage.getItem(
      bulkTemplateStorageKey(params().tenantId)
    );
    if (!raw) {
      setSavedBulkTemplates([]);
      return;
    }
    try {
      const parsed = JSON.parse(raw) as SavedBulkTemplate[];
      setSavedBulkTemplates(Array.isArray(parsed) ? parsed : []);
    } catch {
      setSavedBulkTemplates([]);
    }
  };

  const persistSavedBulkTemplates = (next: SavedBulkTemplate[]) => {
    window.localStorage.setItem(
      bulkTemplateStorageKey(params().tenantId),
      JSON.stringify(next)
    );
    setSavedBulkTemplates(next);
  };

  const saveQueuePreset = () => {
    const name = presetName().trim();
    if (!name) {
      setMessage('Enter a preset name first.');
      return;
    }
    const nextPreset: SavedQueuePreset = {
      name,
      queueView: activeQueueView(),
      queueSort: activeQueueSort(),
      operatorLens: operatorLens().trim(),
    };
    const deduped = savedPresets().filter((preset) => preset.name !== name);
    persistSavedPresets([nextPreset, ...deduped]);
    setPresetName('');
    setMessage(`Saved queue preset ${name}.`);
  };

  const applyQueuePreset = (preset: SavedQueuePreset) => {
    setActiveQueueView(preset.queueView);
    setActiveQueueSort(preset.queueSort);
    setOperatorLens(preset.operatorLens);
    setMessage(`Applied queue preset ${preset.name}.`);
  };

  const deleteQueuePreset = (name: string) => {
    persistSavedPresets(
      savedPresets().filter((preset) => preset.name !== name)
    );
    setMessage(`Deleted queue preset ${name}.`);
  };

  const saveBulkTemplate = () => {
    const name = bulkTemplateName().trim();
    if (!name) {
      setMessage('Enter a bulk template name first.');
      return;
    }
    const draft = bulkDraft();
    const nextTemplate: SavedBulkTemplate = {
      name,
      operatorAssignee: draft.operatorAssignee.trim(),
      shipmentSlaMode: draft.shipmentSlaMode,
      settlementStatus: draft.settlementStatus.trim(),
    };
    const deduped = savedBulkTemplates().filter((item) => item.name !== name);
    persistSavedBulkTemplates([nextTemplate, ...deduped]);
    setBulkTemplateName('');
    setMessage(`Saved bulk template ${name}.`);
  };

  const applyBulkTemplate = (template: SavedBulkTemplate) => {
    setBulkDraft({
      operatorAssignee: template.operatorAssignee,
      shipmentSlaMode: template.shipmentSlaMode,
      shipmentSlaDueAt: template.shipmentSlaMode
        ? toLocalDateTimeValue(resolveShipmentSla(template.shipmentSlaMode))
        : '',
      settlementStatus: template.settlementStatus,
    });
    setMessage(`Loaded bulk template ${template.name}.`);
  };

  const deleteBulkTemplate = (name: string) => {
    persistSavedBulkTemplates(
      savedBulkTemplates().filter((template) => template.name !== name)
    );
    setMessage(`Deleted bulk template ${name}.`);
  };

  const toggleOrderSelection = (orderID: string, checked: boolean) => {
    setSelectedOrderIDs((current) => {
      if (checked) {
        return current.includes(orderID) ? current : [...current, orderID];
      }
      return current.filter((id) => id !== orderID);
    });
  };

  const selectVisibleOrders = () => {
    setSelectedOrderIDs(sortedOrders().map((order) => order.id));
  };

  const clearSelectedOrders = () => {
    setSelectedOrderIDs([]);
  };

  const isSelected = (orderID: string) => selectedOrderIDs().includes(orderID);

  const createMockOrder = async (event: SubmitEvent) => {
    event.preventDefault();
    const candidate = availableCandidates().find(
      (item) => item.id === selectedCandidateId()
    );
    if (!candidate) {
      setMessage(
        'Publish a mock product candidate first before routing orders.'
      );
      return;
    }

    setError('');
    const result = await createRoutedOrder({
      candidateId: candidate.id,
      customerName: customerName().trim(),
      quantity: Math.max(1, Number.parseInt(quantity(), 10) || 1),
      productType: selectedProductType(),
      shipRegion: selectedShipRegion(),
      preferredPartner: effectivePreferredPartner() || undefined,
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) => [result.data, ...current]);
    setCustomerName('');
    setQuantity('1');
    resetPreferredPartnerOverride();
    setMessage(`Created routed order ${result.data.id}.`);
  };

  const composerContextValue = {
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
    createMockOrder,
  };

  const advanceOrder = async (orderId: string) => {
    setError('');
    const result = await advanceRoutedOrder(orderId);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(`Advanced order ${orderId} to the next routing stage.`);
  };

  const raiseException = async (orderId: string) => {
    setError('');
    const result = await openRoutedOrderException(
      orderId,
      selectedExceptionType()
    );
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(
      `Raised ${selectedExceptionType().replaceAll('_', ' ')} on ${orderId}.`
    );
  };

  const updateExceptionStatus = async (orderId: string, nextStatus: string) => {
    setError('');
    const result = await updateRoutedOrderExceptionStatus(orderId, nextStatus);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((order) => (order.id === orderId ? result.data : order))
    );
    setMessage(`Updated exception on ${orderId} to ${nextStatus}.`);
  };

  const saveShipment = async (order: RoutedOrder) => {
    setError('');
    const draft = shipmentDraftFor(order);
    const result = await updateRoutedOrderShipment(order.id, {
      shipmentStatus: draft.shipmentStatus,
      carrier: draft.shipmentCarrier.trim(),
      trackingNumber: draft.shipmentTrackingNumber.trim(),
      trackingUrl: draft.shipmentTrackingUrl.trim(),
      notes: draft.shipmentNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setShipmentDrafts((current) => ({
      ...current,
      [order.id]: {
        shipmentStatus: result.data.shipmentStatus,
        shipmentCarrier: result.data.shipmentCarrier,
        shipmentTrackingNumber: result.data.shipmentTrackingNumber,
        shipmentTrackingUrl: result.data.shipmentTrackingUrl,
        shipmentNotes: result.data.shipmentNotes,
      },
    }));
    setMessage(`Updated manual shipment control on ${order.id}.`);
  };

  const saveSettlement = async (order: RoutedOrder) => {
    setError('');
    const draft = settlementDraftFor(order);
    const result = await updateRoutedOrderSettlement(order.id, {
      fulfillmentCost: draft.fulfillmentCost.trim(),
      shippingCost: draft.shippingCost.trim(),
      settlementStatus: draft.settlementStatus,
      notes: draft.settlementNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setSettlementDrafts((current) => ({
      ...current,
      [order.id]: {
        fulfillmentCost: result.data.fulfillmentCost,
        shippingCost: result.data.shippingCost,
        settlementStatus: result.data.settlementStatus,
        settlementNotes: result.data.settlementNotes,
      },
    }));
    setMessage(`Updated settlement readiness on ${order.id}.`);
  };

  const saveIssueHandling = async (order: RoutedOrder) => {
    setError('');
    const draft = issueDraftFor(order);
    const result = await updateRoutedOrderIssueHandling(order.id, {
      issueCost: draft.issueCost.trim(),
      issueResolution: draft.issueResolution,
      notes: draft.issueNotes.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setIssueDrafts((current) => ({
      ...current,
      [order.id]: {
        issueCost: result.data.issueCost,
        issueResolution: result.data.issueResolution,
        issueNotes: result.data.issueNotes,
      },
    }));
    setMessage(`Updated issue cost handling on ${order.id}.`);
  };

  const saveQueueControl = async (order: RoutedOrder) => {
    setError('');
    const draft = queueDraftFor(order);
    const result = await updateRoutedOrderQueueControl(order.id, {
      operatorAssignee: draft.operatorAssignee.trim() || 'unassigned',
      shipmentSlaDueAt: toIsoDateTime(draft.shipmentSlaDueAt),
      issueSlaDueAt: toIsoDateTime(draft.issueSlaDueAt),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setQueueDrafts((current) => ({
      ...current,
      [order.id]: {
        operatorAssignee: result.data.operatorAssignee,
        shipmentSlaDueAt: toLocalDateTimeValue(result.data.shipmentSlaDueAt),
        issueSlaDueAt: toLocalDateTimeValue(result.data.issueSlaDueAt),
      },
    }));
    setMessage(`Updated queue ownership on ${order.id}.`);
  };

  const rerouteBlockedOrder = async (order: RoutedOrder) => {
    setError('');
    const preferredPartner = rerouteDraftFor(order).preferredPartner.trim();
    if (!preferredPartner) {
      setMessage('Choose a partner before rerouting a blocked order.');
      return;
    }
    const result = await forceRerouteBlockedOrder({
      orderId: order.id,
      preferredPartner,
    });
    if (!result.success) {
      setError(result.message);
      return;
    }
    setOrders((current) =>
      current.map((item) => (item.id === order.id ? result.data : item))
    );
    setRerouteDrafts((current) => ({
      ...current,
      [order.id]: {
        preferredPartner: result.data.partner,
      },
    }));
    setMessage(`Forced reroute for ${order.id} to ${result.data.partner}.`);
  };

  const applyBulkUpdate = async () => {
    setError('');
    if (selectedOrderIDs().length === 0) {
      setMessage('Select at least one routed order first.');
      return;
    }
    const draft = bulkDraft();
    if (
      !draft.operatorAssignee.trim() &&
      !draft.shipmentSlaDueAt.trim() &&
      !draft.settlementStatus.trim()
    ) {
      setMessage('Choose at least one bulk field before applying.');
      return;
    }

    const result = await bulkUpdateRoutedOrders({
      orderIds: selectedOrderIDs(),
      operatorAssignee: draft.operatorAssignee.trim(),
      shipmentSlaDueAt:
        resolveShipmentSla(draft.shipmentSlaMode) ||
        toIsoDateTime(draft.shipmentSlaDueAt),
      settlementStatus: draft.settlementStatus.trim(),
    });
    if (!result.success) {
      setError(result.message);
      return;
    }

    const byID = new Map(result.data.map((order) => [order.id, order]));
    setOrders((current) => current.map((order) => byID.get(order.id) || order));
    setMessage(`Applied bulk update to ${result.data.length} routed orders.`);
    setSelectedOrderIDs([]);
    setBulkDraft({
      operatorAssignee: '',
      shipmentSlaDueAt: '',
      shipmentSlaMode: '',
      settlementStatus: '',
    });
  };

  const applyRelativeShipmentSla = (mode: ShipmentSlaMode) => {
    setBulkDraft((current) => ({
      ...current,
      shipmentSlaMode: mode,
      shipmentSlaDueAt: mode
        ? toLocalDateTimeValue(resolveShipmentSla(mode))
        : current.shipmentSlaDueAt,
    }));
  };

  const boardContextValue = {
    activeQueueView,
    setActiveQueueView,
    activeQueueSort,
    setActiveQueueSort,
    operatorLens,
    setOperatorLens,
    queueViewCount,
    savedPresets,
    presetName,
    setPresetName,
    saveQueuePreset,
    applyQueuePreset,
    deleteQueuePreset,
    selectedOrderIDs,
    selectVisibleOrders,
    clearSelectedOrders,
    bulkDraft,
    setBulkDraft,
    applyRelativeShipmentSla,
    savedBulkTemplates,
    bulkTemplateName,
    setBulkTemplateName,
    saveBulkTemplate,
    applyBulkTemplate,
    deleteBulkTemplate,
    applyBulkUpdate,
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    loadSavedPresets();
    loadSavedBulkTemplates();
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
    void loadRoutingRecommendation();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Order Routing Workspace"
          title={`POD routing board for store ${params().tenantId}`}
          copy="This routing workspace persists store-scoped POD orders in the backend. Published mock products can be routed through production, shipment, issue handling, and settlement readiness."
        />
      </Card>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <InfoAlert>
        Orders and published product candidates now come from backend store
        data. Shipment and settlement control both stay manual on this board so
        the store team can manage POD execution directly.
      </InfoAlert>

      <div class="grid gap-6 lg:grid-cols-[0.96fr_1.04fr]">
        <Card class="space-y-4">
          <TenantOrdersComposerProvider value={composerContextValue}>
            <CreateRoutedOrderPanel />
          </TenantOrdersComposerProvider>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Routing board"
            subtitle="Watch each order move from intake to production, then manage shipment and settlement state directly inside the store-scoped POD workflow."
          />

          <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <TenantOrdersBoardProvider value={boardContextValue}>
              <QueueToolbarPanel />
              <BulkOpsPanel />
            </TenantOrdersBoardProvider>
            <TenantOrdersInsightsProvider value={insightsContextValue}>
              <OrdersInsightsPanel />
            </TenantOrdersInsightsProvider>
          </div>

          <Show
            when={sortedOrders().length > 0}
            fallback={
              <EmptyBlock
                title={
                  orders().length > 0
                    ? 'No orders in this queue view'
                    : 'No routed orders yet'
                }
                copy={
                  orders().length > 0
                    ? 'Adjust the queue view or operator lens to inspect a different operational slice.'
                    : 'Create a routed order on the left to test store-side routing, manual shipment control, and settlement readiness.'
                }
              />
            }
          >
            <div class="space-y-3">
              <TenantOrdersInsightsProvider value={insightsContextValue}>
                <StoreActivityFeedPanel />
              </TenantOrdersInsightsProvider>
              <For each={sortedOrders()}>
                {(order) => (
                  <OrderCard
                    order={order}
                    selected={isSelected(order.id)}
                    actions={{
                      toggleSelected: (checked) =>
                        toggleOrderSelection(order.id, checked),
                      advanceOrder,
                      raiseException,
                      updateExceptionStatus,
                      rerouteBlockedOrder,
                      saveQueueControl,
                      saveIssueHandling,
                      saveSettlement,
                      saveShipment,
                      copyActivitySummary,
                      queueDraftFor,
                      patchQueueDraft,
                      issueDraftFor,
                      patchIssueDraft,
                      settlementDraftFor,
                      patchSettlementDraft,
                      shipmentDraftFor,
                      patchShipmentDraft,
                      rerouteDraftFor,
                      patchRerouteDraft,
                    }}
                    helpers={{
                      queueSort: activeQueueSort(),
                      priorityScoreFor,
                      statusColor,
                      exceptionColor,
                      shipmentColor,
                      settlementColor,
                      activityColor,
                      isOverdue,
                      filteredActivityLogFor,
                      hiddenSystemActivityCountFor,
                    }}
                    ui={{
                      activityFilter: activityFilter(),
                      setActivityFilter,
                      hideSystemActivity: hideSystemActivity(),
                      toggleHideSystemActivity: () =>
                        setHideSystemActivity((current) => !current),
                      activityFilterOptions: activityFilterOptions.map(
                        (option) => ({
                          name: option.name,
                          value: option.value,
                        })
                      ),
                      shipmentOptions,
                      settlementOptions,
                      issueResolutionOptions,
                    }}
                  />
                )}
              </For>
            </div>
          </Show>
        </Card>
      </div>

      <Card class="mt-6 space-y-4">
        <SectionTitle
          title="Routing, shipment, and settlement stages"
          subtitle="Production routing, queue ownership, shipment control, and settlement updates are all managed inside the workspace, so operators can run POD execution manually without relying on external fulfillment callbacks."
        />
        <div class="flex flex-wrap gap-2">
          <For each={routeStatuses}>
            {(stage) => (
              <Badge content={stage.name} color={statusColor(stage.value)} />
            )}
          </For>
          <Badge content="Open issue" color="yellow" />
          <Badge content="Escalated issue" color="red" />
          <Badge content="Resolved issue" color="green" />
          <For each={shipmentOptions}>
            {(stage) => (
              <Badge content={stage.name} color={shipmentColor(stage.value)} />
            )}
          </For>
          <For each={settlementOptions}>
            {(stage) => (
              <Badge
                content={`Settlement ${stage.name}`}
                color={settlementColor(stage.value)}
              />
            )}
          </For>
          <For each={issueResolutionOptions}>
            {(stage) => (
              <Badge
                content={`Issue ${stage.name}`}
                color={
                  stage.value === 'reprint' || stage.value === 'refund'
                    ? 'red'
                    : 'yellow'
                }
              />
            )}
          </For>
        </div>
      </Card>
    </PageShell>
  );
}
