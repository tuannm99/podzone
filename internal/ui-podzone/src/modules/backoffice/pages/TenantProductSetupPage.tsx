import { useParams } from '@tanstack/solid-router'
import { For, Show, createEffect, createSignal } from 'solid-js'
import {
  createProductSetupDraft,
  getProductSetupSnapshot,
  promoteProductSetupCandidate,
  updateProductSetupCandidateStatus,
  type CatalogCandidate,
  type SetupDraft,
} from '@/services/productSetup'
import { tenantStorage } from '@/services/tenantStorage'
import { PageShell } from '@/solid/components/common/PageShell'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '@/solid/components/common/Feedback'
import { createFormStore, required } from '@/solid/forms'
import { useTenantWorkspace } from '@/solid/workspace/context'
import {
  candidateStatusColor,
  checklistCompletion,
  statusBadgeColor,
} from './product-setup/presentation'
import { moneyValue, productSetupInitialValues } from './product-setup/forms'
import { ProductSetupForm } from './product-setup/ProductSetupForm'

export default function TenantProductSetupPage() {
  const params = useParams({ from: '/t/$tenantId/products/setup' })
  const workspace = useTenantWorkspace()

  const setupForm = createFormStore({
    initialValues: productSetupInitialValues,
    validators: {
      name: [required('Enter a product name.')],
      baseCost: [moneyValue('Use a valid amount, for example $8.20.')],
      retailPrice: [moneyValue('Use a valid amount, for example $24.00.')],
      variantColor: [required('Enter a primary color.')],
      variantSize: [required('Enter a primary size.')],
    },
  })
  const [message, setMessage] = createSignal('')
  const [error, setError] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [promotingDraftID, setPromotingDraftID] = createSignal('')
  const [updatingCandidateID, setUpdatingCandidateID] = createSignal('')
  const [drafts, setDrafts] = createSignal<SetupDraft[]>([])
  const [candidates, setCandidates] = createSignal<CatalogCandidate[]>([])
  const currentStoreId = () => workspace?.currentStoreId() || ''
  const currentStore = () => workspace?.currentStore()
  const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
  const storeLabel = () =>
    currentStore()?.name || currentStoreId() || 'selected store'

  const loadState = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await getProductSetupSnapshot()
      if (!result.success) {
        setError(result.message)
        setDrafts([])
        setCandidates([])
        return
      }
      setDrafts(result.data.drafts)
      setCandidates(result.data.candidates)
    } finally {
      setLoading(false)
    }
  }

  const addDraft = async (event: SubmitEvent) => {
    event.preventDefault()
    if (!setupForm.validate()) {
      return
    }
    setupForm.setSubmitting(true)
    setError('')
    setMessage('')
    try {
      const result = await createProductSetupDraft({
        name: setupForm.values.name.trim(),
        partner: setupForm.values.partner.trim(),
        baseCost: setupForm.values.baseCost.trim(),
        retailPrice: setupForm.values.retailPrice.trim(),
        status: setupForm.values.status,
        notes: setupForm.values.notes.trim(),
      })
      if (!result.success) {
        setError(result.message)
        return
      }
      setMessage(`Saved backend product setup draft for ${result.data.name}.`)
      setupForm.reset()
      await loadState()
    } finally {
      setupForm.setSubmitting(false)
    }
  }

  const promoteToCandidate = async (draft: SetupDraft) => {
    setPromotingDraftID(draft.id)
    setError('')
    setMessage('')
    try {
      const result = await promoteProductSetupCandidate({
        draftId: draft.id,
        channel: setupForm.values.channel,
        variantColor: setupForm.values.variantColor.trim(),
        variantSize: setupForm.values.variantSize.trim(),
        artworkChecklist: {
          frontArtwork: setupForm.values.hasFrontArtwork,
          backArtwork: setupForm.values.hasBackArtwork,
          mockupReady: setupForm.values.mockupReady,
          printSpecChecked: setupForm.values.printSpecChecked,
        },
        merchandisingNotes: setupForm.values.notes.trim(),
      })
      if (!result.success) {
        setError(result.message)
        return
      }
      setMessage(`Promoted ${draft.name} into a backend catalog candidate.`)
      await loadState()
    } finally {
      setPromotingDraftID('')
    }
  }

  const updateCandidateStatus = async (
    candidateId: string,
    nextStatus: string,
    successMessage: string
  ) => {
    setUpdatingCandidateID(candidateId)
    setError('')
    setMessage('')
    try {
      const result = await updateProductSetupCandidateStatus(
        candidateId,
        nextStatus
      )
      if (!result.success) {
        setError(result.message)
        return
      }
      setMessage(successMessage)
      await loadState()
    } finally {
      setUpdatingCandidateID('')
    }
  }

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId)
    if (!workspaceReady()) {
      return
    }
    void loadState()
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
        Product setup drafts and candidates on this page are now backend-backed
        per store. Order routing remains a separate prototype flow.
      </InfoAlert>

      <InfoAlert>
        Suggested flow: create a draft, promote it into a candidate, then mock
        publish it before using that candidate in the local order-routing
        prototype.
      </InfoAlert>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      {message() ? <InfoAlert>{message()}</InfoAlert> : null}

      <Show when={loading()}>
        <LoadingInline label="Loading product setup..." />
      </Show>

      <div class="grid gap-6 lg:grid-cols-[0.92fr_1.08fr]">
        <Card class="space-y-4">
          <ProductSetupForm
            form={setupForm}
            saving={setupForm.isSubmitting}
            onSubmit={addDraft}
            onReset={() => setupForm.reset()}
          />
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Experimental setup queue"
            subtitle="A lightweight queue to validate what a future product setup workflow should feel like."
          />
          {drafts().length === 0 ? (
            <EmptyBlock
              title="No setup drafts yet"
              copy="Start with one POD product idea and shape the workflow before deeper architecture decisions are locked in."
            />
          ) : (
            <div class="space-y-3">
              <For each={drafts()}>
                {(draft) => (
                  <div class="rounded-lg border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">{draft.name}</p>
                        <p class="mt-1 text-sm text-gray-500">
                          partner {draft.partner} · cost {draft.baseCost} ·
                          retail {draft.retailPrice}
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
                          color="blue"
                          loading={promotingDraftID() === draft.id}
                          onClick={() => {
                            void promoteToCandidate(draft)
                          }}
                        >
                          Promote to candidate
                        </Button>
                      </div>
                    </div>
                    {draft.notes ? (
                      <p class="mt-3 text-sm text-gray-600">{draft.notes}</p>
                    ) : null}
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
        {candidates().length === 0 ? (
          <EmptyBlock
            title="No catalog candidates yet"
            copy="Promote one of the setup drafts to test how a POD product should move toward publishable state."
          />
        ) : (
          <div class="space-y-3">
            <For each={candidates()}>
              {(candidate) => (
                <div class="rounded-lg border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">
                        {candidate.title}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        sku {candidate.sku} · partner {candidate.partner} ·
                        channel {candidate.channel.replaceAll('_', ' ')}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        cost {candidate.baseCost} · retail{' '}
                        {candidate.retailPrice} · est. margin{' '}
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
                          updatingCandidateID() === candidate.id
                        }
                        loading={updatingCandidateID() === candidate.id}
                        onClick={() => {
                          void updateCandidateStatus(
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
                          updatingCandidateID() === candidate.id
                        }
                        loading={updatingCandidateID() === candidate.id}
                        onClick={() => {
                          void updateCandidateStatus(
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
                          {candidate.artworkChecklist.frontArtwork
                            ? 'ready'
                            : 'pending'}
                        </li>
                        <li>
                          Back artwork:{' '}
                          {candidate.artworkChecklist.backArtwork
                            ? 'ready'
                            : 'pending'}
                        </li>
                        <li>
                          Mockups:{' '}
                          {candidate.artworkChecklist.mockupReady
                            ? 'ready'
                            : 'pending'}
                        </li>
                        <li>
                          Print specs:{' '}
                          {candidate.artworkChecklist.printSpecChecked
                            ? 'ready'
                            : 'pending'}
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
                  <p class="mt-3 text-sm text-gray-600">
                    {candidate.merchandisingNotes}
                  </p>
                  <p class="mt-3 text-sm text-gray-600">
                    Last updated{' '}
                    {new Date(candidate.updatedAt).toLocaleString()}
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
