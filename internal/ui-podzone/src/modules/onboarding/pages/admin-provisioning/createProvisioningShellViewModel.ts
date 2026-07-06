import { createEffect, createResource, createSignal } from 'solid-js'
import { ensureActiveTenant } from '@/services/auth'
import { listUserTenants, type TenantMembership } from '@/services/iam'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'

export type ProvisioningTab = 'pipeline' | 'resources' | 'connections'

function currentUserID() {
    const value = Number(tokenStorage.getUser()?.id)
    return Number.isSafeInteger(value) && value > 0 ? value : 0
}

export function createProvisioningShellViewModel() {
    const hashTab = window.location.hash.slice(1) as ProvisioningTab
    const validTabs = new Set<ProvisioningTab>(['pipeline', 'resources', 'connections'])
    const [activeTab, setActiveTabSignal] = createSignal<ProvisioningTab>(validTabs.has(hashTab) ? hashTab : 'pipeline')
    const [selectedTenantId, setSelectedTenantId] = createSignal('')
    const [memberships] = createResource(
        () => currentUserID() || undefined,
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
            tenantStorage.setTenantID(tenantId)
            return tenantId
        }
    )

    createEffect(() => {
        if (selectedTenantId()) return
        const preferred =
            memberships()?.find((membership) => membership.tenantId === tokenStorage.getActiveTenantID()) ||
            memberships()?.[0]
        if (preferred) setSelectedTenantId(preferred.tenantId)
    })

    const setActiveTab = (tab: ProvisioningTab) => {
        setActiveTabSignal(tab)
        window.history.replaceState(null, '', `${window.location.pathname}#${tab}`)
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
