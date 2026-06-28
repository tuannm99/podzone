import type { Validator } from '@/solid/forms'

export type RoutedOrderFormValues = {
  selectedCandidateId: string
  customerName: string
  quantity: string
  selectedProductType: string
  selectedShipRegion: string
  preferredPartner: string
  manualPartnerOverride: boolean
  selectedExceptionType: string
}

export const routedOrderInitialValues: RoutedOrderFormValues = {
  selectedCandidateId: '',
  customerName: '',
  quantity: '1',
  selectedProductType: 'tshirt',
  selectedShipRegion: 'us',
  preferredPartner: '',
  manualPartnerOverride: false,
  selectedExceptionType: 'artwork_issue',
}

export const positiveInteger =
  (message: string): Validator<RoutedOrderFormValues> =>
  (value) => {
    if (typeof value !== 'string' || value.trim().length === 0) {
      return undefined
    }
    return /^\d+$/.test(value.trim()) && Number(value) > 0 ? undefined : message
  }
