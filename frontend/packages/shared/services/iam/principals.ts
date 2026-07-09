import { http, type HttpError } from '../http'
import {
    normalizePageInfo,
    toCollectionParams,
    type CollectionPage,
    type CollectionQuery,
    type WirePageInfo,
} from '../collection'
import { toFailure } from './result'
import type { IamResult, PolicyStatement, PolicyInfo, UserInlinePolicy, PermissionBoundary } from './types'

export async function listPlatformUserPolicies(
    targetUserId: number,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<PolicyInfo>>> {
    try {
        const { data } = await http.get<{
            policies?: PolicyInfo[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/platform-users/${targetUserId}/policies`, {
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
        return toFailure(error as HttpError, 'Failed to load platform user policies')
    }
}

export async function attachPlatformUserPolicy(targetUserId: number, policyName: string): Promise<IamResult<true>> {
    try {
        await http.post(`/auth/v1/iam/platform-users/${targetUserId}/policies`, {
            targetUserId,
            policyName,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to attach platform user policy')
    }
}

export async function detachPlatformUserPolicy(targetUserId: number, policyName: string): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/platform-users/${targetUserId}/policies/${policyName}`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to detach platform user policy')
    }
}

export async function listPlatformUserInlinePolicies(
    targetUserId: number,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<UserInlinePolicy>>> {
    try {
        const { data } = await http.get<{
            policies?: UserInlinePolicy[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/platform-users/${targetUserId}/inline-policies`, {
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
        return toFailure(error as HttpError, 'Failed to load platform user inline policies')
    }
}

export async function putPlatformUserInlinePolicy(
    targetUserId: number,
    name: string,
    description: string,
    statements: PolicyStatement[]
): Promise<IamResult<true>> {
    try {
        await http.put(`/auth/v1/iam/platform-users/${targetUserId}/inline-policies/${name}`, {
            targetUserId,
            name,
            description,
            statements,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to save platform user inline policy')
    }
}

export async function deletePlatformUserInlinePolicy(targetUserId: number, name: string): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/platform-users/${targetUserId}/inline-policies/${name}`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete platform user inline policy')
    }
}

export async function putPlatformUserPermissionBoundary(
    targetUserId: number,
    policyName: string
): Promise<IamResult<true>> {
    try {
        await http.put(`/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`, {
            targetUserId,
            policyName,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to save platform user boundary')
    }
}

export async function getPlatformUserPermissionBoundary(
    targetUserId: number
): Promise<IamResult<PermissionBoundary | null>> {
    try {
        const { data } = await http.get<{ boundary?: PermissionBoundary }>(
            `/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`
        )
        return { success: true, data: data.boundary || null }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to load platform user boundary')
    }
}

export async function deletePlatformUserPermissionBoundary(targetUserId: number): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/platform-users/${targetUserId}/permission-boundary`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete platform user boundary')
    }
}

export async function listTenantUserPolicies(
    tenantId: string,
    userId: number,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<PolicyInfo>>> {
    try {
        const { data } = await http.get<{
            policies?: PolicyInfo[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies`, {
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
        return toFailure(error as HttpError, 'Failed to load tenant user policies')
    }
}

export async function attachTenantUserPolicy(
    tenantId: string,
    userId: number,
    policyName: string
): Promise<IamResult<true>> {
    try {
        await http.post(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies`, {
            tenantId,
            userId,
            policyName,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to attach tenant user policy')
    }
}

export async function detachTenantUserPolicy(
    tenantId: string,
    userId: number,
    policyName: string
): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/policies/${policyName}`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to detach tenant user policy')
    }
}

export async function listTenantUserInlinePolicies(
    tenantId: string,
    userId: number,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<UserInlinePolicy>>> {
    try {
        const { data } = await http.get<{
            policies?: UserInlinePolicy[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies`, {
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
        return toFailure(error as HttpError, 'Failed to load tenant user inline policies')
    }
}

export async function putTenantUserInlinePolicy(
    tenantId: string,
    userId: number,
    name: string,
    description: string,
    statements: PolicyStatement[]
): Promise<IamResult<true>> {
    try {
        await http.put(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies/${name}`, {
            tenantId,
            userId,
            name,
            description,
            statements,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to save tenant user inline policy')
    }
}

export async function deleteTenantUserInlinePolicy(
    tenantId: string,
    userId: number,
    name: string
): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/inline-policies/${name}`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete tenant user inline policy')
    }
}

export async function putTenantUserPermissionBoundary(
    tenantId: string,
    userId: number,
    policyName: string
): Promise<IamResult<true>> {
    try {
        await http.put(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`, {
            tenantId,
            userId,
            policyName,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to save tenant user boundary')
    }
}

export async function getTenantUserPermissionBoundary(
    tenantId: string,
    userId: number
): Promise<IamResult<PermissionBoundary | null>> {
    try {
        const { data } = await http.get<{ boundary?: PermissionBoundary }>(
            `/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`
        )
        return { success: true, data: data.boundary || null }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to load tenant user boundary')
    }
}

export async function deleteTenantUserPermissionBoundary(tenantId: string, userId: number): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/tenants/${tenantId}/members/${userId}/permission-boundary`)
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete tenant user boundary')
    }
}
