import { createSignal } from 'solid-js'
import type { RolePermissionBoundary, SimulateAccessResult } from '@podzone/shared/services/iam'
import { prettyJSON } from '../presentation'

export function createTrustSimulationState(userID: number) {
    const [roleBoundary, setRoleBoundary] = createSignal<RolePermissionBoundary | null>(null)
    const [simulation, setSimulation] = createSignal<SimulateAccessResult>()
    const [trustRoleName, setTrustRoleName] = createSignal('tenant_admin')
    const [trustBoundaryPolicyName, setTrustBoundaryPolicyName] = createSignal('')
    const [trustJson, setTrustJson] = createSignal(
        prettyJSON([
            {
                effect: 'allow',
                principalType: 'service',
                principalPattern: 'backoffice.podzone.internal',
                tenantPattern: '*',
                externalIdPattern: '',
            },
        ])
    )
    const [simScope, setSimScope] = createSignal('tenant')
    const [simTenantId, setSimTenantId] = createSignal('')
    const [simTargetUserId, setSimTargetUserId] = createSignal(userID ? String(userID) : '')
    const [simAction, setSimAction] = createSignal('order:update')
    const [simResource, setSimResource] = createSignal('*')
    const [simServicePrincipal, setSimServicePrincipal] = createSignal('')
    const [simAttributesJson, setSimAttributesJson] = createSignal(prettyJSON({}))
    const [simSessionTagsJson, setSimSessionTagsJson] = createSignal(prettyJSON({ team: 'ops', lane: 'priority' }))
    const [simSessionPolicyJson, setSimSessionPolicyJson] = createSignal(
        prettyJSON([
            {
                effect: 'allow',
                actionPattern: 'order:update',
                resourcePattern: '*',
                conditions: [],
            },
        ])
    )
    const [simAssumedRoleId, setSimAssumedRoleId] = createSignal('')
    const [simAssumedRoleScope, setSimAssumedRoleScope] = createSignal('tenant')
    const [simAssumedRoleName, setSimAssumedRoleName] = createSignal('')
    const [simAssumedRoleTenantId, setSimAssumedRoleTenantId] = createSignal('')
    const [simAssumedRoleSessionName, setSimAssumedRoleSessionName] = createSignal('')
    const [simAssumedRoleSourceIdentity, setSimAssumedRoleSourceIdentity] = createSignal('')
    const [simAssumedRoleServicePrincipal, setSimAssumedRoleServicePrincipal] = createSignal('')
    const [simAssumedRoleExpiresAt, setSimAssumedRoleExpiresAt] = createSignal('')

    return {
        roleBoundary,
        setRoleBoundary,
        simulation,
        setSimulation,
        trustRoleName,
        setTrustRoleName,
        trustBoundaryPolicyName,
        setTrustBoundaryPolicyName,
        trustJson,
        setTrustJson,
        simScope,
        setSimScope,
        simTenantId,
        setSimTenantId,
        simTargetUserId,
        setSimTargetUserId,
        simAction,
        setSimAction,
        simResource,
        setSimResource,
        simServicePrincipal,
        setSimServicePrincipal,
        simAttributesJson,
        setSimAttributesJson,
        simSessionTagsJson,
        setSimSessionTagsJson,
        simSessionPolicyJson,
        setSimSessionPolicyJson,
        simAssumedRoleId,
        setSimAssumedRoleId,
        simAssumedRoleScope,
        setSimAssumedRoleScope,
        simAssumedRoleName,
        setSimAssumedRoleName,
        simAssumedRoleTenantId,
        setSimAssumedRoleTenantId,
        simAssumedRoleSessionName,
        setSimAssumedRoleSessionName,
        simAssumedRoleSourceIdentity,
        setSimAssumedRoleSourceIdentity,
        simAssumedRoleServicePrincipal,
        setSimAssumedRoleServicePrincipal,
        simAssumedRoleExpiresAt,
        setSimAssumedRoleExpiresAt,
    }
}
