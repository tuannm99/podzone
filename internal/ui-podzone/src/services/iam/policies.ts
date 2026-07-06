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
    PolicyLocator,
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
    scope?: string,
    orgId?: string
): Promise<IamResult<CollectionPage<PolicyInfo>>> {
    try {
        const { data } = await http.get<{
            policies?: PolicyInfo[]
            pageInfo?: WirePageInfo
        }>('/auth/v1/iam/policies', {
            params: {
                ...toCollectionParams(query),
                ...(scope ? { scope } : {}),
                ...(orgId ? { orgId } : {}),
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

export async function getPolicy(locator: PolicyLocator): Promise<IamResult<PolicyInfo>> {
    try {
        const { data } = await http.get<{ policy?: PolicyInfo }>(`/auth/v1/iam/policies/${locator.name}`, {
            params: {
                scope: locator.scope,
                ...(locator.orgId ? { orgId: locator.orgId } : {}),
            },
        })
        if (!data.policy) {
            return { success: false, message: 'Policy not found' }
        }
        return { success: true, data: data.policy }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to load policy')
    }
}

export async function createPolicyVersion(payload: CreatePolicyVersionPayload): Promise<
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
    locator: PolicyLocator,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<PolicyVersionInfo>>> {
    try {
        const { data } = await http.get<{
            versions?: PolicyVersionInfo[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/policies/${locator.name}/versions`, {
            params: {
                ...toCollectionParams(query),
                scope: locator.scope,
                ...(locator.orgId ? { orgId: locator.orgId } : {}),
            },
        })
        return {
            success: true,
            data: {
                items: data.versions || [],
                pageInfo: normalizePageInfo(data.pageInfo, query),
            },
        }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to load policy versions')
    }
}

export async function setDefaultPolicyVersion(locator: PolicyLocator, version: string): Promise<IamResult<true>> {
    try {
        await http.post(`/auth/v1/iam/policies/${locator.name}/versions/${version}:set-default`, {
            name: locator.name,
            scope: locator.scope,
            orgId: locator.orgId,
            version,
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to set default policy version')
    }
}

export async function deletePolicyVersion(locator: PolicyLocator, version: string): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/policies/${locator.name}/versions/${version}`, {
            params: {
                scope: locator.scope,
                ...(locator.orgId ? { orgId: locator.orgId } : {}),
            },
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete policy version')
    }
}

export async function listPolicyAttachments(
    locator: PolicyLocator,
    query: CollectionQuery
): Promise<IamResult<CollectionPage<PolicyAttachmentInfo>>> {
    try {
        const { data } = await http.get<{
            attachments?: PolicyAttachmentInfo[]
            pageInfo?: WirePageInfo
        }>(`/auth/v1/iam/policies/${locator.name}/attachments`, {
            params: {
                ...toCollectionParams(query),
                scope: locator.scope,
                ...(locator.orgId ? { orgId: locator.orgId } : {}),
            },
        })
        return {
            success: true,
            data: {
                items: data.attachments || [],
                pageInfo: normalizePageInfo(data.pageInfo, query),
            },
        }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to load policy attachments')
    }
}

export async function deletePolicy(locator: PolicyLocator): Promise<IamResult<true>> {
    try {
        await http.delete(`/auth/v1/iam/policies/${locator.name}`, {
            params: {
                scope: locator.scope,
                ...(locator.orgId ? { orgId: locator.orgId } : {}),
            },
        })
        return { success: true, data: true }
    } catch (error) {
        return toFailure(error as HttpError, 'Failed to delete policy')
    }
}
