import type { RoutedOrder } from '@/services/orders';

export type ActivityFilter =
  | 'all'
  | 'notes'
  | 'system'
  | 'shipment_note'
  | 'settlement_note'
  | 'issue_note';

export type QueueSort = 'priority' | 'newest';

export type BadgeColor =
  | 'blue'
  | 'indigo'
  | 'green'
  | 'yellow'
  | 'pink'
  | 'dark'
  | 'red';

export type ShipmentDraft = {
  shipmentStatus: string;
  shipmentCarrier: string;
  shipmentTrackingNumber: string;
  shipmentTrackingUrl: string;
  shipmentNotes: string;
};

export type SettlementDraft = {
  fulfillmentCost: string;
  shippingCost: string;
  settlementStatus: string;
  settlementNotes: string;
};

export type IssueDraft = {
  issueCost: string;
  issueResolution: string;
  issueNotes: string;
};

export type QueueDraft = {
  operatorAssignee: string;
  shipmentSlaDueAt: string;
  issueSlaDueAt: string;
};

export type RerouteDraft = {
  preferredPartner: string;
};

export type SelectOption = {
  name: string;
  value: string;
};

export type ActivityDetail = {
  key: string;
  value: string;
};

export type OrderActivity = {
  type: string;
  actor: string;
  createdAt: string;
  message: string;
  details: ActivityDetail[];
};

export type OrderCardActions = {
  toggleSelected: (checked: boolean) => void;
  advanceOrder: (orderId: string) => void;
  raiseException: (orderId: string) => void;
  updateExceptionStatus: (orderId: string, nextStatus: string) => void;
  rerouteBlockedOrder: (order: RoutedOrder) => void;
  saveQueueControl: (order: RoutedOrder) => void;
  saveIssueHandling: (order: RoutedOrder) => void;
  saveSettlement: (order: RoutedOrder) => void;
  saveShipment: (order: RoutedOrder) => void;
  copyActivitySummary: (order: RoutedOrder) => Promise<void>;
  queueDraftFor: (order: RoutedOrder) => QueueDraft;
  patchQueueDraft: (orderId: string, patch: Partial<QueueDraft>) => void;
  issueDraftFor: (order: RoutedOrder) => IssueDraft;
  patchIssueDraft: (orderId: string, patch: Partial<IssueDraft>) => void;
  settlementDraftFor: (order: RoutedOrder) => SettlementDraft;
  patchSettlementDraft: (
    orderId: string,
    patch: Partial<SettlementDraft>
  ) => void;
  shipmentDraftFor: (order: RoutedOrder) => ShipmentDraft;
  patchShipmentDraft: (orderId: string, patch: Partial<ShipmentDraft>) => void;
  rerouteDraftFor: (order: RoutedOrder) => RerouteDraft;
  patchRerouteDraft: (orderId: string, patch: Partial<RerouteDraft>) => void;
};

export type OrderCardHelpers = {
  queueSort: QueueSort;
  priorityScoreFor: (order: RoutedOrder) => number;
  statusColor: (status: string) => BadgeColor;
  exceptionColor: (status: string) => BadgeColor;
  shipmentColor: (status: string) => BadgeColor;
  settlementColor: (status: string) => BadgeColor;
  activityColor: (type: string) => BadgeColor;
  isOverdue: (value?: string) => boolean;
  filteredActivityLogFor: (order: RoutedOrder) => OrderActivity[];
  hiddenSystemActivityCountFor: (order: RoutedOrder) => number;
};

export type OrderCardUi = {
  activityFilter: ActivityFilter;
  setActivityFilter: (value: ActivityFilter) => void;
  hideSystemActivity: boolean;
  toggleHideSystemActivity: () => void;
  activityFilterOptions: SelectOption[];
  shipmentOptions: SelectOption[];
  settlementOptions: SelectOption[];
  issueResolutionOptions: SelectOption[];
};
