import { createEffect, createResource, createSignal, type Accessor } from 'solid-js'
import { getRoutedOrders } from '@/services/orders'
import { getProductSetupSnapshot } from '@/services/productSetup'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'
import { formatMoney, isOverdue, parseMoney } from './presentation'

interface HomeViewModelOptions {
    tenantID: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createHomeViewModel(options: HomeViewModelOptions) {
    const [tenantReady, setTenantReady] = createSignal(false)
    const [activeTenantId, setActiveTenantId] = createSignal('')
    const [draftCount, setDraftCount] = createSignal(0)
    const [publishedCandidateCount, setPublishedCandidateCount] = createSignal(0)
    const [inProductionCount, setInProductionCount] = createSignal(0)
    const [openExceptionCount, setOpenExceptionCount] = createSignal(0)
    const [mockRevenue, setMockRevenue] = createSignal('$0.00')
    const [realizedMarginTotal, setRealizedMarginTotal] = createSignal('$0.00')
    const [pendingSettlementCount, setPendingSettlementCount] = createSignal(0)
    const [disputedSettlementCount, setDisputedSettlementCount] = createSignal(0)
    const [issueCostExposure, setIssueCostExposure] = createSignal('$0.00')
    const [shipmentSlaOverdueCount, setShipmentSlaOverdueCount] = createSignal(0)
    const [issueSlaOverdueCount, setIssueSlaOverdueCount] = createSignal(0)
    const [topPartnerLoad, setTopPartnerLoad] = createSignal('No partner load yet')
    const [issueRate, setIssueRate] = createSignal('0%')
    const [error, setError] = createSignal('')
    const [snapshot] = createResource(
        () => (options.workspaceReady() ? options.tenantID() : undefined),
        async () => Promise.all([getProductSetupSnapshot(), getRoutedOrders()])
    )

    const resetOrders = () => {
        setInProductionCount(0)
        setOpenExceptionCount(0)
        setMockRevenue('$0.00')
        setRealizedMarginTotal('$0.00')
        setPendingSettlementCount(0)
        setDisputedSettlementCount(0)
        setIssueCostExposure('$0.00')
        setShipmentSlaOverdueCount(0)
        setIssueSlaOverdueCount(0)
        setTopPartnerLoad('No partner load yet')
        setIssueRate('0%')
    }

    const applySnapshot = () => {
        setError('')
        const result = snapshot.latest
        if (!result) return
        const [productResult, orderResult] = result
        if (productResult.success) {
            setDraftCount(productResult.data.drafts.length)
            setPublishedCandidateCount(
                productResult.data.candidates.filter((candidate) => candidate.status === 'published_mock').length
            )
        } else {
            setError(productResult.message)
            setDraftCount(0)
            setPublishedCandidateCount(0)
        }
        if (!orderResult.success) {
            setError((current) => current || orderResult.message)
            resetOrders()
            return
        }

        const orders = orderResult.data.orders
        const openIssues = orders.filter(
            (order) => order.exceptionStatus === 'open' || order.exceptionStatus === 'escalated'
        )
        setInProductionCount(orders.filter((order) => order.status === 'in_production').length)
        setOpenExceptionCount(openIssues.length)
        setMockRevenue(formatMoney(orders.reduce((sum, order) => sum + parseMoney(order.total), 0)))
        setRealizedMarginTotal(formatMoney(orders.reduce((sum, order) => sum + parseMoney(order.realizedMargin), 0)))
        setPendingSettlementCount(orders.filter((order) => order.settlementStatus === 'pending').length)
        setDisputedSettlementCount(orders.filter((order) => order.settlementStatus === 'disputed').length)
        setIssueCostExposure(formatMoney(orders.reduce((sum, order) => sum + parseMoney(order.issueCost), 0)))
        setShipmentSlaOverdueCount(
            orders.filter(
                (order) =>
                    !!order.shipmentSlaDueAt &&
                    isOverdue(order.shipmentSlaDueAt) &&
                    order.shipmentStatus !== 'delivered'
            ).length
        )
        setIssueSlaOverdueCount(
            orders.filter(
                (order) =>
                    !!order.issueSlaDueAt &&
                    isOverdue(order.issueSlaDueAt) &&
                    (order.exceptionStatus === 'open' ||
                        order.exceptionStatus === 'escalated' ||
                        order.shipmentStatus === 'delivery_issue')
            ).length
        )
        const loadByPartner = orders.reduce<Record<string, number>>((result, order) => {
            result[order.partner] = (result[order.partner] || 0) + 1
            return result
        }, {})
        const topPartner = Object.entries(loadByPartner).sort((left, right) => right[1] - left[1])[0]
        setTopPartnerLoad(topPartner ? `${topPartner[0]} · ${topPartner[1]} orders` : 'No partner load yet')
        setIssueRate(orders.length > 0 ? `${Math.round((openIssues.length / orders.length) * 100)}%` : '0%')
    }

    createEffect(() => {
        const tenantID = options.tenantID()
        tenantStorage.setTenantID(tenantID)
        const activeId = tokenStorage.getActiveTenantID() || ''
        setActiveTenantId(activeId)
        setTenantReady(activeId === tenantID)
    })

    createEffect(() => {
        applySnapshot()
    })

    return {
        tenantReady,
        activeTenantId,
        draftCount,
        publishedCandidateCount,
        inProductionCount,
        openExceptionCount,
        mockRevenue,
        realizedMarginTotal,
        pendingSettlementCount,
        disputedSettlementCount,
        issueCostExposure,
        shipmentSlaOverdueCount,
        issueSlaOverdueCount,
        topPartnerLoad,
        issueRate,
        error,
    }
}
