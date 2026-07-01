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
  PolicyVersionInfo,
  PolicyAttachmentInfo,
  CreatePolicyPayload,
  CreatePolicyVersionPayload,
} from './types'

export async function createPolicy(
  payload: CreatePolicyPayload
): Promise<IamResult<{ policy?: PolicyInfo; statements?: PolicyStatement[] }>> {
  try {
    const { data } = await http.post<{
      policy?: PolicyInfo
      statements?: PolicyStatement[]
    }>('/auth/v1/iam/policies', payload)
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create policy')
  }
}

export async function listPolicies(
  query: CollectionQuery,
  scope?: string
): Promise<IamResult<CollectionPage<PolicyInfo>>> {
  try {
    const { data } = await http.get<{
      policies?: PolicyInfo[]
      pageInfo?: WirePageInfo
    }>('/auth/v1/iam/policies', {
      params: {
        ...toCollectionParams(query),
        ...(scope ? { scope } : {}),
      },
    })
    return {
      success: true,
      data: {
        items: data.policies || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policies')
  }
}

export async function getPolicy(name: string): Promise<IamResult<PolicyInfo>> {
  try {
    const { data } = await http.get<{ policy?: PolicyInfo }>(
      `/auth/v1/iam/policies/${name}`
    )
    if (!data.policy) {
      return { success: false, message: 'Policy not found' }
    }
    return { success: true, data: data.policy }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy')
  }
}

export async function createPolicyVersion(
  payload: CreatePolicyVersionPayload
): Promise<
  IamResult<{
    policyVersion?: PolicyVersionInfo
    statements?: PolicyStatement[]
  }>
> {
  try {
    const { data } = await http.post<{
      policyVersion?: PolicyVersionInfo
      statements?: PolicyStatement[]
    }>(`/auth/v1/iam/policies/${payload.name}/versions`, payload)
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create policy version')
  }
}

export async function listPolicyVersions(
  name: string
): Promise<IamResult<PolicyVersionInfo[]>> {
  try {
    const { data } = await http.get<{ versions?: PolicyVersionInfo[] }>(
      `/auth/v1/iam/policies/${name}/versions`
    )
    return { success: true, data: data.versions || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy versions')
  }
}

export async function setDefaultPolicyVersion(
  name: string,
  version: string
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/policies/${name}/versions/${version}:set-default`,
      {
        name,
        version,
      }
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to set default policy version')
  }
}

export async function deletePolicyVersion(
  name: string,
  version: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/policies/${name}/versions/${version}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete policy version')
  }
}

export async function listPolicyAttachments(
  name: string
): Promise<IamResult<PolicyAttachmentInfo[]>> {
  try {
    const { data } = await http.get<{ attachments?: PolicyAttachmentInfo[] }>(
      `/auth/v1/iam/policies/${name}/attachments`
    )
    return { success: true, data: data.attachments || [] }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load policy attachments')
  }
}

export async function deletePolicy(name: string): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/policies/${name}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to delete policy')
  }
}
