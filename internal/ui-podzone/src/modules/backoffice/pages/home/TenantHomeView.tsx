import { useParams } from '@tanstack/solid-router'
import { TENANT_GQL_URL } from '@/services/baseurl'
import { PageShell } from '@/solid/components/common/PageShell'
import { EmptyBlock, ErrorAlert } from '@/solid/components/common/Feedback'
import { Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { StatCard } from '@/solid/components/dashboard/StatCard'
import { useTenantWorkspace } from '@/solid/workspace/context'
import { createHomeViewModel } from './createHomeViewModel'
import { TenantHomeSections } from './TenantHomeSections'

export function TenantHomeView() {
    const params = useParams({ from: '/t/$tenantId' })
    const workspace = useTenantWorkspace()
    const currentStoreId = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreId() || 'Select a store'
    const home = createHomeViewModel({
        tenantID: () => params().tenantId,
        workspaceReady,
    })

    return (
        <PageShell>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Store Workspace"
                    title={storeLabel()}
                    copy="This workspace stays scoped to the active tenant session and the selected store. Product setup, order routing, shipment control, and settlement readiness now resolve against backend store data."
                />
            </Card>

            {home.error() ? <ErrorAlert>{home.error()}</ErrorAlert> : null}
            {!workspaceReady() ? (
                <EmptyBlock
                    title="Choose a store first"
                    copy="Use the store switcher in the workspace shell before loading store-scoped POD operations."
                />
            ) : null}

            <div class="grid gap-4 md:grid-cols-3">
                <StatCard label="Store id" value={currentStoreId() || 'pending'} />
                <StatCard label="Transport" value="GraphQL" />
                <StatCard label="Endpoint" value={TENANT_GQL_URL.replace(/^https?:\/\//, '')} />
            </div>

            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard label="Setup drafts" value={String(home.draftCount())} />
                <StatCard label="Published mock products" value={String(home.publishedCandidateCount())} />
                <StatCard label="Orders in production" value={String(home.inProductionCount())} />
                <StatCard label="Open exceptions" value={String(home.openExceptionCount())} />
            </div>

            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard label="Mock revenue" value={home.mockRevenue()} />
                <StatCard label="Realized margin" value={home.realizedMarginTotal()} />
                <StatCard label="Pending settlements" value={String(home.pendingSettlementCount())} />
                <StatCard label="Issue cost" value={home.issueCostExposure()} />
                <StatCard label="Shipment SLA overdue" value={String(home.shipmentSlaOverdueCount())} />
                <StatCard label="Top partner load" value={home.topPartnerLoad()} />
            </div>

            <TenantHomeSections
                tenantId={params().tenantId}
                currentStoreId={currentStoreId}
                tenantReady={home.tenantReady}
                publishedCandidateCount={home.publishedCandidateCount}
                inProductionCount={home.inProductionCount}
                openExceptionCount={home.openExceptionCount}
                mockRevenue={home.mockRevenue}
                realizedMarginTotal={home.realizedMarginTotal}
                pendingSettlementCount={home.pendingSettlementCount}
                disputedSettlementCount={home.disputedSettlementCount}
                issueCostExposure={home.issueCostExposure}
                shipmentSlaOverdueCount={home.shipmentSlaOverdueCount}
                issueSlaOverdueCount={home.issueSlaOverdueCount}
                topPartnerLoad={home.topPartnerLoad}
                issueRate={home.issueRate}
            />
        </PageShell>
    )
}
