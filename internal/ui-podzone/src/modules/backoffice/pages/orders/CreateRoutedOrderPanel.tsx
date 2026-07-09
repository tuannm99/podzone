import { For, Show } from 'solid-js'
import { EmptyBlock, ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { FormInputField, FormSelectField } from '@/solid/forms'
import { useTenantOrdersComposer } from './composer-context'

const productTypeOptions = [
    { name: 'T-shirt', value: 'tshirt' },
    { name: 'Hoodie', value: 'hoodie' },
    { name: 'Tote', value: 'tote' },
    { name: 'Poster', value: 'poster' },
]

const shipRegionOptions = [
    { name: 'US', value: 'us' },
    { name: 'EU', value: 'eu' },
    { name: 'UK', value: 'uk' },
    { name: 'SEA', value: 'sea' },
]

const exceptionOptions = [
    { name: 'Artwork issue', value: 'artwork_issue' },
    { name: 'Partner delay', value: 'partner_delay' },
    { name: 'Address hold', value: 'address_hold' },
    { name: 'Reprint request', value: 'reprint_request' },
]

function joinPartnerCapabilityList(items: string[]) {
    return items.length > 0 ? items.join(', ') : 'Any'
}

function joinShippingCostRules(items: { region: string; cost: string }[]) {
    return items.length > 0 ? items.map((item) => `${item.region}:${item.cost}`).join(', ') : 'No region rules'
}

export function CreateRoutedOrderPanel() {
    const composer = useTenantOrdersComposer()

    return (
        <>
            <SectionTitle
                title="Create routed order"
                subtitle="Use a published mock product candidate as the source, then send the order into the backend-backed POD routing flow."
            />

            <Show
                when={composer.availableCandidates().length > 0}
                fallback={
                    <EmptyBlock
                        title="No published mock products yet"
                        copy="Go to Product setup, promote a draft, and mock publish it from the backend-backed setup workflow before testing order routing."
                    />
                }
            >
                <form class="space-y-4" onSubmit={composer.createMockOrder}>
                    <FormSelectField
                        form={composer.form}
                        name="selectedCandidateId"
                        label="Published mock product"
                        options={composer.availableCandidates().map((candidate) => ({
                            name: `${candidate.title} · ${candidate.partner}`,
                            value: candidate.id,
                        }))}
                    />
                    <div class="grid gap-4 md:grid-cols-2">
                        <FormInputField
                            form={composer.form}
                            name="customerName"
                            label="Customer name"
                            placeholder="Nguyen Minh"
                        />
                        <FormInputField form={composer.form} name="quantity" label="Quantity" placeholder="1" />
                    </div>
                    <div class="grid gap-4 md:grid-cols-3">
                        <FormSelectField
                            form={composer.form}
                            name="selectedProductType"
                            label="Product type"
                            options={productTypeOptions}
                        />
                        <FormSelectField
                            form={composer.form}
                            name="selectedShipRegion"
                            label="Ship region"
                            options={shipRegionOptions}
                        />
                        <Show
                            when={composer.form.values.manualPartnerOverride}
                            fallback={
                                <div class="space-y-2 rounded-lg border border-dashed border-gray-300 bg-gray-50 p-3">
                                    <p class="text-sm font-medium text-gray-700">Partner routing mode</p>
                                    <p class="text-xs text-gray-500">
                                        Auto-route is active. The backend will pick the best eligible partner from
                                        capability, priority, and SLA.
                                    </p>
                                    <Button
                                        type="button"
                                        size="xs"
                                        color="alternative"
                                        onClick={() => composer.form.setValue('manualPartnerOverride', true)}
                                    >
                                        Override partner
                                    </Button>
                                </div>
                            }
                        >
                            <div class="space-y-2">
                                <FormInputField
                                    form={composer.form}
                                    name="preferredPartner"
                                    label="Preferred partner override"
                                    placeholder="optional code or name"
                                />
                                <Button
                                    type="button"
                                    size="xs"
                                    color="alternative"
                                    onClick={composer.resetPreferredPartnerOverride}
                                >
                                    Return to auto-route
                                </Button>
                            </div>
                        </Show>
                    </div>
                    <Show when={composer.routingRecommendation()}>
                        {(recommendation) => (
                            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                                <div class="flex flex-wrap items-center gap-2">
                                    <Badge content="routing recommendation" color="blue" />
                                    <Badge
                                        content={
                                            composer.form.values.manualPartnerOverride
                                                ? 'manual override'
                                                : 'auto-route active'
                                        }
                                        color={composer.form.values.manualPartnerOverride ? 'yellow' : 'green'}
                                    />
                                    <Show when={recommendation().candidatePartner}>
                                        <Badge
                                            content={`candidate default ${recommendation().candidatePartner}`}
                                            color="dark"
                                        />
                                    </Show>
                                    <Show when={recommendation().selectedPartner}>
                                        <Badge content={`selected ${recommendation().selectedPartner}`} color="green" />
                                    </Show>
                                </div>
                                <p class="mt-2 text-sm text-gray-600">{recommendation().summary}</p>
                                <Show when={!recommendation().selectedPartner && recommendation().blockedReason}>
                                    <ErrorAlert>
                                        Routing blocked: {recommendation().blockedReason}
                                        <Show when={recommendation().blockedReasonCode}>
                                            {' '}
                                            ({recommendation().blockedReasonCode})
                                        </Show>
                                    </ErrorAlert>
                                </Show>
                                <Show
                                    when={
                                        !composer.form.values.manualPartnerOverride && recommendation().selectedPartner
                                    }
                                >
                                    <InfoAlert>
                                        Auto-route will create the order against {recommendation().selectedPartner}.
                                        Switch to override only when you need to force a different eligible partner.
                                    </InfoAlert>
                                </Show>
                                <div class="mt-3 space-y-3">
                                    <Show
                                        when={recommendation().options.filter((option) => option.eligible).length > 0}
                                    >
                                        <div class="space-y-2">
                                            <p class="text-sm font-medium text-gray-700">Eligible partners</p>
                                            <For
                                                each={recommendation()
                                                    .options.filter((option) => option.eligible)
                                                    .slice(0, 4)}
                                            >
                                                {(option) => (
                                                    <div class="rounded-md bg-white p-3 text-sm text-gray-600">
                                                        <div class="flex flex-wrap items-center gap-2">
                                                            <Badge content={option.partner.name} color="green" />
                                                            <Badge
                                                                content={`priority ${option.partner.routingPriority}`}
                                                                color="blue"
                                                            />
                                                            <Badge
                                                                content={`${option.partner.slaDays}d sla`}
                                                                color="indigo"
                                                            />
                                                            <Show
                                                                when={
                                                                    recommendation().selectedPartner ===
                                                                    option.partner.name
                                                                }
                                                            >
                                                                <Badge content="recommended" color="green" />
                                                            </Show>
                                                        </div>
                                                        <p class="mt-2">{option.reason}</p>
                                                        <p class="mt-1 text-xs text-gray-500">
                                                            Products:{' '}
                                                            {joinPartnerCapabilityList(
                                                                option.partner.supportedProductTypes
                                                            )}{' '}
                                                            · Regions:{' '}
                                                            {joinPartnerCapabilityList(option.partner.supportedRegions)}
                                                        </p>
                                                        <p class="mt-1 text-xs text-gray-500">
                                                            Partner base fulfillment:{' '}
                                                            {option.partner.baseFulfillmentCost || 'TBD'} · Region cost
                                                            rules:{' '}
                                                            {joinShippingCostRules(option.partner.shippingCostRules)}
                                                        </p>
                                                        <div class="mt-2 flex flex-wrap gap-2">
                                                            <Badge
                                                                content={`fulfillment ${option.estimatedFulfillmentCost}`}
                                                                color="blue"
                                                            />
                                                            <Badge
                                                                content={`shipping ${option.estimatedShippingCost}`}
                                                                color="indigo"
                                                            />
                                                            <Badge
                                                                content={`unit margin ${option.estimatedUnitMargin}`}
                                                                color="green"
                                                            />
                                                        </div>
                                                        <div class="mt-3 flex flex-wrap gap-2">
                                                            <Show
                                                                when={
                                                                    recommendation().selectedPartner ===
                                                                    option.partner.name
                                                                }
                                                                fallback={
                                                                    <Button
                                                                        type="button"
                                                                        size="xs"
                                                                        color="alternative"
                                                                        onClick={() =>
                                                                            composer.applyPreferredPartnerOverride(
                                                                                option.partner.name
                                                                            )
                                                                        }
                                                                    >
                                                                        Force this partner
                                                                    </Button>
                                                                }
                                                            >
                                                                <Button
                                                                    type="button"
                                                                    size="xs"
                                                                    color="green"
                                                                    onClick={composer.resetPreferredPartnerOverride}
                                                                >
                                                                    Use auto-route
                                                                </Button>
                                                            </Show>
                                                        </div>
                                                    </div>
                                                )}
                                            </For>
                                        </div>
                                    </Show>
                                    <Show when={recommendation().options.some((option) => !option.eligible)}>
                                        <div class="space-y-2">
                                            <p class="text-sm font-medium text-gray-700">Blocked by capability</p>
                                            <For
                                                each={recommendation()
                                                    .options.filter((option) => !option.eligible)
                                                    .slice(0, 3)}
                                            >
                                                {(option) => (
                                                    <div class="rounded-md border border-red-100 bg-red-50 p-3 text-sm text-gray-600">
                                                        <div class="flex flex-wrap items-center gap-2">
                                                            <Badge content={option.partner.name} color="red" />
                                                            <Badge
                                                                content={`priority ${option.partner.routingPriority}`}
                                                                color="dark"
                                                            />
                                                            <Badge
                                                                content={`${option.partner.slaDays}d sla`}
                                                                color="dark"
                                                            />
                                                        </div>
                                                        <p class="mt-2">{option.reason}</p>
                                                        <p class="mt-1 text-xs text-gray-500">
                                                            Products:{' '}
                                                            {joinPartnerCapabilityList(
                                                                option.partner.supportedProductTypes
                                                            )}{' '}
                                                            · Regions:{' '}
                                                            {joinPartnerCapabilityList(option.partner.supportedRegions)}
                                                        </p>
                                                        <p class="mt-1 text-xs text-gray-500">
                                                            Partner base fulfillment:{' '}
                                                            {option.partner.baseFulfillmentCost || 'TBD'} · Region cost
                                                            rules:{' '}
                                                            {joinShippingCostRules(option.partner.shippingCostRules)}
                                                        </p>
                                                    </div>
                                                )}
                                            </For>
                                        </div>
                                    </Show>
                                    <Show when={recommendation().options.length === 0}>
                                        <EmptyBlock
                                            title="No active partner profiles returned"
                                            copy="Create or activate partner capabilities first so the routing engine can score eligible print and fulfillment partners."
                                        />
                                    </Show>
                                </div>
                            </div>
                        )}
                    </Show>
                    <FormSelectField
                        form={composer.form}
                        name="selectedExceptionType"
                        label="Default exception scenario"
                        options={exceptionOptions}
                    />
                    <Button type="submit" color="primary" loading={composer.form.isSubmitting()}>
                        {composer.form.values.manualPartnerOverride
                            ? 'Create routed order with override'
                            : 'Create routed order via auto-route'}
                    </Button>
                </form>
            </Show>
        </>
    )
}
