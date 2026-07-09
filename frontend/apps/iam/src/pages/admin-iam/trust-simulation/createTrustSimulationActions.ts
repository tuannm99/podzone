import { createEffect } from 'solid-js'
import {
    deleteRolePermissionBoundary,
    putRolePermissionBoundary,
    putRoleTrustPolicy,
    simulateAccess,
    type PolicyStatement,
    type RoleTrustStatement,
} from '@podzone/shared/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import { parseJSONArray, parseJSONObject } from '../presentation'
import type { RunAction } from '../shared/actions'
import { createSimulationPresets } from './simulation-presets'

export function createTrustSimulationActions(state: AdminIamState, loaders: AdminIamLoaders, runAction: RunAction) {
    createEffect(() => {
        if (
            state.simAssumedRoleScope() === 'tenant' &&
            state.simTenantId().trim() &&
            !state.simAssumedRoleTenantId().trim()
        )
            state.setSimAssumedRoleTenantId(state.simTenantId().trim())
    })

    const buildAssumedRoleSession = () => {
        const roleID = Number.parseInt(state.simAssumedRoleId().trim(), 10)
        if (!Number.isFinite(roleID) || roleID <= 0) return undefined
        return {
            assumedRoleId: roleID,
            assumedRoleScope: state.simAssumedRoleScope().trim(),
            assumedRoleName: state.simAssumedRoleName().trim(),
            assumedRoleTenantId: state.simAssumedRoleTenantId().trim() || undefined,
            assumedRoleSessionName: state.simAssumedRoleSessionName().trim() || undefined,
            assumedRoleSourceIdentity: state.simAssumedRoleSourceIdentity().trim() || undefined,
            assumedRoleServicePrincipal: state.simAssumedRoleServicePrincipal().trim() || undefined,
            assumedRoleExpiresAt: state.simAssumedRoleExpiresAt().trim() || undefined,
            sessionTags: parseJSONObject(state.simSessionTagsJson(), 'Session tags'),
        }
    }

    const handleSaveTrustPolicy = () =>
        runAction(async () => {
            const roleName = state.trustRoleName().trim()
            const statements = parseJSONArray<RoleTrustStatement>(state.trustJson(), 'Trust policy')
            const result = await putRoleTrustPolicy({ roleName, statements })
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Saved trust policy for role ${roleName}.`)
            await loaders.loadTrustPolicy()
        })

    const handleSaveRoleBoundary = () =>
        runAction(async () => {
            const roleName = state.trustRoleName().trim()
            const result = await putRolePermissionBoundary(roleName, state.trustBoundaryPolicyName().trim())
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Saved role boundary for ${roleName}.`)
            await loaders.loadTrustPolicy()
        })

    const handleDeleteRoleBoundary = () =>
        runAction(async () => {
            const roleName = state.trustRoleName().trim()
            const result = await deleteRolePermissionBoundary(roleName)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Deleted role boundary for ${roleName}.`)
            state.setRoleBoundary(null)
            state.setTrustBoundaryPolicyName('')
        })

    const handleSimulate = () =>
        runAction(async () => {
            const assumedRoleSession = buildAssumedRoleSession()
            const result = await simulateAccess({
                scope: state.simScope(),
                tenantId: state.simTenantId().trim() || undefined,
                userId: Number.parseInt(state.simTargetUserId().trim(), 10),
                action: state.simAction().trim(),
                resource: state.simResource().trim(),
                useAssumedRole: Boolean(assumedRoleSession),
                assumedRoleSession,
                sessionPolicy: parseJSONArray<PolicyStatement>(state.simSessionPolicyJson(), 'Session policy'),
                attributes: parseJSONObject(state.simAttributesJson(), 'Simulation attributes'),
                servicePrincipal: state.simServicePrincipal().trim() || undefined,
                sessionTags: parseJSONObject(state.simSessionTagsJson(), 'Session tags'),
            })
            if (!result.success) throw new Error(result.message)
            state.setSimulation(result.data)
            state.setPageMessage(
                `Simulation completed: ${result.data.allowed ? 'allowed' : 'denied'} via ${result.data.decisionSource}.`
            )
        })

    return {
        ...createSimulationPresets(state),
        handleSaveTrustPolicy,
        handleSaveRoleBoundary,
        handleDeleteRoleBoundary,
        handleSimulate,
    }
}
