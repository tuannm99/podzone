import { createContext, useContext } from 'solid-js';
import type { Accessor, ParentProps } from 'solid-js';
import type { RoutedOrder } from '../../../../services/orders';
import type {
  PartnerFinanceSummaryItem,
  StoreActivityFeedEntry,
} from './utils';

export type TenantOrdersInsightsContextValue = {
  tenantId: string;
  blockedOrders: Accessor<RoutedOrder[]>;
  blockedReasonSummary: Accessor<{ code: string; count: number }[]>;
  forcedRerouteSummary: Accessor<{ partner: string; count: number }[]>;
  reconciliationOrders: Accessor<RoutedOrder[]>;
  partnerFinanceSummary: Accessor<PartnerFinanceSummaryItem[]>;
  storeActivityFeed: Accessor<StoreActivityFeedEntry[]>;
  copyStoreActivityFeed: () => Promise<void>;
};

const TenantOrdersInsightsContext =
  createContext<TenantOrdersInsightsContextValue>();

export function TenantOrdersInsightsProvider(
  props: ParentProps<{ value: TenantOrdersInsightsContextValue }>
) {
  return (
    <TenantOrdersInsightsContext.Provider value={props.value}>
      {props.children}
    </TenantOrdersInsightsContext.Provider>
  );
}

export function useTenantOrdersInsights() {
  const ctx = useContext(TenantOrdersInsightsContext);
  if (!ctx) {
    throw new Error('TenantOrdersInsightsContext is missing');
  }
  return ctx;
}
