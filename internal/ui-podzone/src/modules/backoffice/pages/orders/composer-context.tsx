import { createContext, useContext } from 'solid-js';
import type { Accessor, ParentProps, Setter } from 'solid-js';
import type {
  RoutedOrderRecommendation,
} from '@/services/orders';
import type { CatalogCandidate } from '@/services/productSetup';

export type TenantOrdersComposerContextValue = {
  availableCandidates: Accessor<CatalogCandidate[]>;
  selectedCandidateId: Accessor<string>;
  setSelectedCandidateId: Setter<string>;
  customerName: Accessor<string>;
  setCustomerName: Setter<string>;
  quantity: Accessor<string>;
  setQuantity: Setter<string>;
  selectedProductType: Accessor<string>;
  setSelectedProductType: Setter<string>;
  selectedShipRegion: Accessor<string>;
  setSelectedShipRegion: Setter<string>;
  preferredPartner: Accessor<string>;
  setPreferredPartner: Setter<string>;
  manualPartnerOverride: Accessor<boolean>;
  setManualPartnerOverride: Setter<boolean>;
  routingRecommendation: Accessor<RoutedOrderRecommendation | null>;
  selectedExceptionType: Accessor<string>;
  setSelectedExceptionType: Setter<string>;
  applyPreferredPartnerOverride: (partnerName: string) => void;
  resetPreferredPartnerOverride: () => void;
  createMockOrder: (event: SubmitEvent) => Promise<void>;
};

const TenantOrdersComposerContext =
  createContext<TenantOrdersComposerContextValue>();

export function TenantOrdersComposerProvider(
  props: ParentProps<{ value: TenantOrdersComposerContextValue }>
) {
  return (
    <TenantOrdersComposerContext.Provider value={props.value}>
      {props.children}
    </TenantOrdersComposerContext.Provider>
  );
}

export function useTenantOrdersComposer() {
  const ctx = useContext(TenantOrdersComposerContext);
  if (!ctx) {
    throw new Error('TenantOrdersComposerContext is missing');
  }
  return ctx;
}

