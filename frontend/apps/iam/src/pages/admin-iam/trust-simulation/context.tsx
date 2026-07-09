import { createContext, useContext } from 'solid-js'
import type { Accessor, ParentProps, Setter } from 'solid-js'
import type { RolePermissionBoundary, SimulateAccessResult } from '@podzone/shared/services/iam'
import type { SelectOption } from '@podzone/shared/ui/components/common/Primitives'
import type { SearchSelectOption } from '@podzone/shared/ui/components/common/SearchSelectField'
import type { IamPermissionOption } from '../shared/IamPermissionMatrix'

export type ScopeOption = {
    name: string
    value: string
}

export type TenantOption = {
    name: string
    value: string
}

export type AdminIamTrustSimContextValue = {
    permissionOptions: Accessor<IamPermissionOption[]>
    boundaryPolicyOptions: Accessor<SelectOption[]>
    platformUserOptions: Accessor<SearchSelectOption[]>
    platformUsersLoading: Accessor<boolean>
    platformUsersError: Accessor<string>
    searchPlatformUsers: (search: string) => void
    tenantUserOptions: Accessor<SearchSelectOption[]>
    tenantUsersLoading: Accessor<boolean>
    tenantUsersError: Accessor<string>
    searchTenantUsers: (search: string) => void
    trustRoleName: Accessor<string>
    setTrustRoleName: Setter<string>
    loadTrustPolicy: () => Promise<void>
    handleSaveTrustPolicy: () => Promise<void>
    trustBoundaryPolicyName: Accessor<string>
    setTrustBoundaryPolicyName: Setter<string>
    handleSaveRoleBoundary: () => Promise<void>
    handleDeleteRoleBoundary: () => Promise<void>
    roleBoundary: Accessor<RolePermissionBoundary | null>
    trustJson: Accessor<string>
    setTrustJson: Setter<string>
    simScope: Accessor<string>
    setSimScope: Setter<string>
    policyScopeOptions: ScopeOption[]
    simTargetUserId: Accessor<string>
    setSimTargetUserId: Setter<string>
    tenantOptions: Accessor<TenantOption[]>
    simTenantId: Accessor<string>
    setSimTenantId: Setter<string>
    simAction: Accessor<string>
    setSimAction: Setter<string>
    simResource: Accessor<string>
    setSimResource: Setter<string>
    simServicePrincipal: Accessor<string>
    setSimServicePrincipal: Setter<string>
    applyServiceAssumePreset: () => void
    applyTenantAssumePreset: () => void
    applyScopeDownDenyPreset: () => void
    simAttributesJson: Accessor<string>
    setSimAttributesJson: Setter<string>
    simSessionTagsJson: Accessor<string>
    setSimSessionTagsJson: Setter<string>
    simSessionPolicyJson: Accessor<string>
    setSimSessionPolicyJson: Setter<string>
    simAssumedRoleId: Accessor<string>
    setSimAssumedRoleId: Setter<string>
    simAssumedRoleScope: Accessor<string>
    setSimAssumedRoleScope: Setter<string>
    simAssumedRoleName: Accessor<string>
    setSimAssumedRoleName: Setter<string>
    simAssumedRoleTenantId: Accessor<string>
    setSimAssumedRoleTenantId: Setter<string>
    simAssumedRoleSessionName: Accessor<string>
    setSimAssumedRoleSessionName: Setter<string>
    simAssumedRoleSourceIdentity: Accessor<string>
    setSimAssumedRoleSourceIdentity: Setter<string>
    simAssumedRoleServicePrincipal: Accessor<string>
    setSimAssumedRoleServicePrincipal: Setter<string>
    simAssumedRoleExpiresAt: Accessor<string>
    setSimAssumedRoleExpiresAt: Setter<string>
    handleSimulate: () => Promise<void>
    simulation: Accessor<SimulateAccessResult | undefined>
    simulationSourceColor: (source: string) => 'blue' | 'green' | 'yellow' | 'pink' | 'dark' | 'red' | 'indigo'
    statementSourceLabel: (source: string) => string
    simulationLayerTone: (allowed: boolean, reason: string) => string
}

const AdminIamTrustSimContext = createContext<AdminIamTrustSimContextValue>()

export function AdminIamTrustSimProvider(props: ParentProps<{ value: AdminIamTrustSimContextValue }>) {
    return <AdminIamTrustSimContext.Provider value={props.value}>{props.children}</AdminIamTrustSimContext.Provider>
}

export function useAdminIamTrustSim() {
    const ctx = useContext(AdminIamTrustSimContext)
    if (!ctx) {
        throw new Error('AdminIamTrustSimContext is missing')
    }
    return ctx
}
