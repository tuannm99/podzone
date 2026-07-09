import { Show } from 'solid-js'
import { useAuthContext } from '@/modules/shell/auth-context'
import { ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { StatCard } from '@/solid/components/dashboard/StatCard'
import { useAdminHome } from './context'

export function HeaderStats() {
    const vm = useAdminHome()
    const auth = useAuthContext()

    return (
        <>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Seller Backoffice"
                    title="Choose the store you want to operate."
                    copy="Start from a tenant workspace, create or select a store, then enter the store-scoped backoffice."
                />
                <div class="flex flex-wrap gap-3">
                    <Button href="/admin/iam" color="dark" size="sm">
                        Open IAM console
                    </Button>
                </div>
            </Card>

            <Show when={vm.tenantError()}>
                <ErrorAlert>{vm.tenantError()}</ErrorAlert>
            </Show>

            <Show when={vm.tenantMessage()}>
                <InfoAlert>{vm.tenantMessage()}</InfoAlert>
            </Show>

            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard label="Session" value={auth.isAuthenticated() ? 'Active' : 'Missing'} />
                <StatCard label="User" value={vm.user?.username || vm.user?.email || 'Unknown'} />
                <StatCard label="My workspaces" value={`${vm.activeMemberships().length}/${vm.memberships().length}`} />
                <StatCard
                    label="Current store"
                    value={
                        auth.getActiveTenantId()
                            ? auth.getStoreId(auth.getActiveTenantId()) || 'Not selected'
                            : 'Not selected'
                    }
                />
                <StatCard
                    label="Stores with attention"
                    value={String(
                        vm
                            .storeAttention()
                            .filter(
                                (store: { overdueCount: number; disputedCount: number; unassignedCount: number }) =>
                                    store.overdueCount > 0 || store.disputedCount > 0 || store.unassignedCount > 0
                            ).length
                    )}
                />
            </div>
        </>
    )
}
