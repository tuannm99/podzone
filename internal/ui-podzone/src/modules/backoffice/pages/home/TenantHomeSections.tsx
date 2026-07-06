import { tokenStorage } from '@/services/tokenStorage'
import { EmptyBlock } from '@/solid/components/common/Feedback'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { buildOrdersHref, buildTenantHref } from './presentation'

type TenantHomeSectionsProps = {
    tenantId: string
    currentStoreId: () => string
    tenantReady: () => boolean
    publishedCandidateCount: () => number
    inProductionCount: () => number
    openExceptionCount: () => number
    mockRevenue: () => string
    realizedMarginTotal: () => string
    pendingSettlementCount: () => number
    disputedSettlementCount: () => number
    issueCostExposure: () => string
    shipmentSlaOverdueCount: () => number
    issueSlaOverdueCount: () => number
    topPartnerLoad: () => string
    issueRate: () => string
}

export function TenantHomeSections(props: TenantHomeSectionsProps) {
    const tenantHref = (path: string) => buildTenantHref(props.tenantId, props.currentStoreId(), path)
    const ordersHref = (queueView: string, queueSort = 'priority') =>
        buildOrdersHref(props.tenantId, props.currentStoreId(), queueView, queueSort)

    return (
        <>
            <Card class="space-y-4">
                <SectionTitle
                    title="Current store context"
                    subtitle="Requests in this workspace rely on the active store in the signed-in session. The local route value is only used for navigation."
                />
                <div class="flex flex-wrap gap-2">
                    <Badge
                        content={`tenant ${tokenStorage.getActiveTenantID() || 'missing'}`}
                        color={props.tenantReady() ? 'green' : 'yellow'}
                    />
                    <Badge content={`selected store: ${props.currentStoreId() || 'missing'}`} color="indigo" />
                    <Badge content="Authorization: Bearer ..." color="green" />
                </div>
                {!props.tenantReady() ? (
                    <EmptyBlock
                        title="Store session not ready"
                        copy="The client could not confirm this store as the current active workspace yet."
                    />
                ) : null}
            </Card>

            <Card class="space-y-4">
                <SectionTitle
                    title="POD operations pulse"
                    subtitle="Product setup, order routing, shipment control, and settlement metrics now come from backend store data."
                />
                <div class="flex flex-wrap gap-2">
                    <Badge
                        content={`${props.publishedCandidateCount()} published mock products`}
                        color={props.publishedCandidateCount() > 0 ? 'green' : 'dark'}
                    />
                    <Badge
                        content={`${props.inProductionCount()} orders in production`}
                        color={props.inProductionCount() > 0 ? 'blue' : 'dark'}
                    />
                    <Badge
                        content={`${props.openExceptionCount()} active issues`}
                        color={props.openExceptionCount() > 0 ? 'yellow' : 'green'}
                    />
                    <Badge content={`revenue ${props.mockRevenue()}`} color="indigo" />
                    <Badge content={`margin ${props.realizedMarginTotal()}`} color="blue" />
                    <Badge content={`issue cost ${props.issueCostExposure()}`} color="red" />
                    <Badge
                        content={`${props.shipmentSlaOverdueCount()} shipment SLA overdue`}
                        color={props.shipmentSlaOverdueCount() > 0 ? 'red' : 'green'}
                    />
                    <Badge
                        content={`${props.issueSlaOverdueCount()} issue SLA overdue`}
                        color={props.issueSlaOverdueCount() > 0 ? 'red' : 'green'}
                    />
                    <Badge
                        content={`${props.pendingSettlementCount()} pending settlements`}
                        color={props.pendingSettlementCount() > 0 ? 'yellow' : 'green'}
                    />
                    <Badge
                        content={`issue rate ${props.issueRate()}`}
                        color={props.openExceptionCount() > 0 ? 'yellow' : 'green'}
                    />
                </div>
                {!props.publishedCandidateCount() ? (
                    <EmptyBlock
                        title="No published products yet"
                        copy="Start in Product setup, promote a candidate, and mock publish it before testing the rest of this POD workflow."
                    />
                ) : null}
            </Card>

            <Card class="space-y-4">
                <SectionTitle
                    title="Start here"
                    subtitle="A simple guided POD flow that separates the partner record layer from store-side catalog and routing operations."
                />
                <div class="grid gap-4 md:grid-cols-4">
                    <StartCard
                        label="1. Partner setup"
                        title="Confirm who prints and fulfills"
                        copy="Keep partner records current before shaping products or routing test orders."
                        href={tenantHref('/partners')}
                        color="blue"
                        action="Review print partners"
                    />
                    <StartCard
                        label="2. Product setup"
                        title="Build candidates for the store catalog"
                        copy="Create backend-backed drafts, verify artwork readiness, and mock publish what is ready to route."
                        href={tenantHref('/products/setup')}
                        color="green"
                        action="Open product setup"
                    />
                    <StartCard
                        label="3. Order operations"
                        title="Route, ship, and settle"
                        copy="Route backend-backed store orders through production flow, manual shipment status, and settlement readiness."
                        href={tenantHref('/orders')}
                        color="alternative"
                        action="Open orders board"
                    />
                </div>
            </Card>

            <CommercialSnapshot {...props} />
            <ActionShortcuts {...props} tenantHref={tenantHref} />

            <Card class="space-y-4">
                <SectionTitle
                    title="Queue Shortcuts"
                    subtitle="Open the orders board already focused on the slice that needs attention."
                />
                <div class="flex flex-wrap gap-3">
                    <Button
                        href={ordersHref('overdue')}
                        color={
                            props.shipmentSlaOverdueCount() > 0 || props.issueSlaOverdueCount() > 0
                                ? 'red'
                                : 'alternative'
                        }
                    >
                        Overdue queue
                    </Button>
                    <Button
                        href={ordersHref('delivery_issues')}
                        color={props.openExceptionCount() > 0 ? 'dark' : 'alternative'}
                    >
                        Issue queue
                    </Button>
                    <Button
                        href={ordersHref('settlement_pending')}
                        color={props.pendingSettlementCount() > 0 ? 'green' : 'alternative'}
                    >
                        Settlement follow-up
                    </Button>
                    <Button
                        href={tenantHref('/orders/finance')}
                        color={props.disputedSettlementCount() > 0 ? 'red' : 'alternative'}
                    >
                        Finance review
                    </Button>
                    <Button href={ordersHref('all', 'priority')} color="blue">
                        Priority queue
                    </Button>
                </div>
            </Card>

            <Card class="space-y-4">
                <SectionTitle title="Operations" subtitle="Direct links into each experimental POD workflow area." />
                <div class="flex flex-wrap gap-3">
                    <Button href={tenantHref('/products/setup')} color="green">
                        Open product setup
                    </Button>
                    <Button href={tenantHref('/partners')} color="blue">
                        Open print partners
                    </Button>
                    <Button href={tenantHref('/orders')} color="alternative">
                        Open orders
                    </Button>
                </div>
            </Card>
        </>
    )
}

type StartCardProps = {
    label: string
    title: string
    copy: string
    href: string
    color: 'blue' | 'green' | 'alternative'
    action: string
}

function StartCard(props: StartCardProps) {
    return (
        <div class="rounded-lg border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">{props.label}</p>
            <p class="mt-2 text-base font-semibold text-gray-900">{props.title}</p>
            <p class="mt-1 text-sm text-gray-600">{props.copy}</p>
            <div class="mt-4">
                <Button href={props.href} color={props.color}>
                    {props.action}
                </Button>
            </div>
        </div>
    )
}

function CommercialSnapshot(props: TenantHomeSectionsProps) {
    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Commercial snapshot"
                subtitle="A lightweight operational finance view from backend-backed routed orders before any separate analytics stack exists."
            />
            <div class="grid gap-4 md:grid-cols-3">
                <SnapshotCard
                    label="Realized margin"
                    value={props.realizedMarginTotal()}
                    copy="Calculated from routed order revenue minus fulfillment and shipping costs captured in the store workflow."
                />
                <SnapshotCard
                    label="Settlement pressure"
                    value={`${props.pendingSettlementCount()} pending · ${props.disputedSettlementCount()} disputed`}
                    copy="Highlights which routed orders still need reconciliation or manual finance follow-up."
                />
                <SnapshotCard
                    label="Issue cost exposure"
                    value={props.issueCostExposure()}
                    copy="Captures reprint, delivery issue, and other exception costs now reducing realized margin."
                />
                <SnapshotCard
                    label="Queue pressure"
                    value={`${props.shipmentSlaOverdueCount()} shipment · ${props.issueSlaOverdueCount()} issue`}
                    copy="Tracks overdue shipment and issue follow-up deadlines on the operator queue."
                />
            </div>
        </Card>
    )
}

function SnapshotCard(props: { label: string; value: string; copy: string }) {
    return (
        <div class="rounded-lg border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">{props.label}</p>
            <p class="mt-2 text-lg font-semibold text-gray-900">{props.value}</p>
            <p class="mt-1 text-sm text-gray-600">{props.copy}</p>
        </div>
    )
}

function ActionShortcuts(props: TenantHomeSectionsProps & { tenantHref: (path: string) => string }) {
    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Action shortcuts"
                subtitle="Jump straight to the next likely action based on the current operational state."
            />
            <div class="flex flex-wrap gap-3">
                <Button href={props.tenantHref('/products/setup')} color="green">
                    {props.publishedCandidateCount() > 0 ? 'Refine product candidates' : 'Publish first mock product'}
                </Button>
                <Button href={props.tenantHref('/orders')} color="blue">
                    {props.openExceptionCount() > 0 ? 'Review active issues' : 'Review routing board'}
                </Button>
                <Button href={props.tenantHref('/orders/finance')} color="alternative">
                    {props.pendingSettlementCount() > 0 ? 'Review settlement queue' : 'Create first routed order'}
                </Button>
                <Button href={props.tenantHref('/partners')} color="light">
                    {props.topPartnerLoad() === 'No partner load yet' ? 'Set up print partners' : 'Review partner load'}
                </Button>
            </div>
        </Card>
    )
}
