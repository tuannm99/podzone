export const partnerTypeOptions = [
    { name: 'All partner types', value: '' },
    { name: 'Print on demand', value: 'print_on_demand' },
    { name: 'Fulfillment partner', value: 'fulfillment' },
    { name: 'Dropship supplier', value: 'dropship_supplier' },
]

export const partnerStatusOptions = [
    { name: 'All statuses', value: '' },
    { name: 'Active', value: 'active' },
    { name: 'Inactive', value: 'inactive' },
]

export function badgeColorForStatus(status: string) {
    return status === 'active' ? 'green' : 'dark'
}

export function partnerTypeLabel(partnerType: string) {
    return partnerType.replaceAll('_', ' ')
}

export function joinCapabilityList(items: string[]) {
    return (items || []).join(', ')
}

export function parseCapabilityList(raw: string) {
    return raw
        .split(',')
        .map((item) => item.trim().toLowerCase())
        .filter(Boolean)
}

export function joinShippingCostRules(rules: { region: string; cost: string }[] | undefined) {
    return (rules || []).map((rule) => `${rule.region}:${rule.cost}`).join(', ')
}

export function parseShippingCostRules(raw: string) {
    return raw
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean)
        .map((item) => {
            const [region, ...costParts] = item.split(':')
            return {
                region: region.trim().toLowerCase(),
                cost: costParts.join(':').trim(),
            }
        })
        .filter((item) => item.region && item.cost)
}
