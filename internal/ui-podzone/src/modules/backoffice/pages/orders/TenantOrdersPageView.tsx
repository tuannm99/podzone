import { useParams, useSearch } from '@tanstack/solid-router'
import { createEffect, createSignal } from 'solid-js'
import { emptyPageInfo } from '@/services/collection'
import {
    getRoutedOrderPage,
    getRoutedOrderRecommendation,
    getRoutedOrders,
    type RoutedOrder,
    type RoutedOrderRecommendation,
} from '@/services/orders'
import { getProductSetupSnapshot, type CatalogCandidate } from '@/services/productSetup'
import { tenantStorage } from '@/services/tenantStorage'
import { createFormStore, required } from '@/solid/forms'
import { useTenantWorkspace } from '@/solid/workspace/context'
import type { QueueSort, QueueView, TenantOrdersBoardContextValue } from './board-context'
import type { TenantOrdersComposerContextValue } from './composer-context'
import type { TenantOrdersInsightsContextValue } from './context'
import { TenantOrdersView } from './TenantOrdersView'
import type { ActivityFilter } from './order-card/types'
import { isQueueSort, isQueueView } from './presentation'
import { positiveInteger, routedOrderInitialValues } from './forms'
import { useOrderActions } from './useOrderActions'
import { useOrderDrafts } from './useOrderDrafts'
import { useOrderInsights } from './useOrderInsights'
import { useOrderStorage } from './useOrderStorage'

export function TenantOrdersPageView() {
    const params = useParams({ from: '/t/$tenantId/orders' })
    const search = useSearch({ strict: false }) as () => Record<string, unknown>
    const workspace = useTenantWorkspace()

    const [availableCandidates, setAvailableCandidates] = createSignal<CatalogCandidate[]>([])
    const [orders, setOrders] = createSignal<RoutedOrder[]>([])
    const [queueOrders, setQueueOrders] = createSignal<RoutedOrder[]>([])
    const [queuePage, setQueuePage] = createSignal(1)
    const [queuePageInfo, setQueuePageInfo] = createSignal(emptyPageInfo({ page: 1, pageSize: 10 }))
    const [queueSearch, setQueueSearch] = createSignal('')
    const [appliedQueueSearch, setAppliedQueueSearch] = createSignal('')
    const orderForm = createFormStore({
        initialValues: routedOrderInitialValues,
        validators: {
            selectedCandidateId: [required('Choose a published product.')],
            customerName: [required('Enter the customer name.')],
            quantity: [positiveInteger('Quantity must be a positive whole number.')],
        },
    })
    const [routingRecommendation, setRoutingRecommendation] = createSignal<RoutedOrderRecommendation | null>(null)
    const [activeQueueView, setActiveQueueView] = createSignal<QueueView>('all')
    const [activeQueueSort, setActiveQueueSort] = createSignal<QueueSort>('priority')
    const [activityFilter, setActivityFilter] = createSignal<ActivityFilter>('notes')
    const [hideSystemActivity, setHideSystemActivity] = createSignal(true)
    const [operatorLens, setOperatorLens] = createSignal('')
    const [message, setMessage] = createSignal('')
    const [error, setError] = createSignal('')

    const currentStoreId = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreId() || 'selected store'
    const effectivePreferredPartner = () =>
        orderForm.values.manualPartnerOverride ? orderForm.values.preferredPartner.trim() : ''

    const applyPreferredPartnerOverride = (partnerName: string) => {
        orderForm.setValue('manualPartnerOverride', true)
        orderForm.setValue('preferredPartner', partnerName)
    }

    const resetPreferredPartnerOverride = () => {
        orderForm.setValue('manualPartnerOverride', false)
        orderForm.setValue('preferredPartner', '')
    }

    const drafts = useOrderDrafts()
    const insights = useOrderInsights({
        orders,
        storeLabel,
        activeQueueView,
        activeQueueSort,
        operatorLens,
        activityFilter,
        hideSystemActivity,
        setMessage,
        setError,
    })
    const storage = useOrderStorage({
        tenantId: () => params().tenantId,
        storeId: currentStoreId,
        activeQueueView,
        setActiveQueueView,
        activeQueueSort,
        setActiveQueueSort,
        operatorLens,
        setOperatorLens,
        setMessage,
    })
    const actions = useOrderActions({
        orders,
        setOrders,
        availableCandidates,
        orderForm,
        effectivePreferredPartner,
        selectedOrderIDs: storage.selectedOrderIDs,
        setSelectedOrderIDs: storage.setSelectedOrderIDs,
        bulkDraft: storage.bulkDraft,
        setBulkDraft: storage.setBulkDraft,
        drafts,
        setMessage,
        setError,
        onChanged: async () => {
            await Promise.all([loadOrders(), loadQueuePage()])
        },
    })

    let candidatesRequestVersion = 0
    const loadCandidates = async () => {
        const currentRequest = ++candidatesRequestVersion
        const result = await getProductSetupSnapshot()
        if (currentRequest !== candidatesRequestVersion) return
        if (!result.success) {
            setError(result.message)
            setAvailableCandidates([])
            return
        }
        const published = result.data.candidates.filter((candidate) => candidate.status === 'published_mock')
        setAvailableCandidates(published)
        const selectionStillExists = published.some(
            (candidate) => candidate.id === orderForm.values.selectedCandidateId
        )
        if (!selectionStillExists) {
            orderForm.setValue('selectedCandidateId', published[0]?.id || '')
        }
    }

    let ordersRequestVersion = 0
    async function loadOrders() {
        const currentRequest = ++ordersRequestVersion
        const result = await getRoutedOrders()
        if (currentRequest !== ordersRequestVersion) return
        if (!result.success) {
            setError(result.message)
            setOrders([])
            return
        }
        setOrders(result.data.orders)
        if (!operatorLens().trim()) {
            const firstAssigned = result.data.orders.find(
                (order) => order.operatorAssignee && order.operatorAssignee !== 'unassigned'
            )
            if (firstAssigned) {
                setOperatorLens(firstAssigned.operatorAssignee)
            }
        }
    }

    let queueRequestVersion = 0
    async function loadQueuePage() {
        const currentRequest = ++queueRequestVersion
        const filters = [
            ...(activeQueueView() !== 'all'
                ? [
                      {
                          field: 'queueView',
                          operator: 'FILTER_OPERATOR_EQ' as const,
                          values: [activeQueueView()],
                      },
                  ]
                : []),
            ...(activeQueueView() === 'my_queue' && operatorLens().trim()
                ? [
                      {
                          field: 'operatorAssignee',
                          operator: 'FILTER_OPERATOR_EQ' as const,
                          values: [operatorLens().trim()],
                      },
                  ]
                : []),
        ]
        const result = await getRoutedOrderPage({
            page: queuePage(),
            pageSize: 10,
            search: appliedQueueSearch(),
            filters,
            sortBy: activeQueueSort() === 'priority' ? 'priority' : 'createdAt',
            sortDirection: 'SORT_DIRECTION_DESC',
        })
        if (currentRequest !== queueRequestVersion) return
        if (!result.success) {
            setError(result.message)
            setQueueOrders([])
            setQueuePageInfo(emptyPageInfo({ page: queuePage(), pageSize: 10 }))
            return
        }
        setQueueOrders(result.data.items)
        setQueuePageInfo(result.data.pageInfo)
    }

    let recommendationRequestVersion = 0
    const loadRoutingRecommendation = async () => {
        const currentRequest = ++recommendationRequestVersion
        const candidateID = orderForm.values.selectedCandidateId.trim()
        if (!candidateID) {
            setRoutingRecommendation(null)
            return
        }
        const result = await getRoutedOrderRecommendation({
            candidateId: candidateID,
            productType: orderForm.values.selectedProductType,
            shipRegion: orderForm.values.selectedShipRegion,
            preferredPartner: orderForm.values.preferredPartner.trim() || undefined,
        })
        if (currentRequest !== recommendationRequestVersion) return
        if (!result.success) {
            setError(result.message)
            setRoutingRecommendation(null)
            return
        }
        setRoutingRecommendation(result.data)
    }

    const composerContextValue: TenantOrdersComposerContextValue = {
        availableCandidates,
        form: orderForm,
        routingRecommendation,
        applyPreferredPartnerOverride,
        resetPreferredPartnerOverride,
        createMockOrder: actions.createMockOrder,
    }

    const boardContextValue: TenantOrdersBoardContextValue = {
        activeQueueView,
        setActiveQueueView,
        activeQueueSort,
        setActiveQueueSort,
        operatorLens,
        setOperatorLens,
        queueSearch,
        setQueueSearch,
        applyQueueSearch: () => {
            setQueuePage(1)
            setAppliedQueueSearch(queueSearch().trim())
        },
        queueViewCount: insights.queueViewCount,
        savedPresets: storage.savedPresets,
        presetName: storage.presetName,
        setPresetName: storage.setPresetName,
        saveQueuePreset: storage.saveQueuePreset,
        applyQueuePreset: storage.applyQueuePreset,
        deleteQueuePreset: storage.deleteQueuePreset,
        selectedOrderIDs: storage.selectedOrderIDs,
        selectVisibleOrders: () => storage.setSelectedOrderIDs(queueOrders().map((order) => order.id)),
        clearSelectedOrders: storage.clearSelectedOrders,
        bulkDraft: storage.bulkDraft,
        setBulkDraft: storage.setBulkDraft,
        applyRelativeShipmentSla: storage.applyRelativeShipmentSla,
        savedBulkTemplates: storage.savedBulkTemplates,
        bulkTemplateName: storage.bulkTemplateName,
        setBulkTemplateName: storage.setBulkTemplateName,
        saveBulkTemplate: storage.saveBulkTemplate,
        applyBulkTemplate: storage.applyBulkTemplate,
        deleteBulkTemplate: storage.deleteBulkTemplate,
        applyBulkUpdate: actions.applyBulkUpdate,
    }

    const insightsContextValue: TenantOrdersInsightsContextValue = {
        tenantId: params().tenantId,
        storeId: currentStoreId,
        storeLabel,
        blockedOrders: insights.blockedOrders,
        blockedReasonSummary: insights.blockedReasonSummary,
        forcedRerouteSummary: insights.forcedRerouteSummary,
        reconciliationOrders: insights.reconciliationOrders,
        partnerFinanceSummary: insights.partnerFinanceSummary,
        storeActivityFeed: insights.storeActivityFeed,
        copyStoreActivityFeed: insights.copyStoreActivityFeed,
    }

    createEffect(() => {
        tenantStorage.setTenantID(params().tenantId)
        if (!workspaceReady()) {
            return
        }
        storage.loadSavedPresets()
        storage.loadSavedBulkTemplates()
        void loadCandidates()
        void loadOrders()
    })

    let previousQueueKey = ''
    createEffect(() => {
        const queueKey = [
            params().tenantId,
            currentStoreId(),
            activeQueueView(),
            activeQueueSort(),
            operatorLens(),
            appliedQueueSearch(),
        ].join('|')
        if (!workspaceReady()) return
        if (previousQueueKey && previousQueueKey !== queueKey && queuePage() !== 1) {
            previousQueueKey = queueKey
            setQueuePage(1)
            return
        }
        previousQueueKey = queueKey
        void queuePage()
        void loadQueuePage()
    })

    createEffect(() => {
        const current = search()
        const queueView = String(current.queueView || '')
        const queueSort = String(current.queueSort || '')
        const lens = String(current.operatorLens || '')

        if (isQueueView(queueView)) {
            setActiveQueueView(queueView)
        }
        if (isQueueSort(queueSort)) {
            setActiveQueueSort(queueSort)
        }
        if (lens) {
            setOperatorLens(lens)
        }
    })

    createEffect(() => {
        if (!workspaceReady()) {
            setRoutingRecommendation(null)
            return
        }
        void loadRoutingRecommendation()
    })

    return (
        <TenantOrdersView
            storeLabel={storeLabel}
            workspaceReady={workspaceReady}
            message={message}
            error={error}
            orders={orders}
            queueOrders={queueOrders}
            queuePageInfo={queuePageInfo}
            queuePage={queuePage}
            setQueuePage={setQueuePage}
            activityFilter={activityFilter}
            setActivityFilter={setActivityFilter}
            hideSystemActivity={hideSystemActivity}
            setHideSystemActivity={setHideSystemActivity}
            composerContextValue={composerContextValue}
            boardContextValue={boardContextValue}
            insightsContextValue={insightsContextValue}
            storage={storage}
            actions={actions}
            drafts={drafts}
            insights={insights}
        />
    )
}
