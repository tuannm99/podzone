import { createSignal } from 'solid-js'
import type { RoutedOrder } from '@/services/orders'
import { toLocalDateTimeValue } from '../shared/presentation'
import type { IssueDraft, QueueDraft, RerouteDraft, SettlementDraft, ShipmentDraft } from '../order-card/types'

export function useOrderDrafts() {
    const [shipmentDrafts, setShipmentDrafts] = createSignal<Record<string, ShipmentDraft>>({})
    const [settlementDrafts, setSettlementDrafts] = createSignal<Record<string, SettlementDraft>>({})
    const [issueDrafts, setIssueDrafts] = createSignal<Record<string, IssueDraft>>({})
    const [queueDrafts, setQueueDrafts] = createSignal<Record<string, QueueDraft>>({})
    const [rerouteDrafts, setRerouteDrafts] = createSignal<Record<string, RerouteDraft>>({})

    const shipmentDraftFor = (order: RoutedOrder): ShipmentDraft =>
        shipmentDrafts()[order.id] || {
            shipmentStatus: order.shipmentStatus || 'awaiting_label',
            shipmentCarrier: order.shipmentCarrier || '',
            shipmentTrackingNumber: order.shipmentTrackingNumber || '',
            shipmentTrackingUrl: order.shipmentTrackingUrl || '',
            shipmentNotes: order.shipmentNotes || '',
        }

    const settlementDraftFor = (order: RoutedOrder): SettlementDraft =>
        settlementDrafts()[order.id] || {
            fulfillmentCost: order.fulfillmentCost || order.baseCostSnapshot || '$0.00',
            shippingCost: order.shippingCost || '$0.00',
            settlementStatus: order.settlementStatus || 'pending',
            settlementNotes: order.settlementNotes || '',
        }

    const issueDraftFor = (order: RoutedOrder): IssueDraft =>
        issueDrafts()[order.id] || {
            issueCost: order.issueCost || '$0.00',
            issueResolution: order.issueResolution || 'monitor',
            issueNotes: order.issueNotes || '',
        }

    const queueDraftFor = (order: RoutedOrder): QueueDraft =>
        queueDrafts()[order.id] || {
            operatorAssignee: order.operatorAssignee || 'unassigned',
            shipmentSlaDueAt: toLocalDateTimeValue(order.shipmentSlaDueAt),
            issueSlaDueAt: toLocalDateTimeValue(order.issueSlaDueAt),
        }

    const rerouteDraftFor = (order: RoutedOrder): RerouteDraft =>
        rerouteDrafts()[order.id] || {
            preferredPartner: order.partner || '',
        }

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
        }))
    }

    const patchSettlementDraft = (orderId: string, patch: Partial<SettlementDraft>) => {
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
        }))
    }

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
        }))
    }

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
        }))
    }

    const patchRerouteDraft = (orderId: string, patch: Partial<RerouteDraft>) => {
        setRerouteDrafts((current) => ({
            ...current,
            [orderId]: {
                ...(current[orderId] || {
                    preferredPartner: '',
                }),
                ...patch,
            },
        }))
    }

    return {
        setShipmentDrafts,
        setSettlementDrafts,
        setIssueDrafts,
        setQueueDrafts,
        setRerouteDrafts,
        shipmentDraftFor,
        settlementDraftFor,
        issueDraftFor,
        queueDraftFor,
        rerouteDraftFor,
        patchShipmentDraft,
        patchSettlementDraft,
        patchIssueDraft,
        patchQueueDraft,
        patchRerouteDraft,
    }
}
