import { useParams } from '@tanstack/solid-router';
import { Show, createEffect, createSignal, onMount } from 'solid-js';
import { getPartner, type PartnerInfo } from '../../../services/partner';
import { tenantStorage } from '../../../services/tenantStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

function badgeColorForStatus(status: string) {
  return status === 'active' ? 'green' : 'dark';
}

function partnerTypeLabel(partnerType: string) {
  return partnerType.replaceAll('_', ' ');
}

function formatTimestamp(value?: string) {
  if (!value) return 'Not available';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

function DetailRow(props: { label: string; value: string }) {
  return (
    <div class="rounded-2xl border border-gray-200 p-4">
      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-500">
        {props.label}
      </p>
      <p class="mt-2 text-sm text-gray-900">{props.value}</p>
    </div>
  );
}

export default function TenantPartnerDetailPage() {
  const params = useParams({ from: '/t/$tenantId/partners/$partnerId' });

  const [partner, setPartner] = createSignal<PartnerInfo | null>(null);
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal('');

  const loadPartner = async () => {
    setLoading(true);
    setError('');
    try {
      const result = await getPartner(params().partnerId);
      if (!result.success) {
        setPartner(null);
        setError(result.message);
        return;
      }
      setPartner(result.data);
    } finally {
      setLoading(false);
    }
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  onMount(() => {
    void loadPartner();
  });

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
            void loadPartner();
          }}
        >
          Reload record
        </Button>
      </div>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <InfoAlert>
        This page reads a real partner record from the partner service. It is separate from the browser-local prototype flows used by product and order demo screens.
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
                <Badge
                  content={current().status}
                  color={badgeColorForStatus(current().status)}
                />
                <Badge
                  content={partnerTypeLabel(current().partnerType)}
                  color="indigo"
                />
                <Badge content={`store ${current().tenantId}`} color="blue" />
              </div>
            </Card>

            <div class="grid gap-4 md:grid-cols-2">
              <DetailRow label="Partner id" value={current().id} />
              <DetailRow label="Partner code" value={current().code || 'Not set'} />
              <DetailRow
                label="Contact name"
                value={current().contactName || 'Not set'}
              />
              <DetailRow
                label="Contact email"
                value={current().contactEmail || 'Not set'}
              />
              <DetailRow
                label="Created at"
                value={formatTimestamp(current().createdAt)}
              />
              <DetailRow
                label="Updated at"
                value={formatTimestamp(current().updatedAt)}
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
  );
}
