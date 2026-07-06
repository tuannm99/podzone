import type { Validator } from '@/solid/forms'

export type ProductSetupFormValues = {
    name: string
    partner: string
    baseCost: string
    retailPrice: string
    status: string
    notes: string
    channel: string
    variantColor: string
    variantSize: string
    hasFrontArtwork: boolean
    hasBackArtwork: boolean
    mockupReady: boolean
    printSpecChecked: boolean
}

export const productSetupInitialValues: ProductSetupFormValues = {
    name: '',
    partner: '',
    baseCost: '',
    retailPrice: '',
    status: 'draft',
    notes: '',
    channel: 'website_store',
    variantColor: 'Black',
    variantSize: 'M',
    hasFrontArtwork: true,
    hasBackArtwork: false,
    mockupReady: false,
    printSpecChecked: false,
}

export const moneyValue =
    (message: string): Validator<ProductSetupFormValues> =>
    (value) => {
        if (typeof value !== 'string' || value.trim().length === 0) {
            return undefined
        }
        return /^\$?\d+(?:\.\d{1,2})?$/.test(value.trim()) ? undefined : message
    }
