import { createResource, type Accessor } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  emptyPageInfo,
  type CollectionPage,
  type CollectionQuery,
} from '@/services/collection'

export function createPaginatedResource<T>(
  initialQuery: CollectionQuery,
  fetcher: (query: CollectionQuery) => Promise<CollectionPage<T>>,
  options?: { enabled?: Accessor<boolean> }
) {
  const [query, setQuery] = createStore(initialQuery)
  const [resource, { refetch }] = createResource(
    () => (options?.enabled && !options.enabled() ? undefined : { ...query }),
    fetcher
  )

  const items = () => resource.latest?.items || []
  const pageInfo = () =>
    resource.latest?.pageInfo || emptyPageInfo({ ...query })
  const loading = () => resource.loading
  const error = () =>
    resource.error instanceof Error ? resource.error.message : ''
  const updateQuery = (patch: Partial<CollectionQuery>) => {
    setQuery({ ...patch, page: patch.page ?? 1 })
  }
  const reload = async () => void (await refetch())

  return {
    query,
    items,
    pageInfo,
    loading,
    error,
    updateQuery,
    reload,
  }
}
