import type { Validator } from '@podzone/shared/ui/forms'

export type PartnerFormValues = {
    name: string
    code: string
    contactName: string
    contactEmail: string
    notes: string
    partnerType: string
    supportedProductTypes: string
    supportedRegions: string
    slaDays: string
    routingPriority: string
    baseFulfillmentCost: string
    shippingCostRules: string
}

export const partnerInitialValues: PartnerFormValues = {
    name: '',
    code: '',
    contactName: '',
    contactEmail: '',
    notes: '',
    partnerType: 'print_on_demand',
    supportedProductTypes: 'tshirt, hoodie',
    supportedRegions: 'us, eu',
    slaDays: '3',
    routingPriority: '100',
    baseFulfillmentCost: '$8.50',
    shippingCostRules: 'us:$4.50, eu:$6.00',
}

export const nonNegativeInteger =
    (message: string): Validator<PartnerFormValues> =>
    (value) => {
        if (typeof value !== 'string' || value.trim().length === 0) {
            return undefined
        }
        return /^\d+$/.test(value.trim()) ? undefined : message
    }

export const shippingRules =
    (message: string): Validator<PartnerFormValues> =>
    (value) => {
        if (typeof value !== 'string' || value.trim().length === 0) {
            return undefined
        }
        const valid = value.split(',').every((item) => {
            const [region, cost, ...rest] = item.trim().split(':')
            return Boolean(
                region?.trim() && cost?.trim() && rest.length === 0 && /^\$?\d+(?:\.\d{1,2})?$/.test(cost.trim())
            )
        })
        return valid ? undefined : message
    }
