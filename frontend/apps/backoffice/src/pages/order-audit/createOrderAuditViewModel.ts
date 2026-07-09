import { useAuthContext } from '@/solid/context/auth-context'
import { createEffect, createResource, createSignal, on, type Accessor } from 'solid-js'
import { getRoutedOrderActivities, type RoutedOrderActivityFeedEntry } from '@/services/orders'

import { formatFeedSummary, resolveSinceIso, type ActivityFilter, type TimeWindow } from './presentation'

interface OrderAuditViewModelOptions {
    tenantID: Accessor<string>
    storeID: Accessor<string>
    storeLabel: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createOrderAuditViewModel(options: OrderAuditViewModelOptions) {
    const auth = useAuthContext()
    const [entries, setEntries] = createSignal<RoutedOrderActivityFeedEntry[]>([])
    const [nextCursor, setNextCursor] = createSignal<string>()
    const [total, setTotal] = createSignal(0)
    const [activityFilter, setActivityFilter] = createSignal<ActivityFilter>('notes')
    const [hideSystemActivity, setHideSystemActivity] = createSignal(true)
    const [timeWindow, setTimeWindow] = createSignal<TimeWindow>('7d')
    const [actorFilter, setActorFilter] = createSignal('')
    const [orderFilter, setOrderFilter] = createSignal('')
    const [partnerFilter, setPartnerFilter] = createSignal('')
    const [assigneeFilter, setAssigneeFilter] = createSignal('')
    const [appliedActorFilter, setAppliedActorFilter] = createSignal('')
    const [appliedOrderFilter, setAppliedOrderFilter] = createSignal('')
    const [appliedPartnerFilter, setAppliedPartnerFilter] = createSignal('')
    const [appliedAssigneeFilter, setAppliedAssigneeFilter] = createSignal('')
    const [message, setMessage] = createSignal('')
    const [error, setError] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [loadingMore, setLoadingMore] = createSignal(false)
    let requestVersion = 0
    let filterVersion = 0
    const [feedResource] = createResource(
        () =>
            options.workspaceReady()
                ? [
                      options.tenantID(),
                      options.storeID(),
                      activityFilter(),
                      resolveSinceIso(timeWindow()),
                      String(activityFilter() === 'all' ? !hideSystemActivity() : activityFilter() === 'system'),
                      appliedActorFilter(),
                      appliedOrderFilter(),
                      appliedPartnerFilter(),
                      appliedAssigneeFilter(),
                  ].join('|')
                : undefined,
        async () =>
            getRoutedOrderActivities({
                activityType: activityFilter(),
                actorContains: appliedActorFilter(),
                orderId: appliedOrderFilter(),
                partner: appliedPartnerFilter(),
                assignee: appliedAssigneeFilter(),
                since: resolveSinceIso(timeWindow()),
                limit: 50,
                includeSystem: activityFilter() === 'all' ? !hideSystemActivity() : activityFilter() === 'system',
            })
    )

    const loadEntries = async (after?: string, append = false) => {
        const currentRequest = ++requestVersion
        const capturedFilterVersion = filterVersion
        setError('')
        if (append) {
            setLoadingMore(true)
        } else {
            setLoading(true)
            setLoadingMore(false)
        }
        try {
            const result = await getRoutedOrderActivities({
                activityType: activityFilter(),
                actorContains: appliedActorFilter(),
                orderId: appliedOrderFilter(),
                partner: appliedPartnerFilter(),
                assignee: appliedAssigneeFilter(),
                since: resolveSinceIso(timeWindow()),
                limit: 50,
                after,
                includeSystem: activityFilter() === 'all' ? !hideSystemActivity() : activityFilter() === 'system',
            })
            if (currentRequest !== requestVersion) return
            if (capturedFilterVersion !== filterVersion) return
            if (!result.success) {
                setError(result.message)
                if (!append) {
                    setEntries([])
                    setNextCursor(undefined)
                    setTotal(0)
                }
                return
            }
            setEntries((current) => (append ? [...current, ...result.data.entries] : result.data.entries))
            setNextCursor(result.data.nextCursor)
            setTotal(result.data.total)
        } finally {
            if (currentRequest === requestVersion) {
                setLoading(false)
                setLoadingMore(false)
            }
        }
    }
    const auditFeed = () => entries()
    const copyFeed = async () => {
        try {
            await navigator.clipboard.writeText(formatFeedSummary(options.storeLabel(), entries()))
            setMessage(`Copied audit feed for ${options.storeLabel()}.`)
        } catch {
            setError('Could not copy audit feed to clipboard.')
        }
    }
    const applyFilters = () => {
        filterVersion++
        setAppliedActorFilter(actorFilter().trim())
        setAppliedOrderFilter(orderFilter().trim())
        setAppliedPartnerFilter(partnerFilter().trim())
        setAppliedAssigneeFilter(assigneeFilter().trim())
    }
    const loadMore = async () => {
        if (nextCursor()) await loadEntries(nextCursor(), true)
    }

    createEffect(() => {
        auth.setActiveTenantId(options.tenantID())
    })

    createEffect(() => {
        setLoading(feedResource.loading)
        if (feedResource.loading) return
        const result = feedResource.latest
        if (!result) return
        if (!result.success) {
            setError(result.message)
            setEntries([])
            setNextCursor(undefined)
            setTotal(0)
            return
        }
        setError('')
        setEntries(result.data.entries)
        setNextCursor(result.data.nextCursor)
        setTotal(result.data.total)
    })

    createEffect(
        on(
            [() => options.tenantID(), () => options.storeID()],
            () => {
                setActorFilter('')
                setOrderFilter('')
                setPartnerFilter('')
                setAssigneeFilter('')
                setAppliedActorFilter('')
                setAppliedOrderFilter('')
                setAppliedPartnerFilter('')
                setAppliedAssigneeFilter('')
            },
            { defer: true }
        )
    )

    return {
        entries,
        nextCursor,
        total,
        activityFilter,
        setActivityFilter,
        hideSystemActivity,
        setHideSystemActivity,
        timeWindow,
        setTimeWindow,
        actorFilter,
        setActorFilter,
        orderFilter,
        setOrderFilter,
        partnerFilter,
        setPartnerFilter,
        assigneeFilter,
        setAssigneeFilter,
        message,
        error,
        loading,
        loadingMore,
        applyFilters,
        auditFeed,
        copyFeed,
        loadMore,
    }
}
