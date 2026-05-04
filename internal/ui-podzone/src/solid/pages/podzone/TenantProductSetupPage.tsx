import { useParams } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal, onMount } from 'solid-js';
import {
  createProductSetupDraft,
  getProductSetupSnapshot,
  promoteProductSetupCandidate,
  updateProductSetupCandidateStatus,
  type ArtworkChecklist,
  type CatalogCandidate,
  type SetupDraft,
} from '../../../services/productSetup';
import { tenantStorage } from '../../../services/tenantStorage';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Button, Card, InputField, SelectField, TextareaField } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '../../components/common/Feedback';

const statusOptions = [
  { name: 'Draft', value: 'draft' },
  { name: 'Ready for review', value: 'ready_for_review' },
  { name: 'Publish candidate', value: 'publish_candidate' },
];

const channelOptions = [
  { name: 'Website store', value: 'website_store' },
  { name: 'Marketplace mock', value: 'marketplace_mock' },
  { name: 'Wholesale mock', value: 'wholesale_mock' },
];

function statusBadgeColor(status: string) {
  switch (status) {
    case 'publish_candidate':
      return 'green';
    case 'ready_for_review':
      return 'yellow';
    default:
      return 'dark';
  }
}

function candidateStatusColor(status: string) {
  switch (status) {
    case 'published_mock':
      return 'green';
    case 'ready':
      return 'blue';
    case 'archived':
      return 'dark';
    default:
      return 'yellow';
  }
}

function checklistCompletion(checklist: ArtworkChecklist) {
  const completed = [
    checklist.frontArtwork,
    checklist.backArtwork,
    checklist.mockupReady,
    checklist.printSpecChecked,
  ].filter(Boolean).length;
  return `${completed}/4 ready`;
}

export default function TenantProductSetupPage() {
  const params = useParams({ from: '/t/$tenantId/products/setup' });

  const [name, setName] = createSignal('');
  const [partner, setPartner] = createSignal('');
  const [baseCost, setBaseCost] = createSignal('');
  const [retailPrice, setRetailPrice] = createSignal('');
  const [status, setStatus] = createSignal('draft');
  const [notes, setNotes] = createSignal('');
  const [channel, setChannel] = createSignal('website_store');
  const [variantColor, setVariantColor] = createSignal('Black');
  const [variantSize, setVariantSize] = createSignal('M');
  const [hasFrontArtwork, setHasFrontArtwork] = createSignal(true);
  const [hasBackArtwork, setHasBackArtwork] = createSignal(false);
  const [mockupReady, setMockupReady] = createSignal(false);
  const [printSpecChecked, setPrintSpecChecked] = createSignal(false);
  const [message, setMessage] = createSignal('');
  const [error, setError] = createSignal('');
  const [loading, setLoading] = createSignal(false);
  const [savingDraft, setSavingDraft] = createSignal(false);
  const [promotingDraftID, setPromotingDraftID] = createSignal('');
  const [drafts, setDrafts] = createSignal<SetupDraft[]>([]);
  const [candidates, setCandidates] = createSignal<CatalogCandidate[]>([]);

  const resetForm = () => {
    setName('');
    setPartner('');
    setBaseCost('');
    setRetailPrice('');
    setStatus('draft');
    setNotes('');
    setChannel('website_store');
    setVariantColor('Black');
    setVariantSize('M');
    setHasFrontArtwork(true);
    setHasBackArtwork(false);
    setMockupReady(false);
    setPrintSpecChecked(false);
  };

  const loadState = async () => {
    setLoading(true);
    setError('');
    try {
      const result = await getProductSetupSnapshot();
      if (!result.success) {
        setError(result.message);
        setDrafts([]);
        setCandidates([]);
        return;
      }
      setDrafts(result.data.drafts);
      setCandidates(result.data.candidates);
    } finally {
      setLoading(false);
    }
  };

  const addDraft = async (event: SubmitEvent) => {
    event.preventDefault();
    if (!name().trim()) {
      setError('Product name is required.');
      return;
    }
    setSavingDraft(true);
    setError('');
    setMessage('');
    const result = await createProductSetupDraft({
      name: name().trim(),
      partner: partner().trim(),
      baseCost: baseCost().trim(),
      retailPrice: retailPrice().trim(),
      status: status(),
      notes: notes().trim(),
    });
    setSavingDraft(false);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setMessage(`Saved backend product setup draft for ${result.data.name}.`);
    resetForm();
    await loadState();
  };

  const promoteToCandidate = async (draft: SetupDraft) => {
    setPromotingDraftID(draft.id);
    setError('');
    setMessage('');
    const result = await promoteProductSetupCandidate({
      draftId: draft.id,
      channel: channel(),
      variantColor: variantColor(),
      variantSize: variantSize(),
      artworkChecklist: {
        frontArtwork: hasFrontArtwork(),
        backArtwork: hasBackArtwork(),
        mockupReady: mockupReady(),
        printSpecChecked: printSpecChecked(),
      },
      merchandisingNotes: notes().trim(),
    });
    setPromotingDraftID('');
    if (!result.success) {
      setError(result.message);
      return;
    }
    setMessage(`Promoted ${draft.name} into a backend catalog candidate.`);
    await loadState();
  };

  const updateCandidateStatus = async (candidateId: string, nextStatus: string) => {
    setError('');
    const result = await updateProductSetupCandidateStatus(candidateId, nextStatus);
    if (!result.success) {
      setError(result.message);
      return;
    }
    await loadState();
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  onMount(() => {
    void loadState();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Product Setup Prototype"
          title={`POD product setup for store ${params().tenantId}`}
          copy="This workspace now persists store-scoped product setup drafts and candidates in the backend, while still using lightweight POD-first language for early operational shaping."
        />
      </Card>

      <InfoAlert>
        Product setup drafts and candidates on this page are now backend-backed per store. Order routing remains a separate prototype flow.
      </InfoAlert>

      <InfoAlert>
        Suggested flow: create a draft, promote it into a candidate, then mock publish it before using that candidate in the local order-routing prototype.
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
          <SectionTitle
            title="New setup draft"
            subtitle="Capture the commercial and execution basics for a POD product before deeper automation exists."
          />
          <form class="space-y-4" onSubmit={addDraft}>
            <InputField
              label="Product name"
              value={name()}
              placeholder="Signature Tee"
              onInput={(event) => setName(event.currentTarget.value)}
            />
            <InputField
              label="Preferred print partner"
              value={partner()}
              placeholder="Acme Print Lab"
              onInput={(event) => setPartner(event.currentTarget.value)}
            />
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Base cost"
                value={baseCost()}
                placeholder="$8.20"
                onInput={(event) => setBaseCost(event.currentTarget.value)}
              />
              <InputField
                label="Retail price"
                value={retailPrice()}
                placeholder="$24.00"
                onInput={(event) => setRetailPrice(event.currentTarget.value)}
              />
            </div>
            <SelectField
              label="Draft status"
              value={status()}
              options={statusOptions}
              onChange={(event) => setStatus(event.currentTarget.value)}
            />
            <SelectField
              label="Mock publish channel"
              value={channel()}
              options={channelOptions}
              onChange={(event) => setChannel(event.currentTarget.value)}
            />
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Primary color"
                value={variantColor()}
                placeholder="Black"
                onInput={(event) => setVariantColor(event.currentTarget.value)}
              />
              <InputField
                label="Primary size"
                value={variantSize()}
                placeholder="M"
                onInput={(event) => setVariantSize(event.currentTarget.value)}
              />
            </div>
            <div class="rounded-2xl border border-gray-200 p-4">
              <p class="text-sm font-semibold text-gray-900">Artwork readiness</p>
              <div class="mt-3 grid gap-3 md:grid-cols-2">
                <label class="flex items-center gap-2 text-sm text-gray-600">
                  <input
                    type="checkbox"
                    checked={hasFrontArtwork()}
                    onChange={(event) => setHasFrontArtwork(event.currentTarget.checked)}
                  />
                  Front artwork prepared
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-600">
                  <input
                    type="checkbox"
                    checked={hasBackArtwork()}
                    onChange={(event) => setHasBackArtwork(event.currentTarget.checked)}
                  />
                  Back artwork prepared
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-600">
                  <input
                    type="checkbox"
                    checked={mockupReady()}
                    onChange={(event) => setMockupReady(event.currentTarget.checked)}
                  />
                  Mockups exported
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-600">
                  <input
                    type="checkbox"
                    checked={printSpecChecked()}
                    onChange={(event) => setPrintSpecChecked(event.currentTarget.checked)}
                  />
                  Print specs checked
                </label>
              </div>
            </div>
            <TextareaField
              label="Setup notes"
              value={notes()}
              rows={4}
              onInput={(event) => setNotes(event.currentTarget.value)}
            />
            <div class="flex flex-wrap gap-3">
              <Button type="submit" loading={savingDraft()}>
                Save setup draft
              </Button>
              <Button type="button" color="alternative" onClick={resetForm}>
                Clear form
              </Button>
            </div>
          </form>
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
                  <div class="rounded-2xl border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">{draft.name}</p>
                        <p class="mt-1 text-sm text-gray-500">
                          partner {draft.partner} · cost {draft.baseCost} · retail {draft.retailPrice}
                        </p>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge content={draft.status.replaceAll('_', ' ')} color={statusBadgeColor(draft.status)} />
                        <Button
                          type="button"
                          size="xs"
                          color="blue"
                          loading={promotingDraftID() === draft.id}
                          onClick={() => {
                            void promoteToCandidate(draft);
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
                <div class="rounded-2xl border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">{candidate.title}</p>
                        <p class="mt-1 text-sm text-gray-500">
                          sku {candidate.sku} · partner {candidate.partner} · channel {candidate.channel.replaceAll('_', ' ')}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        cost {candidate.baseCost} · retail {candidate.retailPrice} · est. margin {candidate.estimatedMargin}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={candidate.status.replaceAll('_', ' ')} color={candidateStatusColor(candidate.status)} />
                        <Button
                          type="button"
                          size="xs"
                          color="green"
                          disabled={
                            candidate.status === 'published_mock' ||
                            !candidate.artworkChecklist.mockupReady ||
                            !candidate.artworkChecklist.printSpecChecked
                          }
                          onClick={() => {
                            void updateCandidateStatus(candidate.id, 'published_mock');
                            setMessage(`Mock published ${candidate.title}.`);
                          }}
                        >
                        Mock publish
                      </Button>
                      <Button
                        type="button"
                        size="xs"
                        color="light"
                        disabled={candidate.status === 'archived'}
                        onClick={() => {
                          void updateCandidateStatus(candidate.id, 'archived');
                          setMessage(`Archived candidate ${candidate.title}.`);
                        }}
                      >
                        Archive
                      </Button>
                    </div>
                  </div>
                  <div class="mt-3 grid gap-3 md:grid-cols-2">
                    <div class="rounded-xl bg-gray-50 p-3">
                      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
                        Artwork checklist
                      </p>
                      <p class="mt-2 text-sm text-gray-700">
                        {checklistCompletion(candidate.artworkChecklist)}
                      </p>
                      <ul class="mt-2 space-y-1 text-sm text-gray-600">
                        <li>
                          Front artwork: {candidate.artworkChecklist.frontArtwork ? 'ready' : 'pending'}
                        </li>
                        <li>
                          Back artwork: {candidate.artworkChecklist.backArtwork ? 'ready' : 'pending'}
                        </li>
                        <li>
                          Mockups: {candidate.artworkChecklist.mockupReady ? 'ready' : 'pending'}
                        </li>
                        <li>
                          Print specs: {candidate.artworkChecklist.printSpecChecked ? 'ready' : 'pending'}
                        </li>
                      </ul>
                    </div>
                    <div class="rounded-xl bg-gray-50 p-3">
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
  );
}
