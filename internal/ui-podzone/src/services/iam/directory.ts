import {
  normalizePageInfo,
  toCollectionParams,
  type CollectionPage,
  type CollectionQuery,
  type WirePageInfo,
} from '../collection'
import { http, type HttpError } from '../http'
import { toFailure } from './result'
import type {
  DirectoryScope,
  DirectoryUser,
  IamResult,
  PermissionInfo,
} from './types'

function scopeParams(scope: DirectoryScope) {
  return {
    scope: scope.scope,
    ...(scope.orgId ? { orgId: scope.orgId } : {}),
    ...(scope.tenantId ? { tenantId: scope.tenantId } : {}),
  }
}

export async function listDirectoryUsers(
  query: CollectionQuery,
  scope: DirectoryScope
): Promise<IamResult<CollectionPage<DirectoryUser>>> {
  try {
    const { data } = await http.get<{
      users?: DirectoryUser[]
      pageInfo?: WirePageInfo
    }>('/auth/v1/iam/directory/users', {
      params: {
        ...scopeParams(scope),
        ...toCollectionParams(query),
      },
    })
    return {
      success: true,
      data: {
        items: data.users || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load users')
  }
}

export async function listPermissions(
  query: CollectionQuery,
  scope: DirectoryScope
): Promise<IamResult<CollectionPage<PermissionInfo>>> {
  try {
    const { data } = await http.get<{
      permissions?: PermissionInfo[]
      pageInfo?: WirePageInfo
    }>('/auth/v1/iam/permissions', {
      params: {
        ...scopeParams(scope),
        ...toCollectionParams(query),
      },
    })
    return {
      success: true,
      data: {
        items: data.permissions || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load permissions')
  }
}
