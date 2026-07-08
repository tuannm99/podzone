import { createEffect, createResource, createSignal, type Accessor } from 'solid-js'
import { getRoutedOrderActivities, type RoutedOrderActivityFeedEntry } from '@/services/orders'
import { tenantStorage } from '@/services/tenantStorage'
import { formatFeedSummary, resolveSinceIso, type ActivityFilter, type TimeWindow } from './presentation'

interface OrderAuditViewModelOptions {
    tenantID: Accessor<string>
    storeID: Accessor<string>
    storeLabel: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createOrderAuditViewModel(options: OrderAuditViewModelOptions) {
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
    const [message, setMessage] = createSignal('')
    const [error, setError] = createSignal('')
    const [loading, setLoading] = createSignal(false)
    const [loadingMore, setLoadingMore] = createSignal(false)
    let requestVersion = 0
    const [feedResource] = createResource(
        () =>
            options.workspaceReady()
                ? {
                      tenantID: options.tenantID(),
                      storeID: options.storeID(),
                      activityType: activityFilter(),
                      actorContains: actorFilter().trim(),
                      orderId: orderFilter().trim(),
                      partner: partnerFilter().trim(),
                      assignee: assigneeFilter().trim(),
                      since: resolveSinceIso(timeWindow()),
                      includeSystem: activityFilter() === 'all' ? !hideSystemActivity() : activityFilter() === 'system',
                  }
                : undefined,
        async (source) =>
            getRoutedOrderActivities({
                activityType: source.activityType,
                actorContains: source.actorContains,
                orderId: source.orderId,
                partner: source.partner,
                assignee: source.assignee,
                since: source.since,
                limit: 50,
                includeSystem: source.includeSystem,
            })
    )

    const loadEntries = async (after?: string, append = false) => {
        const currentRequest = ++requestVersion
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
                actorContains: actorFilter().trim(),
                orderId: orderFilter().trim(),
                partner: partnerFilter().trim(),
                assignee: assigneeFilter().trim(),
                since: resolveSinceIso(timeWindow()),
                limit: 50,
                after,
                includeSystem: activityFilter() === 'all' ? !hideSystemActivity() : activityFilter() === 'system',
            })
            if (currentRequest !== requestVersion) return
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
    const loadMore = async () => {
        if (nextCursor()) await loadEntries(nextCursor(), true)
    }

    createEffect(() => {
        tenantStorage.setTenantID(options.tenantID())
    })

    createEffect(() => {
        setLoading(feedResource.loading)
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
        loadEntries,
        auditFeed,
        copyFeed,
        loadMore,
    }
}
