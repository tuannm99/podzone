import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps } from 'solid-js'
import type { RoutedOrderRecommendation } from '@podzone/shared/services/orders'
import type { CatalogCandidate } from '@podzone/shared/services/productSetup'
import type { FormStore } from '@podzone/shared/ui/forms'
import type { RoutedOrderFormValues } from './forms'

export type TenantOrdersComposerContextValue = {
    availableCandidates: Accessor<CatalogCandidate[]>
    form: FormStore<RoutedOrderFormValues>
    routingRecommendation: Accessor<RoutedOrderRecommendation | null>
    applyPreferredPartnerOverride: (partnerName: string) => void
    resetPreferredPartnerOverride: () => void
    createMockOrder: (event: SubmitEvent) => Promise<void>
}

const TenantOrdersComposerContext = createContext<TenantOrdersComposerContextValue>()

export function TenantOrdersComposerProvider(props: ParentProps<{ value: TenantOrdersComposerContextValue }>) {
    return (
        <TenantOrdersComposerContext.Provider value={props.value}>
            {props.children}
        </TenantOrdersComposerContext.Provider>
    )
}

export function useTenantOrdersComposer() {
    const ctx = useContext(TenantOrdersComposerContext)
    if (!ctx) {
        throw new Error('TenantOrdersComposerContext is missing')
    }
    return ctx
}
