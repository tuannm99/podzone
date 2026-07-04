import { http, type HttpError } from '../http'
import {
  normalizePageInfo,
  toCollectionParams,
  type CollectionPage,
  type CollectionQuery,
  type WirePageInfo,
} from '../collection'
import { toFailure } from './result'
import type {
  IamResult,
  PolicyStatement,
  PolicyInfo,
  GroupInlinePolicy,
  GroupInfo,
  CreateGroupPayload,
} from './types'

export async function createGroup(
  payload: CreateGroupPayload
): Promise<IamResult<{ group?: GroupInfo }>> {
  try {
    const { data } = await http.post<{ group?: GroupInfo }>(
      '/auth/v1/iam/groups',
      payload
    )
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create group')
  }
}

export async function listGroups(
  scope: string,
  orgId: string | undefined,
  tenantId: string | undefined,
  query: CollectionQuery
): Promise<IamResult<CollectionPage<GroupInfo>>> {
  try {
    const { data } = await http.get<{
      groups?: GroupInfo[]
      pageInfo?: WirePageInfo
    }>('/auth/v1/iam/groups', {
      params: {
        scope,
        ...(orgId ? { orgId } : {}),
        ...(tenantId ? { tenantId } : {}),
        ...toCollectionParams(query),
      },
    })
    return {
      success: true,
      data: {
        items: data.groups || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load groups')
  }
}

export async function deleteGroup(groupId: number): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete group')
  }
}

export async function addGroupMember(
  groupId: number,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/groups/${groupId}/members`, {
      groupId,
      userId,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to add group member')
  }
}

export async function listGroupMembers(
  groupId: number,
  query: CollectionQuery
): Promise<IamResult<CollectionPage<number>>> {
  try {
    const { data } = await http.get<{
      userIds?: number[]
      pageInfo?: WirePageInfo
    }>(`/auth/v1/iam/groups/${groupId}/members`, {
      params: toCollectionParams(query),
    })
    return {
      success: true,
      data: {
        items: data.userIds || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load group members')
  }
}

export async function removeGroupMember(
  groupId: number,
  userId: number
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/members/${userId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove group member')
  }
}

export async function attachGroupPolicy(
  groupId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/groups/${groupId}/policies`, {
      groupId,
      policyName,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to attach group policy')
  }
}

export async function listGroupPolicies(
  groupId: number,
  query: CollectionQuery
): Promise<IamResult<CollectionPage<PolicyInfo>>> {
  try {
    const { data } = await http.get<{
      policies?: PolicyInfo[]
      pageInfo?: WirePageInfo
    }>(`/auth/v1/iam/groups/${groupId}/policies`, {
      params: toCollectionParams(query),
    })
    return {
      success: true,
      data: {
        items: data.policies || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load group policies')
  }
}

export async function detachGroupPolicy(
  groupId: number,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/policies/${policyName}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to detach group policy')
  }
}

export async function listGroupInlinePolicies(
  groupId: number,
  query: CollectionQuery
): Promise<IamResult<CollectionPage<GroupInlinePolicy>>> {
  try {
    const { data } = await http.get<{
      policies?: GroupInlinePolicy[]
      pageInfo?: WirePageInfo
    }>(`/auth/v1/iam/groups/${groupId}/inline-policies`, {
      params: toCollectionParams(query),
    })
    return {
      success: true,
      data: {
        items: data.policies || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load group inline policies')
  }
}

export async function putGroupInlinePolicy(
  groupId: number,
  name: string,
  description: string,
  statements: PolicyStatement[]
): Promise<IamResult<true>> {
  try {
    await http.put(`/auth/v1/iam/groups/${groupId}/inline-policies/${name}`, {
      groupId,
      name,
      description,
      statements,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to save group inline policy')
  }
}

export async function deleteGroupInlinePolicy(
  groupId: number,
  name: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/groups/${groupId}/inline-policies/${name}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete group inline policy')
  }
}
