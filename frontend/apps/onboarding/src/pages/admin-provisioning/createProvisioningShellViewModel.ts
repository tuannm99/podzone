import { createEffect, createResource, createSignal } from 'solid-js'
import { ensureActiveTenant } from '@podzone/shared/services/auth'
import { listUserTenants, type TenantMembership } from '@podzone/shared/services/iam'

import { useAuthContext } from '@podzone/shared/auth'

export type ProvisioningTab = 'pipeline' | 'resources' | 'connections'

function resolvedUserID(id: number) {
    return Number.isSafeInteger(id) && id > 0 ? id : 0
}

export function createProvisioningShellViewModel() {
    const auth = useAuthContext()
    const [activeTab, setActiveTabSignal] = createSignal<ProvisioningTab>('pipeline')
    const [selectedTenantId, setSelectedTenantId] = createSignal('')
    const [memberships] = createResource(
        () => resolvedUserID(auth.getUserId()) || undefined,
        async (userId): Promise<TenantMembership[]> => {
            const result = await listUserTenants(userId)
            if (!result.success) throw new Error(result.message)
            return result.data.filter((membership) => membership.status === 'active')
        }
    )
    const [tenantSession] = createResource(
        () => selectedTenantId().trim() || undefined,
        async (tenantId) => {
            const result = await ensureActiveTenant(tenantId)
            if (!result.success) {
                throw new Error(result.data.message || 'Failed to activate workspace')
            }
            auth.setActiveTenantId(tenantId)
            return tenantId
        }
    )

    createEffect(() => {
        if (selectedTenantId()) return
        const preferred =
            memberships()?.find((membership) => membership.tenantId === auth.getActiveTenantId()) || memberships()?.[0]
        if (preferred) setSelectedTenantId(preferred.tenantId)
    })

    const setActiveTab = (tab: ProvisioningTab) => {
        setActiveTabSignal(tab)
    }
    const workspaceReady = () => Boolean(selectedTenantId()) && tenantSession.latest === selectedTenantId().trim()
    const error = () => {
        const source = memberships.error || tenantSession.error
        return source instanceof Error ? source.message : ''
    }

    return {
        activeTab,
        setActiveTab,
        selectedTenantId,
        setSelectedTenantId,
        memberships: () => memberships() || [],
        loading: () => memberships.loading || tenantSession.loading,
        workspaceReady,
        error,
    }
}

export type ProvisioningShellViewModel = ReturnType<typeof createProvisioningShellViewModel>
