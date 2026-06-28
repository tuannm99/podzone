import { TENANT_GQL_URL } from './baseurl'
import { http, type HttpError } from './http'
import { storeStorage } from './storeStorage'
import { tokenStorage } from './tokenStorage'

type GraphQLErrorItem = {
  message?: string
  extensions?: {
    code?: string
    permission?: string
    resource?: string
  }
}

type GraphQLPayload<T> = {
  data?: T
  errors?: GraphQLErrorItem[]
}

function toMessage(error: unknown, fallback: string): string {
  if (
    typeof error === 'object' &&
    error &&
    'message' in error &&
    typeof error.message === 'string'
  ) {
    return error.message
  }
  return fallback
}

function graphQLErrorMessage(error?: GraphQLErrorItem): string {
  if (!error) return 'GraphQL request failed'
  const code = error.extensions?.code
  const permission = error.extensions?.permission
  const resource = error.extensions?.resource

  if (code === 'FORBIDDEN' && permission) {
    return resource && resource !== '*'
      ? `Missing permission: ${permission} on ${resource}`
      : `Missing permission: ${permission}`
  }

  if (code === 'UNAUTHENTICATED') {
    return error.message || 'Authentication is required.'
  }

  return error.message || 'GraphQL request failed'
}

export async function postBackofficeGraphQL<T>(
  query: string,
  variables?: Record<string, unknown>,
  options?: {
    includeStoreHeader?: boolean
  }
): Promise<{ success: true; data: T } | { success: false; message: string }> {
  try {
    const tenantId = tokenStorage.getActiveTenantID()
    const includeStoreHeader = options?.includeStoreHeader ?? true
    const storeId =
      includeStoreHeader && tenantId ? storeStorage.getStoreID(tenantId) : ''
    const { data } = await http.post<GraphQLPayload<T>>(
      TENANT_GQL_URL,
      {
        query,
        variables,
      },
      {
        headers: storeId ? { 'X-Store-ID': storeId } : undefined,
      }
    )

    if (data.errors && data.errors.length > 0) {
      return {
        success: false,
        message: graphQLErrorMessage(data.errors[0]),
      }
    }

    if (!data.data) {
      return {
        success: false,
        message: 'GraphQL request returned no data',
      }
    }

    return { success: true, data: data.data }
  } catch (error) {
    return {
      success: false,
      message: toMessage(error as HttpError, 'GraphQL request failed'),
    }
  }
}
