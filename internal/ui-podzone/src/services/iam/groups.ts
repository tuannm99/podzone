import { http, type HttpError } from '../http'
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
  tenantId?: string
): Promise<IamResult<GroupInfo[]>> {
  try {
    const params = new URLSearchParams({ scope })
    if (tenantId) params.set('tenantId', tenantId)
    const { data } = await http.get<{ groups?: GroupInfo[] }>(
      `/auth/v1/iam/groups?${params.toString()}`
    )
    return { success: true, data: data.groups || [] }
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
  groupId: number
): Promise<IamResult<number[]>> {
  try {
    const { data } = await http.get<{ userIds?: number[] }>(
      `/auth/v1/iam/groups/${groupId}/members`
    )
    return { success: true, data: data.userIds || [] }
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
  groupId: number
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/groups/${groupId}/policies`
    )
    return { success: true, data: data.policies || [] }
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
  groupId: number
): Promise<IamResult<GroupInlinePolicy[]>> {
  try {
    const { data } = await http.get<{ policies?: GroupInlinePolicy[] }>(
      `/auth/v1/iam/groups/${groupId}/inline-policies`
    )
    return { success: true, data: data.policies || [] }
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
