import { useParams } from '@tanstack/solid-router'
import { Show, createEffect, createResource } from 'solid-js'
import { getPartner, type PartnerInfo } from '@/services/partner'
import { tenantStorage } from '@/services/tenantStorage'
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'

function badgeColorForStatus(status: string) {
    return status === 'active' ? 'green' : 'dark'
}

function partnerTypeLabel(partnerType: string) {
    return partnerType.replaceAll('_', ' ')
}

function formatTimestamp(value?: string) {
    if (!value) return 'Not available'
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function formatCapabilityList(items?: string[]) {
    return items && items.length > 0 ? items.join(', ') : 'Any'
}

function formatShippingCostRules(items?: { region: string; cost: string }[]) {
    return items && items.length > 0 ? items.map((item) => `${item.region}:${item.cost}`).join(', ') : 'No region rules'
}

function DetailRow(props: { label: string; value: string }) {
    return (
        <div class="rounded-lg border border-gray-200 p-4">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">{props.label}</p>
            <p class="mt-2 text-sm text-gray-900">{props.value}</p>
        </div>
    )
}

export function TenantPartnerDetailView() {
    const params = useParams({ from: '/t/$tenantId/partners/$partnerId' })

    const [partnerResource, { refetch: reloadPartner }] = createResource(
        () => params().partnerId,
        async (partnerID): Promise<{ partnerID: string; partner: PartnerInfo }> => {
            const result = await getPartner(partnerID)
            if (!result.success) {
                throw new Error(result.message)
            }
            return { partnerID, partner: result.data }
        }
    )
    const partner = () =>
        partnerResource()?.partnerID === params().partnerId ? partnerResource()?.partner || null : null
    const loading = () => partnerResource.loading
    const error = () => (partnerResource.error instanceof Error ? partnerResource.error.message : '')

    createEffect(() => {
        tenantStorage.setTenantID(params().tenantId)
    })

    return (
        <PageShell>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Partner Record"
                    title={`Partner ${params().partnerId}`}
                    copy="Inspect one backend-backed partner record for this store, including identity, contact, type, and activation status."
                />
            </Card>

            <div class="flex flex-wrap gap-3">
                <Button color="light" href={`/t/${params().tenantId}/partners`}>
                    Back to partners
                </Button>
                <Button
                    color="alternative"
                    onClick={() => {
                        void reloadPartner()
                    }}
                >
                    Reload record
                </Button>
            </div>

            <Show when={error()}>
                <ErrorAlert>{error()}</ErrorAlert>
            </Show>

            <InfoAlert>
                This page reads a real partner record from the partner service. It is separate from the browser-local
                prototype flows used by product and order demo screens.
            </InfoAlert>

            <Show when={loading()}>
                <LoadingInline label="Loading partner record..." />
            </Show>

            <Show
                when={!loading() && partner()}
                fallback={
                    !loading() ? (
                        <EmptyBlock
                            title="Partner record unavailable"
                            copy="The requested partner could not be loaded for this store."
                        />
                    ) : null
                }
            >
                {(current) => (
                    <>
                        <Card class="space-y-4">
                            <SectionTitle
                                title={current().name}
                                subtitle="Store-scoped partner identity and current activation state."
                            />
                            <div class="flex flex-wrap gap-2">
                                <Badge content={current().status} color={badgeColorForStatus(current().status)} />
                                <Badge content={partnerTypeLabel(current().partnerType)} color="indigo" />
                                <Badge content={`tenant ${current().tenantId}`} color="blue" />
                            </div>
                        </Card>

                        <div class="grid gap-4 md:grid-cols-2">
                            <DetailRow label="Partner id" value={current().id} />
                            <DetailRow label="Partner code" value={current().code || 'Not set'} />
                            <DetailRow label="Contact name" value={current().contactName || 'Not set'} />
                            <DetailRow label="Contact email" value={current().contactEmail || 'Not set'} />
                            <DetailRow label="Created at" value={formatTimestamp(current().createdAt)} />
                            <DetailRow label="Updated at" value={formatTimestamp(current().updatedAt)} />
                            <DetailRow
                                label="Supported product types"
                                value={formatCapabilityList(current().supportedProductTypes)}
                            />
                            <DetailRow
                                label="Supported regions"
                                value={formatCapabilityList(current().supportedRegions)}
                            />
                            <DetailRow label="SLA days" value={String(current().slaDays || 0)} />
                            <DetailRow label="Routing priority" value={String(current().routingPriority || 0)} />
                            <DetailRow label="Base fulfillment cost" value={current().baseFulfillmentCost || 'TBD'} />
                            <DetailRow
                                label="Shipping cost rules"
                                value={formatShippingCostRules(current().shippingCostRules)}
                            />
                        </div>

                        <Card class="space-y-4">
                            <SectionTitle
                                title="Operational notes"
                                subtitle="Business notes stored on the partner record."
                            />
                            <Show
                                when={current().notes}
                                fallback={
                                    <EmptyBlock
                                        title="No notes yet"
                                        copy="This partner record does not currently include any stored operating notes."
                                    />
                                }
                            >
                                <p class="text-sm leading-6 text-gray-700">{current().notes}</p>
                            </Show>
                        </Card>
                    </>
                )}
            </Show>
        </PageShell>
    )
}
