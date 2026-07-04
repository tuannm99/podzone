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
  OrganizationInfo,
  PolicyInfo,
  CreateOrganizationPayload,
  CreateOrganizationResult,
  OrganizationMembership,
} from './types'

export async function createOrganization(
  payload: CreateOrganizationPayload
): Promise<IamResult<CreateOrganizationResult>> {
  try {
    const { data } = await http.post<CreateOrganizationResult>(
      '/auth/v1/iam/organizations',
      payload
    )
    return { success: true, data }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to create organization')
  }
}

export async function listOrganizations(query: CollectionQuery): Promise<
  IamResult<
    CollectionPage<OrganizationInfo> & {
      canManagePlatform: boolean
    }
  >
> {
  try {
    const { data } = await http.get<{
      organizations?: OrganizationInfo[]
      pageInfo?: WirePageInfo
      canManagePlatform?: boolean
    }>('/auth/v1/iam/organizations', {
      params: toCollectionParams(query),
    })
    return {
      success: true,
      data: {
        items: data.organizations || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
        canManagePlatform: data.canManagePlatform === true,
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load organizations')
  }
}

export async function listOrganizationMembers(
  orgId: string,
  query: CollectionQuery
): Promise<IamResult<CollectionPage<OrganizationMembership>>> {
  try {
    const { data } = await http.get<{
      memberships?: OrganizationMembership[]
      pageInfo?: WirePageInfo
    }>(`/auth/v1/iam/organizations/${orgId}/members`, {
      params: toCollectionParams(query),
    })
    return {
      success: true,
      data: {
        items: data.memberships || [],
        pageInfo: normalizePageInfo(data.pageInfo, query),
      },
    }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to load organization members')
  }
}

export async function addOrganizationMember(
  orgId: string,
  userId: number,
  roleName: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/organizations/${orgId}/members`, {
      orgId,
      userId,
      roleName,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to add organization member')
  }
}

export async function removeOrganizationMember(
  orgId: string,
  userId: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/organizations/${orgId}/members/${userId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(error as HttpError, 'Failed to remove organization member')
  }
}

export async function attachTenantToOrganization(
  orgId: string,
  tenantId: string
): Promise<IamResult<true>> {
  try {
    await http.post(`/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`, {
      orgId,
      tenantId,
    })
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to attach tenant to organization'
    )
  }
}

export async function detachTenantFromOrganization(
  orgId: string,
  tenantId: string
): Promise<IamResult<true>> {
  try {
    await http.delete(`/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`)
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to detach tenant from organization'
    )
  }
}

export async function attachServiceControlPolicy(
  orgId: string,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.post(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies`,
      { orgId, policyName }
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to attach service control policy'
    )
  }
}

export async function detachServiceControlPolicy(
  orgId: string,
  policyName: string
): Promise<IamResult<true>> {
  try {
    await http.delete(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies/${policyName}`
    )
    return { success: true, data: true }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to detach service control policy'
    )
  }
}

export async function listServiceControlPolicies(
  orgId: string
): Promise<IamResult<PolicyInfo[]>> {
  try {
    const { data } = await http.get<{ policies?: PolicyInfo[] }>(
      `/auth/v1/iam/organizations/${orgId}/service-control-policies`
    )
    return { success: true, data: data.policies || [] }
  } catch (error) {
    return toFailure(
      error as HttpError,
      'Failed to load service control policies'
    )
  }
}
