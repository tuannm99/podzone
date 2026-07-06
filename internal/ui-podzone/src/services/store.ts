import { postBackofficeGraphQL } from './backofficeGraphql'
import {
  normalizePageInfo,
  toGraphQLCollectionInput,
  type CollectionPage,
  type CollectionQuery,
  type WirePageInfo,
} from './collection'

export type StoreInfo = {
  id: string
  name: string
  ownerId: string
  isActive: boolean
  description: string
  status: string
  createdAt?: string
  updatedAt?: string
}

type StoreResult<T> =
  { success: true; data: T } | { success: false; message: string }

export async function listStores(
  query: CollectionQuery
): Promise<StoreResult<CollectionPage<StoreInfo>>> {
  const result = await postBackofficeGraphQL<{
    stores: {
      items?: StoreInfo[]
      pageInfo?: WirePageInfo
    }
  }>(
    `
    query Stores($collection: CollectionInput) {
      stores(collection: $collection) {
        items {
          id
          name
          ownerId: owner_id
          isActive: is_active
          description
          status
          createdAt: created_at
          updatedAt: updated_at
        }
        pageInfo {
          total
          page
          pageSize
          totalPages
          hasNext
          hasPrevious
        }
      }
    }
  `,
    { collection: toGraphQLCollectionInput(query) },
    { includeStoreHeader: false }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return {
    success: true,
    data: {
      items: result.data.stores?.items || [],
      pageInfo: normalizePageInfo(result.data.stores?.pageInfo, query),
    },
  }
}

export async function listAllStores(): Promise<StoreResult<StoreInfo[]>> {
  const stores: StoreInfo[] = []
  for (let page = 1; ; page += 1) {
    const result = await listStores({
      page,
      pageSize: 100,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    })
    if (!result.success) return result
    stores.push(...result.data.items)
    if (!result.data.pageInfo.hasNext) break
  }
  return { success: true, data: stores }
}

export async function getStore(id: string): Promise<StoreResult<StoreInfo>> {
  const result = await postBackofficeGraphQL<{ store?: StoreInfo }>(
    `
    query Store($id: ID!) {
      store(id: $id) {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    { id },
    { includeStoreHeader: false }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  if (!result.data.store) {
    return { success: false, message: 'Store not found' }
  }
  return { success: true, data: result.data.store }
}

export async function createStore(input: {
  name: string
  description?: string
}): Promise<StoreResult<StoreInfo>> {
  const result = await postBackofficeGraphQL<{ createStore: StoreInfo }>(
    `
    mutation CreateStore($input: CreateStoreInput!) {
      createStore(input: $input) {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    {
      input: {
        name: input.name,
        description: input.description || '',
      },
    },
    { includeStoreHeader: false }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.createStore }
}

export async function activateStore(
  id: string
): Promise<StoreResult<StoreInfo>> {
  const result = await postBackofficeGraphQL<{ activateStore: StoreInfo }>(
    `
    mutation ActivateStore($id: ID!) {
      activateStore(id: $id) {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
    }
  `,
    { id },
    { includeStoreHeader: false }
  )
  if (!result.success) {
    return { success: false, message: result.message }
  }
  return { success: true, data: result.data.activateStore }
}
