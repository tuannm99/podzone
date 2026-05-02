import { useParams } from '@tanstack/solid-router';
import { For, createEffect, createSignal } from 'solid-js';
import { tenantStorage } from '../../../services/tenantStorage';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Button, Card, InputField, SelectField, TextareaField } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { EmptyBlock, InfoAlert } from '../../components/common/Feedback';

type SetupDraft = {
  id: string;
  name: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  status: string;
  notes: string;
};

type CandidateVariant = {
  id: string;
  label: string;
  color: string;
  size: string;
  status: string;
};

type ArtworkChecklist = {
  frontArtwork: boolean;
  backArtwork: boolean;
  mockupReady: boolean;
  printSpecChecked: boolean;
};

type CatalogCandidate = {
  id: string;
  draftId: string;
  title: string;
  sku: string;
  partner: string;
  baseCost: string;
  retailPrice: string;
  estimatedMargin: string;
  status: string;
  channel: string;
  updatedAt: string;
  variants: CandidateVariant[];
  artworkChecklist: ArtworkChecklist;
  merchandisingNotes: string;
};

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

function productSetupStorageKey(tenantId: string) {
  return `podzone:product-setup:${tenantId}`;
}

function seedDrafts(): SetupDraft[] {
  return [
    {
      id: 'draft-tee-001',
      name: 'Signature Tee',
      partner: 'Acme Print Lab',
      baseCost: '$8.20',
      retailPrice: '$24.00',
      status: 'ready_for_review',
      notes: 'Front print approved. Need final mockup export.',
    },
    {
      id: 'draft-mug-002',
      name: 'Ceramic Mug',
      partner: 'North Fulfillment',
      baseCost: '$5.10',
      retailPrice: '$18.00',
      status: 'draft',
      notes: 'Handle packaging variant before publish.',
    },
  ];
}

function parseMoney(value: string) {
  const cleaned = value.replace(/[^0-9.]/g, '');
  const parsed = Number.parseFloat(cleaned);
  return Number.isFinite(parsed) ? parsed : null;
}

function formatMoney(value: number) {
  return `$${value.toFixed(2)}`;
}

function estimateMargin(baseCost: string, retailPrice: string) {
  const base = parseMoney(baseCost);
  const retail = parseMoney(retailPrice);
  if (base === null || retail === null) return 'TBD';
  return formatMoney(retail - base);
}

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

  const persistState = (nextDrafts: SetupDraft[], nextCandidates: CatalogCandidate[]) => {
    localStorage.setItem(
      productSetupStorageKey(params().tenantId),
      JSON.stringify({
        drafts: nextDrafts,
        candidates: nextCandidates,
      })
    );
  };

  const loadState = () => {
    const raw = localStorage.getItem(productSetupStorageKey(params().tenantId));
    if (!raw) {
      const seededDrafts = seedDrafts();
      setDrafts(seededDrafts);
      setCandidates([]);
      persistState(seededDrafts, []);
      return;
    }

    try {
      const parsed = JSON.parse(raw) as {
        drafts?: SetupDraft[];
        candidates?: CatalogCandidate[];
      };
      setDrafts(parsed.drafts || seedDrafts());
      setCandidates(parsed.candidates || []);
    } catch {
      const seededDrafts = seedDrafts();
      setDrafts(seededDrafts);
      setCandidates([]);
      persistState(seededDrafts, []);
    }
  };

  const addDraft = (event: SubmitEvent) => {
    event.preventDefault();
    if (!name().trim()) return;
    const nextDraft: SetupDraft = {
      id: `draft-${Date.now()}`,
      name: name().trim(),
      partner: partner().trim() || 'Unassigned',
      baseCost: baseCost().trim() || 'TBD',
      retailPrice: retailPrice().trim() || 'TBD',
      status: status(),
      notes: notes().trim(),
    };
    const nextDrafts = [nextDraft, ...drafts()];
    setDrafts(nextDrafts);
    persistState(nextDrafts, candidates());
    setMessage(`Saved experimental product setup draft for ${nextDraft.name}.`);
    resetForm();
  };

  const promoteToCandidate = (draft: SetupDraft) => {
    const existing = candidates().find((candidate) => candidate.draftId === draft.id);
    if (existing) {
      setMessage(`Catalog candidate for ${draft.name} already exists.`);
      return;
    }

    const nextCandidate: CatalogCandidate = {
      id: `candidate-${Date.now()}`,
      draftId: draft.id,
      title: draft.name,
      sku: `${draft.name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'product'}-${Date.now().toString().slice(-4)}`,
      partner: draft.partner,
      baseCost: draft.baseCost,
      retailPrice: draft.retailPrice,
      estimatedMargin: estimateMargin(draft.baseCost, draft.retailPrice),
      status: 'ready',
      channel: channel(),
      updatedAt: new Date().toISOString(),
      variants: [
        {
          id: `variant-${Date.now()}`,
          label: `${variantColor()} / ${variantSize()}`,
          color: variantColor(),
          size: variantSize(),
          status: 'ready',
        },
      ],
      artworkChecklist: {
        frontArtwork: hasFrontArtwork(),
        backArtwork: hasBackArtwork(),
        mockupReady: mockupReady(),
        printSpecChecked: printSpecChecked(),
      },
      merchandisingNotes: draft.notes || 'No extra merchandising notes yet.',
    };

    const nextCandidates = [nextCandidate, ...candidates()];
    setCandidates(nextCandidates);
    persistState(drafts(), nextCandidates);
    setMessage(`Promoted ${draft.name} into a catalog candidate.`);
  };

  const updateCandidateStatus = (candidateId: string, nextStatus: string) => {
    const nextCandidates = candidates().map((candidate) =>
      candidate.id === candidateId
        ? {
            ...candidate,
            status: nextStatus,
            updatedAt: new Date().toISOString(),
          }
        : candidate
    );
    setCandidates(nextCandidates);
    persistState(drafts(), nextCandidates);
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  createEffect(() => {
    loadState();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Product Setup"
          title={`POD product setup for store ${params().tenantId}`}
          copy="This is an experimental workspace for shaping products before they become real catalog records. It stays intentionally lightweight while the architecture is still evolving."
        />
      </Card>

      <InfoAlert>
        This screen is intentionally local-first. It does not depend on external partner, catalog, or cloud integrations yet.
      </InfoAlert>

      <InfoAlert>
        Typical demo flow: seed a store from the workspace home, refine one draft here, promote it to a candidate, then mock publish before testing order routing.
      </InfoAlert>

      {message() ? <InfoAlert>{message()}</InfoAlert> : null}

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
              <Button type="submit">Save setup draft</Button>
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
                          onClick={() => {
                            promoteToCandidate(draft);
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
          subtitle="A mock publishing layer that lets the team test product readiness before any real channel integration exists."
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
                            updateCandidateStatus(candidate.id, 'published_mock');
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
                          updateCandidateStatus(candidate.id, 'archived');
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
