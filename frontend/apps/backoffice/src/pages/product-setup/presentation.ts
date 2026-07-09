import type { ArtworkChecklist } from '@podzone/shared/services/productSetup'

export const statusOptions = [
    { name: 'Draft', value: 'draft' },
    { name: 'Ready for review', value: 'ready_for_review' },
    { name: 'Publish candidate', value: 'publish_candidate' },
]

export const channelOptions = [
    { name: 'Website store', value: 'website_store' },
    { name: 'Marketplace mock', value: 'marketplace_mock' },
    { name: 'Wholesale mock', value: 'wholesale_mock' },
]

export function statusBadgeColor(status: string) {
    switch (status) {
        case 'publish_candidate':
            return 'green'
        case 'ready_for_review':
            return 'yellow'
        default:
            return 'dark'
    }
}

export function candidateStatusColor(status: string) {
    switch (status) {
        case 'published_mock':
            return 'green'
        case 'ready':
            return 'blue'
        case 'archived':
            return 'dark'
        default:
            return 'yellow'
    }
}

export function checklistCompletion(checklist: ArtworkChecklist) {
    const completed = [
        checklist.frontArtwork,
        checklist.backArtwork,
        checklist.mockupReady,
        checklist.printSpecChecked,
    ].filter(Boolean).length
    return `${completed}/4 ready`
}
