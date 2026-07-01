import { createResource } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  emptyPageInfo,
  type CollectionPage,
  type CollectionQuery,
} from '@/services/collection'

export function createPaginatedResource<T>(
  initialQuery: CollectionQuery,
  fetcher: (query: CollectionQuery) => Promise<CollectionPage<T>>
) {
  const [query, setQuery] = createStore(initialQuery)
  const [resource, { refetch }] = createResource(() => ({ ...query }), fetcher)

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
