import {
  createEffect,
  createResource,
  createSignal,
  on,
  type Accessor,
} from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  emptyPageInfo,
  type CollectionPage,
  type CollectionQuery,
} from '@/services/collection'

export function createPaginatedResource<T>(
  initialQuery: CollectionQuery,
  fetcher: (query: CollectionQuery) => Promise<CollectionPage<T>>,
  options?: {
    enabled?: Accessor<boolean>
    dependency?: Accessor<unknown>
  }
) {
  type ResourceSource = {
    query: CollectionQuery
    dependency: unknown
  }

  const [query, setQuery] = createStore(initialQuery)
  const [readError, setReadError] = createSignal('')
  const [resource, { mutate, refetch }] = createResource<
    CollectionPage<T>,
    ResourceSource
  >(
    () =>
      options?.enabled && !options.enabled()
        ? undefined
        : {
            query: { ...query },
            dependency: options?.dependency?.(),
          },
    async ({ query: currentQuery }, info): Promise<CollectionPage<T>> => {
      try {
        const page = await fetcher(currentQuery)
        setReadError('')
        return page
      } catch (error) {
        setReadError(
          error instanceof Error ? error.message : 'Failed to load collection'
        )
        return (
          info.value || {
            items: [],
            pageInfo: emptyPageInfo(currentQuery),
          }
        )
      }
    }
  )
  if (options?.dependency) {
    createEffect(
      on(options.dependency, () => setQuery('page', 1), { defer: true })
    )
  }

  const items = () => resource.latest?.items || []
  const pageInfo = () =>
    resource.latest?.pageInfo || emptyPageInfo({ ...query })
  const loading = () => resource.loading
  const resolved = () =>
    !readError() &&
    (resource.state === 'ready' || resource.state === 'refreshing')
  const error = readError
  const updateQuery = (patch: Partial<CollectionQuery>) => {
    setQuery({ ...patch, page: patch.page ?? 1 })
  }
  const reload = async () => void (await refetch())
  const clear = () => {
    setReadError('')
    mutate(undefined)
  }

  return {
    query,
    items,
    pageInfo,
    loading,
    resolved,
    error,
    updateQuery,
    reload,
    clear,
  }
}
