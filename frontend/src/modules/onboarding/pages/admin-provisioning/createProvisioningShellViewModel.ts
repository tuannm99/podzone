import { useNavigate, useSearch } from '@tanstack/solid-router'
import { createEffect, createResource, createSignal } from 'solid-js'
import { ensureActiveTenant } from '@/services/auth'
import { listUserTenants, type TenantMembership } from '@/services/iam'

import { useAuthContext } from '@/modules/shell/auth-context'

export type ProvisioningTab = 'pipeline' | 'resources' | 'connections'

function resolvedUserID(id: number) {
    return Number.isSafeInteger(id) && id > 0 ? id : 0
}

export function createProvisioningShellViewModel() {
    const auth = useAuthContext()
    const navigate = useNavigate()
    const search = useSearch({ from: '/admin/provisioning' })
    const validTabs = new Set<ProvisioningTab>(['pipeline', 'resources', 'connections'])
    const activeTab = (): ProvisioningTab => {
        const tab = search().tab
        return validTabs.has(tab as ProvisioningTab) ? (tab as ProvisioningTab) : 'pipeline'
    }
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
            memberships()?.find((membership) => membership.tenantId === auth.getActiveTenantId()) ||
            memberships()?.[0]
        if (preferred) setSelectedTenantId(preferred.tenantId)
    })

    const setActiveTab = (tab: ProvisioningTab) => {
        const current = search()
        void navigate({ to: '/admin/provisioning', search: { ...current, tab } } as unknown as Parameters<
            typeof navigate
        >[0])
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
