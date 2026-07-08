import { useParams, useSearch } from '@tanstack/solid-router'
import { createEffect, createResource, createSignal } from 'solid-js'
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
import type { ActivityFilter } from './order-card/types'
import { isQueueSort, isQueueView } from './presentation'
import { positiveInteger, routedOrderInitialValues } from './forms'
import { useOrderActions } from './useOrderActions'
import { useOrderDrafts } from './useOrderDrafts'
import { useOrderInsights } from './useOrderInsights'
import { useOrderStorage } from './useOrderStorage'

export function createTenantOrdersViewModel() {
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
            await Promise.all([reloadOrders(), reloadQueuePage()])
        },
    })

    const [candidatesResource] = createResource(
        () => (workspaceReady() ? `${params().tenantId}|${currentStoreId()}` : undefined),
        async () => getProductSetupSnapshot()
    )
    const [ordersResource, { refetch: reloadOrders }] = createResource(
        () => (workspaceReady() ? `${params().tenantId}|${currentStoreId()}` : undefined),
        async () => getRoutedOrders()
    )
    const queueFilters = () => [
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
    const [queuePageResource, { refetch: reloadQueuePage }] = createResource(
        () =>
            workspaceReady()
                ? {
                      tenantId: params().tenantId,
                      storeId: currentStoreId(),
                      page: queuePage(),
                      search: appliedQueueSearch(),
                      filters: queueFilters(),
                      sortBy: activeQueueSort() === 'priority' ? 'priority' : 'createdAt',
                  }
                : undefined,
        async (source) =>
            getRoutedOrderPage({
                page: source.page,
                pageSize: 10,
                search: source.search,
                filters: source.filters,
                sortBy: source.sortBy,
                sortDirection: 'SORT_DIRECTION_DESC',
            })
    )
    const [recommendationResource] = createResource(
        () => {
            if (!workspaceReady()) return undefined
            const candidateID = orderForm.values.selectedCandidateId.trim()
            if (!candidateID) return undefined
            return {
                candidateId: candidateID,
                productType: orderForm.values.selectedProductType,
                shipRegion: orderForm.values.selectedShipRegion,
                preferredPartner: orderForm.values.preferredPartner.trim() || undefined,
            }
        },
        async (source) => getRoutedOrderRecommendation(source)
    )

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
    })

    createEffect(() => {
        const result = candidatesResource.latest
        if (!result) return
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
    })

    createEffect(() => {
        const result = ordersResource.latest
        if (!result) return
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
    })

    createEffect(() => {
        const result = queuePageResource.latest
        if (!result) return
        if (!result.success) {
            setError(result.message)
            setQueueOrders([])
            setQueuePageInfo(emptyPageInfo({ page: queuePage(), pageSize: 10 }))
            return
        }
        setQueueOrders(result.data.items)
        setQueuePageInfo(result.data.pageInfo)
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
        const result = recommendationResource.latest
        if (!workspaceReady() || !result) {
            setRoutingRecommendation(null)
            return
        }
        if (!result.success) {
            setError(result.message)
            setRoutingRecommendation(null)
            return
        }
        setRoutingRecommendation(result.data)
    })

    return {
        storeLabel,
        workspaceReady,
        message,
        error,
        orders,
        queueOrders,
        queuePageInfo,
        queuePage,
        setQueuePage,
        activityFilter,
        setActivityFilter,
        hideSystemActivity,
        setHideSystemActivity,
        composerContextValue,
        boardContextValue,
        insightsContextValue,
        storage,
        actions,
        drafts,
        insights,
    }
}

export type TenantOrdersViewModel = ReturnType<typeof createTenantOrdersViewModel>
