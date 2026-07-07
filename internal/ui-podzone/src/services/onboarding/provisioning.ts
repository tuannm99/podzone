import { ONBOARDING_API_URL } from '../baseurl'
import {
    normalizePageInfo,
    toCollectionParams,
    type CollectionPage,
    type CollectionQuery,
    type WirePageInfo,
} from '../collection'
import { http, type HttpError } from '../http'
import type { StoreRequestStatus } from '../onboarding'

export type ProvisioningResult<T> = { success: true; data: T } | { success: false; message: string }

export type StoreRequestTransition = {
    id: string
    request_id: string
    from: StoreRequestStatus | ''
    to: StoreRequestStatus
    actor: Record<string, string>
    step: string
    reason: string
    error_code: string
    created_at: string
}

export type DatabaseClusterResource = {
    name: string
    engine: string
    region: string
    placement_db: string
    max_tenants: number
    current_tenants: number
    max_schemas: number
    current_schemas: number
    max_connections: number
    current_connections: number
    status: string
    healthy: boolean
    created_at: string
    updated_at: string
}

export type KubernetesNamespaceResource = {
    name: string
    max_tenants: number
    current_tenants: number
    cpu_milli: number
    memory_mi: number
    status: string
    healthy: boolean
}

export type KubernetesClusterResource = {
    name: string
    region: string
    namespaces: KubernetesNamespaceResource[]
    status: string
    healthy: boolean
    created_at: string
    updated_at: string
}

export type RuntimePoolResource = {
    name: string
    kind: string
    max_tenants: number
    current_tenants: number
    status: string
    healthy: boolean
    created_at: string
    updated_at: string
}

export type DatabaseClusterHealthCheck = {
    name: string
    healthy: boolean
    current_tenants: number
    current_schemas: number
    current_connections: number
    message?: string
    checked_at: string
}

export type InfrastructureConnection = {
    tenant_id: string
    infra_type: string
    name: string
    endpoint: string
    secret_ref: string
    status: string
    version: number
    meta: Record<string, string>
    config: Record<string, unknown>
    created_at: string
    updated_at: string
    deleted_at?: string
}

export type UpsertInfrastructureConnection = {
    infra_type: string
    name: string
    endpoint: string
    secret_ref: string
    status: string
    cluster_name: string
    mode: string
    db_name: string
    schema_name: string
    meta: Record<string, string>
    config: Record<string, unknown>
}

export type PlacementRoute = {
    cluster_name: string
    mode: string
    db_name: string
    schema_name: string
}

export type PlacementStatus = {
    tenant_id: string
    allocation_id?: string
    allocation_ready: boolean
    route_ready: boolean
    in_sync: boolean
    needs_repair: boolean
    reason?: string
    allocation?: PlacementRoute
    route?: PlacementRoute
    updated_at?: string
}

export type PlacementReconcileResponse = {
    status: PlacementStatus
    repaired: boolean
    kv_store_key: string
    published_at?: string
}

type ResourceKind = 'database-clusters' | 'kubernetes-clusters' | 'runtime-pools'

function errorMessage(error: unknown) {
    const requestError = error as HttpError
    return requestError?.message || 'Provisioning request failed'
}

function tenantHeaders(tenantId: string) {
    return { 'X-Tenant-ID': tenantId.trim() }
}

async function listPage<T>(
    path: string,
    query: CollectionQuery,
    headers?: Record<string, string>
): Promise<ProvisioningResult<CollectionPage<T>>> {
    try {
        const response = await http.get<{
            items?: T[]
            pageInfo?: WirePageInfo
        }>(`${ONBOARDING_API_URL}${path}`, {
            headers,
            params: toCollectionParams(query),
        })
        return {
            success: true,
            data: {
                items: response.data.items || [],
                pageInfo: normalizePageInfo(response.data.pageInfo, query),
            },
        }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export function listStoreRequestTransitions(tenantId: string, requestId: string, query: CollectionQuery) {
    return listPage<StoreRequestTransition>(
        `/onboarding/v1/requests/${encodeURIComponent(requestId)}/transitions`,
        query,
        tenantHeaders(tenantId)
    )
}

export function listDatabaseClusters(query: CollectionQuery) {
    return listPage<DatabaseClusterResource>('/onboarding/v1/infras/resources/database-clusters', query)
}

export function listKubernetesClusters(query: CollectionQuery) {
    return listPage<KubernetesClusterResource>('/onboarding/v1/infras/resources/kubernetes-clusters', query)
}

export function listRuntimePools(query: CollectionQuery) {
    return listPage<RuntimePoolResource>('/onboarding/v1/infras/resources/runtime-pools', query)
}

async function upsertResource<T>(kind: ResourceKind, name: string, resource: T): Promise<ProvisioningResult<void>> {
    try {
        await http.put(
            `${ONBOARDING_API_URL}/onboarding/v1/infras/resources/${kind}/${encodeURIComponent(name)}`,
            resource
        )
        return { success: true, data: undefined }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export function upsertDatabaseCluster(resource: DatabaseClusterResource) {
    return upsertResource('database-clusters', resource.name, resource)
}

export async function checkDatabaseClusterHealth(
    name: string
): Promise<ProvisioningResult<DatabaseClusterHealthCheck>> {
    try {
        const response = await http.post<DatabaseClusterHealthCheck>(
            `${ONBOARDING_API_URL}/onboarding/v1/infras/resources/database-clusters/${encodeURIComponent(name)}/health-check`
        )
        return { success: true, data: response.data }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export function upsertKubernetesCluster(resource: KubernetesClusterResource) {
    return upsertResource('kubernetes-clusters', resource.name, resource)
}

export function upsertRuntimePool(resource: RuntimePoolResource) {
    return upsertResource('runtime-pools', resource.name, resource)
}

export async function deleteResource(kind: ResourceKind, name: string): Promise<ProvisioningResult<void>> {
    try {
        await http.delete(`${ONBOARDING_API_URL}/onboarding/v1/infras/resources/${kind}/${encodeURIComponent(name)}`)
        return { success: true, data: undefined }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export function listInfrastructureConnections(tenantId: string, query: CollectionQuery) {
    return listPage<InfrastructureConnection>('/onboarding/v1/infras/connections', query, tenantHeaders(tenantId))
}

export async function upsertInfrastructureConnection(
    tenantId: string,
    connection: UpsertInfrastructureConnection
): Promise<ProvisioningResult<void>> {
    try {
        await http.post(`${ONBOARDING_API_URL}/onboarding/v1/infras/connections`, connection, {
            headers: tenantHeaders(tenantId),
        })
        return { success: true, data: undefined }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export async function deleteInfrastructureConnection(
    tenantId: string,
    infraType: string,
    name: string
): Promise<ProvisioningResult<void>> {
    try {
        await http.delete(
            `${ONBOARDING_API_URL}/onboarding/v1/infras/connections/${encodeURIComponent(infraType)}/${encodeURIComponent(name)}`,
            { headers: tenantHeaders(tenantId) }
        )
        return { success: true, data: undefined }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export async function getPlacementStatus(tenantId: string): Promise<ProvisioningResult<PlacementStatus>> {
    try {
        const response = await http.get<PlacementStatus>(
            `${ONBOARDING_API_URL}/onboarding/v1/infras/placements/${encodeURIComponent(tenantId)}/status`
        )
        return { success: true, data: response.data }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}

export async function reconcilePlacement(tenantId: string): Promise<ProvisioningResult<PlacementReconcileResponse>> {
    try {
        const response = await http.post<PlacementReconcileResponse>(
            `${ONBOARDING_API_URL}/onboarding/v1/infras/placements/${encodeURIComponent(tenantId)}/reconcile`
        )
        return { success: true, data: response.data }
    } catch (error) {
        return { success: false, message: errorMessage(error) }
    }
}
