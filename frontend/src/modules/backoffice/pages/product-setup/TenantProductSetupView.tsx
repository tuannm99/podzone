import { useParams } from '@tanstack/solid-router'
import { For, Show } from 'solid-js'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { useTenantWorkspace } from '@/modules/shell/workspace/context'
import { createProductSetupViewModel } from './createProductSetupViewModel'
import { candidateStatusColor, checklistCompletion, statusBadgeColor } from './presentation'
import { ProductSetupForm } from './ProductSetupForm'

export function TenantProductSetupView() {
    const params = useParams({ from: '/t/$tenantId/products/setup' })
    const workspace = useTenantWorkspace()

    const currentStoreId = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreId() || 'selected store'
    const setup = createProductSetupViewModel({
        tenantID: () => params().tenantId,
        workspaceReady,
    })

    return (
        <PageShell>
            <Card class="space-y-4">
                <SectionLead
                    eyebrow="Product Setup Prototype"
                    title={`POD product setup for ${storeLabel()}`}
                    copy="This workspace now persists store-scoped product setup drafts and candidates in the backend, while still using lightweight POD-first language for early operational shaping."
                />
            </Card>

            <Show when={!workspaceReady()}>
                <EmptyBlock
                    title="Choose a store first"
                    copy="Use the workspace store switcher before loading store-scoped product setup drafts and candidates."
                />
            </Show>

            <InfoAlert>
                Product setup drafts and candidates on this page are now backend-backed per store. Order routing remains
                a separate prototype flow.
            </InfoAlert>

            <InfoAlert>
                Suggested flow: create a draft, promote it into a candidate, then mock publish it before using that
                candidate in the local order-routing prototype.
            </InfoAlert>

            <Show when={setup.error()}>
                <ErrorAlert>{setup.error()}</ErrorAlert>
            </Show>

            {setup.message() ? <InfoAlert>{setup.message()}</InfoAlert> : null}

            <Show when={setup.loading()}>
                <LoadingInline label="Loading product setup..." />
            </Show>

            <div class="grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
                <Card class="space-y-4">
                    <ProductSetupForm
                        form={setup.form}
                        saving={setup.form.isSubmitting}
                        onSubmit={setup.addDraft}
                        onReset={() => setup.form.reset()}
                    />
                </Card>

                <Card class="space-y-4">
                    <SectionTitle
                        title="Experimental setup queue"
                        subtitle="A lightweight queue to validate what a future product setup workflow should feel like."
                    />
                    {setup.drafts().length === 0 ? (
                        <EmptyBlock
                            title="No setup drafts yet"
                            copy="Start with one POD product idea and shape the workflow before deeper architecture decisions are locked in."
                        />
                    ) : (
                        <div class="space-y-3">
                            <For each={setup.drafts()}>
                                {(draft) => (
                                    <div class="rounded-lg border border-gray-200 p-4">
                                        <div class="flex flex-wrap items-center justify-between gap-3">
                                            <div>
                                                <p class="font-semibold text-gray-900">{draft.name}</p>
                                                <p class="mt-1 text-sm text-gray-500">
                                                    partner {draft.partner} · cost {draft.baseCost} · retail{' '}
                                                    {draft.retailPrice}
                                                </p>
                                            </div>
                                            <div class="flex flex-wrap items-center gap-2">
                                                <Badge
                                                    content={draft.status.replaceAll('_', ' ')}
                                                    color={statusBadgeColor(draft.status)}
                                                />
                                                <Button
                                                    type="button"
                                                    size="xs"
                                                    color="primary"
                                                    loading={setup.promotingDraftID() === draft.id}
                                                    onClick={() => {
                                                        void setup.promoteToCandidate(draft)
                                                    }}
                                                >
                                                    Promote to candidate
                                                </Button>
                                            </div>
                                        </div>
                                        {draft.notes ? <p class="mt-3 text-sm text-gray-600">{draft.notes}</p> : null}
                                    </div>
                                )}
                            </For>
                        </div>
                    )}
                </Card>
            </div>

            <Card class="mt-6 space-y-4">
                <SectionTitle
                    title="Catalog candidates"
                    subtitle="A prototype publishing layer that lets the team test product readiness before any real channel integration exists."
                />
                {setup.candidates().length === 0 ? (
                    <EmptyBlock
                        title="No catalog candidates yet"
                        copy="Promote one of the setup drafts to test how a POD product should move toward publishable state."
                    />
                ) : (
                    <div class="space-y-3">
                        <For each={setup.candidates()}>
                            {(candidate) => (
                                <div class="rounded-lg border border-gray-200 p-4">
                                    <div class="flex flex-wrap items-center justify-between gap-3">
                                        <div>
                                            <p class="font-semibold text-gray-900">{candidate.title}</p>
                                            <p class="mt-1 text-sm text-gray-500">
                                                sku {candidate.sku} · partner {candidate.partner} · channel{' '}
                                                {candidate.channel.replaceAll('_', ' ')}
                                            </p>
                                            <p class="mt-1 text-sm text-gray-500">
                                                cost {candidate.baseCost} · retail {candidate.retailPrice} · est. margin{' '}
                                                {candidate.estimatedMargin}
                                            </p>
                                        </div>
                                        <div class="flex flex-wrap items-center gap-2">
                                            <Badge
                                                content={candidate.status.replaceAll('_', ' ')}
                                                color={candidateStatusColor(candidate.status)}
                                            />
                                            <Button
                                                type="button"
                                                size="xs"
                                                color="green"
                                                disabled={
                                                    candidate.status === 'published_mock' ||
                                                    !candidate.artworkChecklist.mockupReady ||
                                                    !candidate.artworkChecklist.printSpecChecked ||
                                                    setup.updatingCandidateID() === candidate.id
                                                }
                                                loading={setup.updatingCandidateID() === candidate.id}
                                                onClick={() => {
                                                    void setup.updateCandidateStatus(
                                                        candidate.id,
                                                        'published_mock',
                                                        `Mock published ${candidate.title}.`
                                                    )
                                                }}
                                            >
                                                Mock publish
                                            </Button>
                                            <Button
                                                type="button"
                                                size="xs"
                                                color="light"
                                                disabled={
                                                    candidate.status === 'archived' ||
                                                    setup.updatingCandidateID() === candidate.id
                                                }
                                                loading={setup.updatingCandidateID() === candidate.id}
                                                onClick={() => {
                                                    void setup.updateCandidateStatus(
                                                        candidate.id,
                                                        'archived',
                                                        `Archived candidate ${candidate.title}.`
                                                    )
                                                }}
                                            >
                                                Archive
                                            </Button>
                                        </div>
                                    </div>
                                    <div class="mt-3 grid gap-3 md:grid-cols-2">
                                        <div class="rounded-md bg-gray-50 p-3">
                                            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
                                                Artwork checklist
                                            </p>
                                            <p class="mt-2 text-sm text-gray-700">
                                                {checklistCompletion(candidate.artworkChecklist)}
                                            </p>
                                            <ul class="mt-2 space-y-1 text-sm text-gray-600">
                                                <li>
                                                    Front artwork:{' '}
                                                    {candidate.artworkChecklist.frontArtwork ? 'ready' : 'pending'}
                                                </li>
                                                <li>
                                                    Back artwork:{' '}
                                                    {candidate.artworkChecklist.backArtwork ? 'ready' : 'pending'}
                                                </li>
                                                <li>
                                                    Mockups:{' '}
                                                    {candidate.artworkChecklist.mockupReady ? 'ready' : 'pending'}
                                                </li>
                                                <li>
                                                    Print specs:{' '}
                                                    {candidate.artworkChecklist.printSpecChecked ? 'ready' : 'pending'}
                                                </li>
                                            </ul>
                                        </div>
                                        <div class="rounded-md bg-gray-50 p-3">
                                            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
                                                Variant starter
                                            </p>
                                            <div class="mt-2 space-y-1 text-sm text-gray-600">
                                                <For each={candidate.variants}>
                                                    {(variant) => (
                                                        <p>
                                                            {variant.label} · {variant.status}
                                                        </p>
                                                    )}
                                                </For>
                                            </div>
                                        </div>
                                    </div>
                                    <p class="mt-3 text-sm text-gray-600">{candidate.merchandisingNotes}</p>
                                    <p class="mt-3 text-sm text-gray-600">
                                        Last updated {new Date(candidate.updatedAt).toLocaleString()}
                                    </p>
                                </div>
                            )}
                        </For>
                    </div>
                )}
            </Card>
        </PageShell>
    )
}
