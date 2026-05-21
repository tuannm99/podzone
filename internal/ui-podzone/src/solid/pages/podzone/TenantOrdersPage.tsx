import { useParams, useSearch } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal } from 'solid-js';
import {
  advanceRoutedOrder,
  bulkUpdateRoutedOrders,
  createRoutedOrder,
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
  Button,
  Card,
  InputField,
  SelectField,
  TextareaField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const routeStatuses = [
  { name: 'Queued', value: 'queued' },
  { name: 'In production', value: 'in_production' },
  { name: 'Shipped', value: 'shipped' },
];

const exceptionOptions = [
  { name: 'Artwork issue', value: 'artwork_issue' },
  { name: 'Partner delay', value: 'partner_delay' },
  { name: 'Address hold', value: 'address_hold' },
  { name: 'Reprint request', value: 'reprint_request' },
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

const productTypeOptions = [
  { name: 'T-shirt', value: 'tshirt' },
  { name: 'Hoodie', value: 'hoodie' },
  { name: 'Tote', value: 'tote' },
  { name: 'Poster', value: 'poster' },
];

const shipRegionOptions = [
  { name: 'US', value: 'us' },
  { name: 'EU', value: 'eu' },
  { name: 'UK', value: 'uk' },
  { name: 'SEA', value: 'sea' },
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

type QueueView =
  | 'all'
  | 'my_queue'
  | 'overdue'
  | 'delivery_issues'
  | 'settlement_pending';

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

function joinPartnerCapabilityList(items: string[]) {
  return items.length > 0 ? items.join(', ') : 'Any';
}

function joinShippingCostRules(
  items: { region: string; cost: string }[]
) {
  return items.length > 0
    ? items.map((item) => `${item.region}:${item.cost}`).join(', ')
    : 'No region rules';
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

function formatActivityTime(value: string) {
  return new Date(value).toLocaleString();
}

function formatActivityActor(actor: string) {
  const normalized = actor.trim();
  if (!normalized) {
    return 'system';
  }
  return normalized;
}

function formatActivitySummary(order: RoutedOrder, activities: RoutedOrder['activityLog']) {
  const header = [
    `Order ${order.id}`,
    `Product: ${order.productTitle}`,
    `Operator: ${order.operatorAssignee || 'unassigned'}`,
    `Route status: ${order.status}`,
    `Shipment: ${order.shipmentStatus}`,
    `Settlement: ${order.settlementStatus}`,
  ];

  const activityLines = activities.map((activity) => {
    const details = activity.details
      .map((detail) => `${detail.key}=${detail.value}`)
      .join(', ');
    return [
      `[${formatActivityTime(activity.createdAt)}]`,
      activity.type,
      `by ${formatActivityActor(activity.actor)}`,
      activity.message,
      details ? `(${details})` : '',
    ]
      .filter(Boolean)
      .join(' ');
  });

  return [...header, '', 'Recent activity:', ...activityLines].join('\n');
}

function formatStoreActivitySummary(
  tenantId: string,
  entries: {
    orderId: string;
    productTitle: string;
    operatorAssignee: string;
    activity: RoutedOrder['activityLog'][number];
  }[]
) {
  const lines = entries.map((entry) => {
    const details = entry.activity.details
      .map((detail) => `${detail.key}=${detail.value}`)
      .join(', ');
    return [
      `[${formatActivityTime(entry.activity.createdAt)}]`,
      entry.orderId,
      `(${entry.productTitle})`,
      `owner ${entry.operatorAssignee || 'unassigned'}`,
      entry.activity.type,
      `by ${formatActivityActor(entry.activity.actor)}`,
      entry.activity.message,
      details ? `(${details})` : '',
    ]
      .filter(Boolean)
      .join(' ');
  });

  return [
    `Store activity feed for ${tenantId}`,
    '',
    ...lines,
  ].join('\n');
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
    value === 'settlement_pending'
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
  const [activeQueueSort, setActiveQueueSort] = createSignal<QueueSort>('priority');
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
  const [issueDrafts, setIssueDrafts] = createSignal<Record<string, IssueDraft>>(
    {}
  );
  const [queueDrafts, setQueueDrafts] = createSignal<Record<string, QueueDraft>>(
    {}
  );
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
      fulfillmentCost: order.fulfillmentCost || order.baseCostSnapshot || '$0.00',
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

  const patchShipmentDraft = (orderId: string, patch: Partial<ShipmentDraft>) => {
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
          order.operatorAssignee &&
          order.operatorAssignee !== 'unassigned'
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
    if (order.shipmentStatus === 'delivery_issue') {
      return 1;
    }
    if (order.settlementStatus === 'disputed') {
      return 2;
    }
    if (
      order.exceptionStatus === 'open' ||
      order.exceptionStatus === 'escalated'
    ) {
      return 3;
    }
    if (order.status === 'in_production') {
      return 4;
    }
    if (order.settlementStatus === 'pending') {
      return 5;
    }
    return 6;
  };

  const sortedOrders = () => {
    const ranked = [...filteredOrders()];
    if (activeQueueSort() === 'newest') {
      return ranked.sort(
        (a, b) =>
          new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime()
      );
    }
    return ranked.sort((a, b) => {
      const priorityDelta = priorityScoreFor(a) - priorityScoreFor(b);
      if (priorityDelta !== 0) {
        return priorityDelta;
      }
      return new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime();
    });
  };

  const queueViewCount = (view: QueueView) =>
    orders().filter((order) => matchesQueueView(order, view)).length;

  const matchesActivityFilter = (activity: RoutedOrder['activityLog'][number]) => {
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
        order.activityLog
          .filter(matchesActivityFilter)
          .map((activity) => ({
            orderId: order.id,
            productTitle: order.productTitle,
            operatorAssignee: order.operatorAssignee,
            activity,
          }))
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
    const summary = formatStoreActivitySummary(params().tenantId, storeActivityFeed());
    try {
      await navigator.clipboard.writeText(summary);
      setMessage(`Copied store activity feed for ${params().tenantId}.`);
    } catch {
      setError('Could not copy store activity feed to clipboard.');
    }
  };

  const loadSavedPresets = () => {
    const raw = window.localStorage.getItem(queuePresetStorageKey(params().tenantId));
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
    persistSavedPresets(savedPresets().filter((preset) => preset.name !== name));
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
      setMessage('Publish a mock product candidate first before routing orders.');
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
    setMessage(`Raised ${selectedExceptionType().replaceAll('_', ' ')} on ${orderId}.`);
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
    setOrders((current) =>
      current.map((order) => byID.get(order.id) || order)
    );
    setMessage(`Applied bulk update to ${result.data.length} routed orders.`);
    setSelectedOrderIDs([]);
    setBulkDraft({
      operatorAssignee: '',
      shipmentSlaDueAt: '',
      shipmentSlaMode: '',
      settlementStatus: '',
    });
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
        Orders and published product candidates now come from backend store data. Shipment and settlement control both stay manual on this board so the store team can manage POD execution directly.
      </InfoAlert>

      <div class="grid gap-6 lg:grid-cols-[0.96fr_1.04fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create routed order"
            subtitle="Use a published mock product candidate as the source, then send the order into the backend-backed POD routing flow."
          />

          <Show
            when={availableCandidates().length > 0}
            fallback={
              <EmptyBlock
                title="No published mock products yet"
                copy="Go to Product setup, promote a draft, and mock publish it from the backend-backed setup workflow before testing order routing."
              />
            }
          >
            <form class="space-y-4" onSubmit={createMockOrder}>
              <SelectField
                label="Published mock product"
                value={selectedCandidateId()}
                options={availableCandidates().map((candidate) => ({
                  name: `${candidate.title} · ${candidate.partner}`,
                  value: candidate.id,
                }))}
                onChange={(event) =>
                  setSelectedCandidateId(event.currentTarget.value)
                }
              />
              <div class="grid gap-4 md:grid-cols-2">
                <InputField
                  label="Customer name"
                  value={customerName()}
                  placeholder="Nguyen Minh"
                  onInput={(event) => setCustomerName(event.currentTarget.value)}
                />
                <InputField
                  label="Quantity"
                  value={quantity()}
                  placeholder="1"
                  onInput={(event) => setQuantity(event.currentTarget.value)}
                />
              </div>
              <div class="grid gap-4 md:grid-cols-3">
                <SelectField
                  label="Product type"
                  value={selectedProductType()}
                  options={productTypeOptions}
                  onChange={(event) =>
                    setSelectedProductType(event.currentTarget.value)
                  }
                />
                <SelectField
                  label="Ship region"
                  value={selectedShipRegion()}
                  options={shipRegionOptions}
                  onChange={(event) =>
                    setSelectedShipRegion(event.currentTarget.value)
                  }
                />
                <Show
                  when={manualPartnerOverride()}
                  fallback={
                    <div class="space-y-2 rounded-2xl border border-dashed border-gray-300 bg-gray-50 p-3">
                      <p class="text-sm font-medium text-gray-700">
                        Partner routing mode
                      </p>
                      <p class="text-xs text-gray-500">
                        Auto-route is active. The backend will pick the best eligible
                        partner from capability, priority, and SLA.
                      </p>
                      <Button
                        type="button"
                        size="xs"
                        color="alternative"
                        onClick={() => setManualPartnerOverride(true)}
                      >
                        Override partner
                      </Button>
                    </div>
                  }
                >
                  <div class="space-y-2">
                    <InputField
                      label="Preferred partner override"
                      value={preferredPartner()}
                      placeholder="optional code or name"
                      onInput={(event) =>
                        setPreferredPartner(event.currentTarget.value)
                      }
                    />
                    <Button
                      type="button"
                      size="xs"
                      color="alternative"
                      onClick={resetPreferredPartnerOverride}
                    >
                      Return to auto-route
                    </Button>
                  </div>
                </Show>
              </div>
              <Show when={routingRecommendation()}>
                {(recommendation) => (
                  <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content="routing recommendation" color="blue" />
                      <Badge
                        content={
                          manualPartnerOverride()
                            ? 'manual override'
                            : 'auto-route active'
                        }
                        color={manualPartnerOverride() ? 'yellow' : 'green'}
                      />
                      <Show when={recommendation().candidatePartner}>
                        <Badge
                          content={`candidate default ${recommendation().candidatePartner}`}
                          color="dark"
                        />
                      </Show>
                      <Show when={recommendation().selectedPartner}>
                        <Badge
                          content={`selected ${recommendation().selectedPartner}`}
                          color="green"
                        />
                      </Show>
                    </div>
                    <p class="mt-2 text-sm text-gray-600">
                      {recommendation().summary}
                    </p>
                    <Show
                      when={!manualPartnerOverride() && recommendation().selectedPartner}
                    >
                      <InfoAlert>
                        Auto-route will create the order against{' '}
                        {recommendation().selectedPartner}. Switch to override only
                        when you need to force a different eligible partner.
                      </InfoAlert>
                    </Show>
                    <div class="mt-3 space-y-3">
                      <Show
                        when={recommendation().options.filter((option) => option.eligible).length > 0}
                      >
                        <div class="space-y-2">
                          <p class="text-sm font-medium text-gray-700">
                            Eligible partners
                          </p>
                          <For
                            each={recommendation()
                              .options.filter((option) => option.eligible)
                              .slice(0, 4)}
                          >
                            {(option) => (
                              <div class="rounded-xl bg-white p-3 text-sm text-gray-600">
                                <div class="flex flex-wrap items-center gap-2">
                                  <Badge
                                    content={option.partner.name}
                                    color="green"
                                  />
                                  <Badge
                                    content={`priority ${option.partner.routingPriority}`}
                                    color="blue"
                                  />
                                  <Badge
                                    content={`${option.partner.slaDays}d sla`}
                                    color="indigo"
                                  />
                                  <Show
                                    when={
                                      recommendation().selectedPartner ===
                                      option.partner.name
                                    }
                                  >
                                    <Badge content="recommended" color="green" />
                                  </Show>
                                </div>
                                <p class="mt-2">{option.reason}</p>
                                <p class="mt-1 text-xs text-gray-500">
                                  Products:{' '}
                                  {joinPartnerCapabilityList(
                                    option.partner.supportedProductTypes
                                  )} ·
                                  Regions:{' '}
                                  {joinPartnerCapabilityList(
                                    option.partner.supportedRegions
                                  )}
                                </p>
                                <p class="mt-1 text-xs text-gray-500">
                                  Partner base fulfillment:{' '}
                                  {option.partner.baseFulfillmentCost || 'TBD'} ·
                                  Region cost rules:{' '}
                                  {joinShippingCostRules(
                                    option.partner.shippingCostRules
                                  )}
                                </p>
                                <div class="mt-2 flex flex-wrap gap-2">
                                  <Badge
                                    content={`fulfillment ${option.estimatedFulfillmentCost}`}
                                    color="blue"
                                  />
                                  <Badge
                                    content={`shipping ${option.estimatedShippingCost}`}
                                    color="indigo"
                                  />
                                  <Badge
                                    content={`unit margin ${option.estimatedUnitMargin}`}
                                    color="green"
                                  />
                                </div>
                                <div class="mt-3 flex flex-wrap gap-2">
                                  <Show
                                    when={
                                      recommendation().selectedPartner ===
                                      option.partner.name
                                    }
                                    fallback={
                                      <Button
                                        type="button"
                                        size="xs"
                                        color="alternative"
                                        onClick={() =>
                                          applyPreferredPartnerOverride(
                                            option.partner.name
                                          )
                                        }
                                      >
                                        Force this partner
                                      </Button>
                                    }
                                  >
                                    <Button
                                      type="button"
                                      size="xs"
                                      color="green"
                                      onClick={resetPreferredPartnerOverride}
                                    >
                                      Use auto-route
                                    </Button>
                                  </Show>
                                </div>
                              </div>
                            )}
                          </For>
                        </div>
                      </Show>
                      <Show
                        when={recommendation().options.some((option) => !option.eligible)}
                      >
                        <div class="space-y-2">
                          <p class="text-sm font-medium text-gray-700">
                            Blocked by capability
                          </p>
                          <For
                            each={recommendation()
                              .options.filter((option) => !option.eligible)
                              .slice(0, 3)}
                          >
                            {(option) => (
                              <div class="rounded-xl border border-red-100 bg-red-50 p-3 text-sm text-gray-600">
                                <div class="flex flex-wrap items-center gap-2">
                                  <Badge
                                    content={option.partner.name}
                                    color="red"
                                  />
                                  <Badge
                                    content={`priority ${option.partner.routingPriority}`}
                                    color="dark"
                                  />
                                  <Badge
                                    content={`${option.partner.slaDays}d sla`}
                                    color="dark"
                                  />
                                </div>
                                <p class="mt-2">{option.reason}</p>
                                <p class="mt-1 text-xs text-gray-500">
                                  Products:{' '}
                                  {joinPartnerCapabilityList(
                                    option.partner.supportedProductTypes
                                  )} ·
                                  Regions:{' '}
                                  {joinPartnerCapabilityList(
                                    option.partner.supportedRegions
                                  )}
                                </p>
                                <p class="mt-1 text-xs text-gray-500">
                                  Partner base fulfillment:{' '}
                                  {option.partner.baseFulfillmentCost || 'TBD'} ·
                                  Region cost rules:{' '}
                                  {joinShippingCostRules(
                                    option.partner.shippingCostRules
                                  )}
                                </p>
                              </div>
                            )}
                          </For>
                        </div>
                      </Show>
                      <Show
                        when={
                          recommendation().options.length === 0
                        }
                      >
                        <EmptyBlock
                          title="No active partner profiles returned"
                          copy="Create or activate partner capabilities first so the routing engine can score eligible print and fulfillment partners."
                        />
                      </Show>
                    </div>
                  </div>
                )}
              </Show>
              <SelectField
                label="Default exception scenario"
                value={selectedExceptionType()}
                options={exceptionOptions}
                onChange={(event) =>
                  setSelectedExceptionType(event.currentTarget.value)
                }
              />
              <Button type="submit" color="blue">
                {manualPartnerOverride()
                  ? 'Create routed order with override'
                  : 'Create routed order via auto-route'}
              </Button>
            </form>
          </Show>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Routing board"
            subtitle="Watch each order move from intake to production, then manage shipment and settlement state directly inside the store-scoped POD workflow."
          />

          <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
            <div class="grid gap-4 md:grid-cols-[0.7fr_1.3fr]">
              <InputField
                label="Operator lens"
                value={operatorLens()}
                placeholder="linh.nguyen"
                onInput={(event) => setOperatorLens(event.currentTarget.value)}
              />
              <div class="space-y-2">
                <p class="text-sm font-medium text-gray-700">Queue views</p>
                <div class="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    size="xs"
                    color={activeQueueView() === 'all' ? 'blue' : 'alternative'}
                    onClick={() => setActiveQueueView('all')}
                  >
                    All · {queueViewCount('all')}
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color={activeQueueView() === 'my_queue' ? 'blue' : 'alternative'}
                    onClick={() => setActiveQueueView('my_queue')}
                  >
                    My queue · {queueViewCount('my_queue')}
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color={activeQueueView() === 'overdue' ? 'red' : 'alternative'}
                    onClick={() => setActiveQueueView('overdue')}
                  >
                    Overdue · {queueViewCount('overdue')}
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color={
                      activeQueueView() === 'delivery_issues'
                        ? 'red'
                        : 'alternative'
                    }
                    onClick={() => setActiveQueueView('delivery_issues')}
                  >
                    Delivery issues · {queueViewCount('delivery_issues')}
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color={
                      activeQueueView() === 'settlement_pending'
                        ? 'green'
                        : 'alternative'
                    }
                    onClick={() => setActiveQueueView('settlement_pending')}
                  >
                    Settlement pending · {queueViewCount('settlement_pending')}
                  </Button>
                </div>
              </div>
            </div>
            <div class="mt-4 flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium text-gray-700">Sort</span>
              <Button
                type="button"
                size="xs"
                color={activeQueueSort() === 'priority' ? 'dark' : 'alternative'}
                onClick={() => setActiveQueueSort('priority')}
              >
                Priority first
              </Button>
              <Button
                type="button"
                size="xs"
                color={activeQueueSort() === 'newest' ? 'dark' : 'alternative'}
                onClick={() => setActiveQueueSort('newest')}
              >
                Newest
              </Button>
            </div>
            <div class="mt-4 rounded-2xl border border-gray-200 bg-white p-4">
              <div class="grid gap-4 md:grid-cols-[0.8fr_1.2fr]">
                <InputField
                  label="Save queue preset"
                  value={presetName()}
                  placeholder="Linh overdue"
                  onInput={(event) => setPresetName(event.currentTarget.value)}
                />
                <div class="space-y-2">
                  <p class="text-sm font-medium text-gray-700">Saved presets</p>
                  <div class="flex flex-wrap gap-2">
                    <Show
                      when={savedPresets().length > 0}
                      fallback={
                        <p class="text-sm text-gray-500">
                          No saved presets for this store yet.
                        </p>
                      }
                    >
                      <For each={savedPresets()}>
                        {(preset) => (
                          <div class="flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-2 py-1">
                            <button
                              type="button"
                              class="text-sm font-medium text-gray-700"
                              onClick={() => applyQueuePreset(preset)}
                            >
                              {preset.name}
                            </button>
                            <button
                              type="button"
                              class="text-xs font-semibold text-red-600"
                              onClick={() => deleteQueuePreset(preset.name)}
                            >
                              remove
                            </button>
                          </div>
                        )}
                      </For>
                    </Show>
                  </div>
                </div>
              </div>
              <div class="mt-4">
                <Button
                  type="button"
                  size="sm"
                  color="green"
                  onClick={saveQueuePreset}
                >
                  Save current queue view
                </Button>
              </div>
            </div>
            <div class="mt-4 rounded-2xl border border-gray-200 bg-white p-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-gray-900">
                    Bulk ops
                  </p>
                  <p class="text-sm text-gray-500">
                    Selected {selectedOrderIDs().length} order(s) in the current queue workflow.
                  </p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    size="xs"
                    color="alternative"
                    onClick={selectVisibleOrders}
                  >
                    Select visible
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color="light"
                    onClick={clearSelectedOrders}
                  >
                    Clear
                  </Button>
                </div>
              </div>
              <div class="mt-4 grid gap-4 md:grid-cols-3">
                <InputField
                  label="Bulk owner"
                  value={bulkDraft().operatorAssignee}
                  placeholder="linh.nguyen"
                  onInput={(event) =>
                    setBulkDraft((current) => ({
                      ...current,
                      operatorAssignee: event.currentTarget.value,
                    }))
                  }
                />
                <InputField
                  label="Bulk shipment SLA"
                  type="datetime-local"
                  value={bulkDraft().shipmentSlaDueAt}
                  onInput={(event) =>
                    setBulkDraft((current) => ({
                      ...current,
                      shipmentSlaDueAt: event.currentTarget.value,
                      shipmentSlaMode: '',
                    }))
                  }
                />
                <SelectField
                  label="Relative SLA"
                  value={bulkDraft().shipmentSlaMode}
                  options={[
                    { name: 'No preset', value: '' },
                    { name: '+2h', value: 'plus_2h' },
                    { name: '+4h', value: 'plus_4h' },
                    { name: 'End of day', value: 'end_of_day' },
                  ]}
                  onChange={(event) =>
                    setBulkDraft((current) => ({
                      ...current,
                      shipmentSlaMode: event.currentTarget.value as ShipmentSlaMode,
                      shipmentSlaDueAt: event.currentTarget.value
                        ? toLocalDateTimeValue(
                            resolveShipmentSla(
                              event.currentTarget.value as ShipmentSlaMode
                            )
                          )
                        : current.shipmentSlaDueAt,
                    }))
                  }
                />
                <SelectField
                  label="Bulk settlement status"
                  value={bulkDraft().settlementStatus}
                  options={[
                    { name: 'No change', value: '' },
                    ...settlementOptions,
                  ]}
                  onChange={(event) =>
                    setBulkDraft((current) => ({
                      ...current,
                      settlementStatus: event.currentTarget.value,
                    }))
                  }
                />
              </div>
              <div class="mt-4 rounded-2xl border border-gray-200 bg-gray-50 p-4">
                <div class="grid gap-4 md:grid-cols-[0.8fr_1.2fr]">
                  <InputField
                    label="Save bulk template"
                    value={bulkTemplateName()}
                    placeholder="Carrier claim follow-up"
                    onInput={(event) =>
                      setBulkTemplateName(event.currentTarget.value)
                    }
                  />
                  <div class="space-y-2">
                    <p class="text-sm font-medium text-gray-700">
                      Saved bulk templates
                    </p>
                    <div class="flex flex-wrap gap-2">
                      <Show
                        when={savedBulkTemplates().length > 0}
                        fallback={
                          <p class="text-sm text-gray-500">
                            No saved bulk templates for this store yet.
                          </p>
                        }
                      >
                        <For each={savedBulkTemplates()}>
                          {(template) => (
                            <div class="flex items-center gap-2 rounded-full border border-gray-200 bg-white px-2 py-1">
                              <button
                                type="button"
                                class="text-sm font-medium text-gray-700"
                                onClick={() => applyBulkTemplate(template)}
                              >
                                {template.name}
                              </button>
                              <button
                                type="button"
                                class="text-xs font-semibold text-red-600"
                                onClick={() => deleteBulkTemplate(template.name)}
                              >
                                remove
                              </button>
                            </div>
                          )}
                        </For>
                      </Show>
                    </div>
                  </div>
                </div>
                <div class="mt-4">
                  <Button
                    type="button"
                    size="xs"
                    color="green"
                    onClick={saveBulkTemplate}
                  >
                    Save current bulk template
                  </Button>
                </div>
              </div>
              <div class="mt-4">
                <Button
                  type="button"
                  size="sm"
                  color="dark"
                  onClick={applyBulkUpdate}
                >
                  Apply bulk update
                </Button>
              </div>
            </div>
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
              <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="text-sm font-semibold text-slate-900">
                      Store activity feed
                    </p>
                    <p class="text-sm text-slate-500">
                      Latest activity across the current queue slice for store {params().tenantId}.
                    </p>
                  </div>
                  <Button
                    type="button"
                    size="xs"
                    color="light"
                    onClick={() => {
                      void copyStoreActivityFeed();
                    }}
                  >
                    Copy feed
                  </Button>
                  <Button
                    type="button"
                    size="xs"
                    color="alternative"
                    href={`/t/${params().tenantId}/orders/audit`}
                  >
                    Open full audit
                  </Button>
                </div>
                <div class="mt-4 space-y-3">
                  <Show
                    when={storeActivityFeed().length > 0}
                    fallback={
                      <div class="rounded-xl border border-dashed border-slate-200 bg-white p-3 text-sm text-slate-500">
                        No store activity matches the current queue and activity filters.
                      </div>
                    }
                  >
                    <For each={storeActivityFeed()}>
                      {(entry) => (
                        <div class="rounded-xl border border-slate-200 bg-white p-3">
                          <div class="flex flex-wrap items-center justify-between gap-2">
                            <div class="flex flex-wrap items-center gap-2">
                              <Badge
                                content={entry.activity.type.replaceAll('_', ' ')}
                                color={activityColor(entry.activity.type)}
                              />
                              <p class="text-xs font-semibold text-slate-700">
                                {entry.orderId}
                              </p>
                              <p class="text-xs text-slate-500">
                                {entry.productTitle}
                              </p>
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
              <For each={sortedOrders()}>
                {(order) => (
                  <div class="rounded-2xl border border-gray-200 bg-white p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div class="flex items-start gap-3">
                        <label class="mt-1">
                          <input
                            type="checkbox"
                            checked={isSelected(order.id)}
                            onChange={(event) =>
                              toggleOrderSelection(
                                order.id,
                                event.currentTarget.checked
                              )
                            }
                          />
                        </label>
                        <div>
                        <p class="text-base font-semibold text-gray-900">
                          {order.id}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          {order.productTitle} · {order.partner}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          customer {order.customerName} · qty {order.quantity} · total {order.total}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          owner {order.operatorAssignee || 'unassigned'}
                        </p>
                        </div>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Show when={activeQueueSort() === 'priority'}>
                          <Badge
                            content={`priority ${priorityScoreFor(order) + 1}`}
                            color={priorityScoreFor(order) < 3 ? 'red' : 'dark'}
                          />
                        </Show>
                        <Badge
                          content={order.status.replaceAll('_', ' ')}
                          color={statusColor(order.status)}
                        />
                        <Show when={order.exceptionStatus}>
                          <Badge
                            content={`${order.exceptionStatus} issue`}
                            color={exceptionColor(order.exceptionStatus)}
                          />
                        </Show>
                        <Badge
                          content={order.shipmentStatus.replaceAll('_', ' ')}
                          color={shipmentColor(order.shipmentStatus)}
                        />
                        <Badge
                          content={order.settlementStatus.replaceAll('_', ' ')}
                          color={settlementColor(order.settlementStatus)}
                        />
                        <Button
                          type="button"
                          size="xs"
                          color="green"
                          disabled={
                            order.status === 'shipped' ||
                            order.exceptionStatus === 'open' ||
                            order.exceptionStatus === 'escalated'
                          }
                          onClick={() => {
                            advanceOrder(order.id);
                          }}
                        >
                          Advance route
                        </Button>
                        <Button
                          type="button"
                          size="xs"
                          color="alternative"
                          disabled={
                            order.exceptionStatus === 'open' ||
                            order.exceptionStatus === 'resolved'
                          }
                          onClick={() => {
                            raiseException(order.id);
                          }}
                        >
                          Raise issue
                        </Button>
                        <Show when={order.exceptionStatus === 'open'}>
                          <Button
                            type="button"
                            size="xs"
                            color="blue"
                            onClick={() => {
                              updateExceptionStatus(order.id, 'escalated');
                            }}
                          >
                            Escalate
                          </Button>
                          <Button
                            type="button"
                            size="xs"
                            color="light"
                            onClick={() => {
                              updateExceptionStatus(order.id, 'resolved');
                            }}
                          >
                            Resolve
                          </Button>
                        </Show>
                      </div>
                    </div>

                    <Show when={order.exceptionType}>
                      <div class="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-3">
                        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-amber-700">
                          Exception
                        </p>
                        <p class="mt-2 text-sm text-amber-900">
                          {order.exceptionType.replaceAll('_', ' ')} ·{' '}
                          {order.exceptionStatus || 'draft'}
                        </p>
                      </div>
                    </Show>

                    <div class="mt-3 rounded-xl border border-sky-200 bg-sky-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-sky-700">
                        Queue ownership
                      </p>
                      <div class="mt-3 grid gap-4 md:grid-cols-2">
                        <InputField
                          label="Operator assignee"
                          value={queueDraftFor(order).operatorAssignee}
                          placeholder="linh.nguyen"
                          onInput={(event) =>
                            patchQueueDraft(order.id, {
                              operatorAssignee: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Shipment SLA due"
                          type="datetime-local"
                          value={queueDraftFor(order).shipmentSlaDueAt}
                          onInput={(event) =>
                            patchQueueDraft(order.id, {
                              shipmentSlaDueAt: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Issue SLA due"
                          type="datetime-local"
                          value={queueDraftFor(order).issueSlaDueAt}
                          onInput={(event) =>
                            patchQueueDraft(order.id, {
                              issueSlaDueAt: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="xs"
                          color="blue"
                          onClick={() => {
                            saveQueueControl(order);
                          }}
                        >
                          Save queue control
                        </Button>
                        <Badge
                          content={`owner ${order.operatorAssignee || 'unassigned'}`}
                          color="indigo"
                        />
                        <Show when={order.shipmentSlaDueAt}>
                          <Badge
                            content={`shipment SLA ${isOverdue(order.shipmentSlaDueAt) ? 'overdue' : 'set'}`}
                            color={isOverdue(order.shipmentSlaDueAt) ? 'red' : 'blue'}
                          />
                        </Show>
                        <Show when={order.issueSlaDueAt}>
                          <Badge
                            content={`issue SLA ${isOverdue(order.issueSlaDueAt) ? 'overdue' : 'set'}`}
                            color={isOverdue(order.issueSlaDueAt) ? 'red' : 'blue'}
                          />
                        </Show>
                      </div>
                    </div>

                    <Show
                      when={
                        order.exceptionType ||
                        order.shipmentStatus === 'delivery_issue'
                      }
                    >
                      <div class="mt-3 rounded-xl border border-rose-200 bg-rose-50 p-3">
                        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-rose-700">
                          Issue cost handling
                        </p>
                        <div class="mt-3 grid gap-4 md:grid-cols-2">
                          <InputField
                            label="Issue cost"
                            value={issueDraftFor(order).issueCost}
                            placeholder="$6.00"
                            onInput={(event) =>
                              patchIssueDraft(order.id, {
                                issueCost: event.currentTarget.value,
                              })
                            }
                          />
                          <SelectField
                            label="Resolution path"
                            value={issueDraftFor(order).issueResolution}
                            options={issueResolutionOptions}
                            onChange={(event) =>
                              patchIssueDraft(order.id, {
                                issueResolution: event.currentTarget.value,
                              })
                            }
                          />
                        </div>
                        <div class="mt-4">
                          <TextareaField
                            label="Issue notes"
                            value={issueDraftFor(order).issueNotes}
                            rows={3}
                            onInput={(event) =>
                              patchIssueDraft(order.id, {
                                issueNotes: event.currentTarget.value,
                              })
                            }
                          />
                        </div>
                        <div class="mt-3 flex flex-wrap items-center gap-2">
                          <Button
                            type="button"
                            size="xs"
                            color="red"
                            onClick={() => {
                              saveIssueHandling(order);
                            }}
                          >
                            Save issue handling
                          </Button>
                          <Badge content={`cost ${order.issueCost}`} color="red" />
                          <Badge
                            content={order.issueResolution.replaceAll('_', ' ')}
                            color="yellow"
                          />
                        </div>
                      </div>
                    </Show>

                    <div class="mt-3 rounded-xl border border-emerald-200 bg-emerald-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                        Settlement readiness
                      </p>
                      <div class="mt-3 grid gap-3 md:grid-cols-2">
                        <div class="rounded-xl border border-emerald-200 bg-white p-3">
                          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                            Base cost snapshot
                          </p>
                          <p class="mt-2 text-sm font-semibold text-gray-900">
                            {order.baseCostSnapshot}
                          </p>
                        </div>
                        <div class="rounded-xl border border-emerald-200 bg-white p-3">
                          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-emerald-700">
                            Realized margin
                          </p>
                          <p class="mt-2 text-sm font-semibold text-gray-900">
                            {order.realizedMargin}
                          </p>
                        </div>
                      </div>
                      <div class="mt-3 grid gap-4 md:grid-cols-2">
                        <InputField
                          label="Fulfillment cost"
                          value={settlementDraftFor(order).fulfillmentCost}
                          placeholder="$9.50"
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              fulfillmentCost: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Shipping cost"
                          value={settlementDraftFor(order).shippingCost}
                          placeholder="$4.25"
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              shippingCost: event.currentTarget.value,
                            })
                          }
                        />
                        <SelectField
                          label="Settlement status"
                          value={settlementDraftFor(order).settlementStatus}
                          options={settlementOptions}
                          onChange={(event) =>
                            patchSettlementDraft(order.id, {
                              settlementStatus: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-4">
                        <TextareaField
                          label="Settlement notes"
                          value={settlementDraftFor(order).settlementNotes}
                          rows={3}
                          onInput={(event) =>
                            patchSettlementDraft(order.id, {
                              settlementNotes: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="xs"
                          color="green"
                          onClick={() => {
                            saveSettlement(order);
                          }}
                        >
                          Save settlement state
                        </Button>
                        <Badge
                          content={`margin ${order.realizedMargin}`}
                          color="green"
                        />
                        <Badge
                          content={`settlement ${order.settlementStatus.replaceAll('_', ' ')}`}
                          color={settlementColor(order.settlementStatus)}
                        />
                      </div>
                    </div>

                    <div class="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
                        Manual shipment control
                      </p>
                      <div class="mt-3 grid gap-4 md:grid-cols-2">
                        <SelectField
                          label="Shipment status"
                          value={shipmentDraftFor(order).shipmentStatus}
                          options={shipmentOptions}
                          onChange={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentStatus: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Carrier"
                          value={shipmentDraftFor(order).shipmentCarrier}
                          placeholder="UPS"
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentCarrier: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Tracking number"
                          value={shipmentDraftFor(order).shipmentTrackingNumber}
                          placeholder="1Z999..."
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentTrackingNumber: event.currentTarget.value,
                            })
                          }
                        />
                        <InputField
                          label="Tracking URL"
                          value={shipmentDraftFor(order).shipmentTrackingUrl}
                          placeholder="https://tracking.example/..."
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentTrackingUrl: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-4">
                        <TextareaField
                          label="Shipment notes"
                          value={shipmentDraftFor(order).shipmentNotes}
                          rows={3}
                          onInput={(event) =>
                            patchShipmentDraft(order.id, {
                              shipmentNotes: event.currentTarget.value,
                            })
                          }
                        />
                      </div>
                      <div class="mt-3 flex flex-wrap items-center gap-2">
                        <Button
                          type="button"
                          size="xs"
                          color="blue"
                          onClick={() => {
                            saveShipment(order);
                          }}
                        >
                          Save shipment state
                        </Button>
                        <Show when={order.shipmentCarrier || order.shipmentTrackingNumber}>
                          <Badge
                            content={`${order.shipmentCarrier || 'manual carrier'} ${order.shipmentTrackingNumber || ''}`.trim()}
                            color="indigo"
                          />
                        </Show>
                        <Show when={order.deliveredAt}>
                          <Badge content="Delivered confirmed" color="green" />
                        </Show>
                      </div>
                    </div>

                    <div class="mt-3 rounded-xl bg-gray-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
                        Timeline
                      </p>
                      <div class="mt-2 space-y-1 text-sm text-gray-600">
                        <For each={order.timeline}>
                          {(entry) => <p>{entry}</p>}
                        </For>
                      </div>
                    </div>

                    <div class="mt-3 rounded-xl border border-slate-200 bg-white p-3">
                      <div class="flex flex-wrap items-center justify-between gap-3">
                        <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-600">
                          Activity log
                        </p>
                        <div class="flex flex-wrap items-center gap-2">
                          <div class="min-w-[11rem]">
                            <SelectField
                              label=""
                              value={activityFilter()}
                              options={activityFilterOptions.map((option) => ({
                                name: option.name,
                                value: option.value,
                              }))}
                              onChange={(event) =>
                                setActivityFilter(event.currentTarget.value as ActivityFilter)
                              }
                            />
                          </div>
                          <Show when={activityFilter() === 'all'}>
                            <Button
                              type="button"
                              size="xs"
                              color={hideSystemActivity() ? 'dark' : 'light'}
                              onClick={() =>
                                setHideSystemActivity((current) => !current)
                              }
                            >
                              {hideSystemActivity()
                                ? 'Show system'
                                : 'Hide system'}
                            </Button>
                          </Show>
                          <Button
                            type="button"
                            size="xs"
                            color="light"
                            onClick={() => {
                              void copyActivitySummary(order);
                            }}
                          >
                            Copy summary
                          </Button>
                        </div>
                      </div>
                      <div class="mt-3 space-y-3">
                        <Show
                          when={filteredActivityLogFor(order).length > 0}
                          fallback={
                            <div class="rounded-xl border border-dashed border-slate-200 bg-slate-50 p-3 text-sm text-slate-500">
                              <Show
                                when={hiddenSystemActivityCountFor(order) > 0}
                                fallback={'No activity matches the current filter.'}
                              >
                                {hiddenSystemActivityCountFor(order)} system updates are hidden.
                              </Show>
                            </div>
                          }
                        >
                        <For each={filteredActivityLogFor(order)}>
                          {(activity) => (
                            <div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
                              <div class="flex flex-wrap items-center justify-between gap-2">
                                <div class="flex flex-wrap items-center gap-2">
                                  <Badge
                                    content={activity.type.replaceAll('_', ' ')}
                                    color={activityColor(activity.type)}
                                  />
                                  <p class="text-xs font-medium text-slate-500">
                                    {formatActivityActor(activity.actor)}
                                  </p>
                                </div>
                                <p class="text-xs text-slate-500">
                                  {formatActivityTime(activity.createdAt)}
                                </p>
                              </div>
                              <p class="mt-2 text-sm text-slate-700">
                                {activity.message}
                              </p>
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
                  </div>
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
                color={stage.value === 'reprint' || stage.value === 'refund' ? 'red' : 'yellow'}
              />
            )}
          </For>
        </div>
      </Card>
    </PageShell>
  );
}
