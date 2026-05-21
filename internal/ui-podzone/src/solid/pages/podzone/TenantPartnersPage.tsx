import { useParams } from '@tanstack/solid-router';
import { For, Show, createEffect, createSignal, onMount } from 'solid-js';
import {
  createPartner,
  listPartners,
  updatePartner,
  updatePartnerStatus,
  type PartnerInfo,
} from '../../../services/partner';
import { tenantStorage } from '../../../services/tenantStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import {
  Badge,
  Button,
  Card,
  InputField,
  SelectField,
  TextareaField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const partnerTypeOptions = [
  { name: 'All partner types', value: '' },
  { name: 'Print on demand', value: 'print_on_demand' },
  { name: 'Fulfillment partner', value: 'fulfillment' },
  { name: 'Dropship supplier', value: 'dropship_supplier' },
];

const partnerStatusOptions = [
  { name: 'All statuses', value: '' },
  { name: 'Active', value: 'active' },
  { name: 'Inactive', value: 'inactive' },
];

function badgeColorForStatus(status: string) {
  return status === 'active' ? 'green' : 'dark';
}

function partnerTypeLabel(partnerType: string) {
  return partnerType.replaceAll('_', ' ');
}

function joinCapabilityList(items: string[]) {
  return (items || []).join(', ');
}

function parseCapabilityList(raw: string) {
  return raw
    .split(',')
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean);
}

function joinShippingCostRules(
  rules: { region: string; cost: string }[] | undefined
) {
  return (rules || [])
    .map((rule) => `${rule.region}:${rule.cost}`)
    .join(', ');
}

function parseShippingCostRules(raw: string) {
  return raw
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      const [region, ...costParts] = item.split(':');
      return {
        region: region.trim().toLowerCase(),
        cost: costParts.join(':').trim(),
      };
    })
    .filter((item) => item.region && item.cost);
}

export default function TenantPartnersPage() {
  const params = useParams({ from: '/t/$tenantId/partners' });

  const [partners, setPartners] = createSignal<PartnerInfo[]>([]);
  const [loading, setLoading] = createSignal(false);
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal('');
  const [message, setMessage] = createSignal('');

  const [name, setName] = createSignal('');
  const [code, setCode] = createSignal('');
  const [contactName, setContactName] = createSignal('');
  const [contactEmail, setContactEmail] = createSignal('');
  const [notes, setNotes] = createSignal('');
  const [partnerType, setPartnerType] = createSignal('print_on_demand');
  const [supportedProductTypes, setSupportedProductTypes] = createSignal('tshirt, hoodie');
  const [supportedRegions, setSupportedRegions] = createSignal('us, eu');
  const [slaDays, setSlaDays] = createSignal('3');
  const [routingPriority, setRoutingPriority] = createSignal('100');
  const [baseFulfillmentCost, setBaseFulfillmentCost] = createSignal('$8.50');
  const [shippingCostRules, setShippingCostRules] = createSignal(
    'us:$4.50, eu:$6.00'
  );
  const [editingPartnerId, setEditingPartnerId] = createSignal('');
  const [filterPartnerType, setFilterPartnerType] = createSignal('');
  const [filterStatus, setFilterStatus] = createSignal('');

  const isEditing = () => editingPartnerId().trim().length > 0;

  const resetForm = () => {
    setEditingPartnerId('');
    setName('');
    setCode('');
    setContactName('');
    setContactEmail('');
    setNotes('');
    setPartnerType('print_on_demand');
    setSupportedProductTypes('tshirt, hoodie');
    setSupportedRegions('us, eu');
    setSlaDays('3');
    setRoutingPriority('100');
    setBaseFulfillmentCost('$8.50');
    setShippingCostRules('us:$4.50, eu:$6.00');
  };

  const loadPartners = async () => {
    setLoading(true);
    setError('');
    try {
      const result = await listPartners(
        params().tenantId,
        filterPartnerType(),
        filterStatus()
      );
      if (!result.success) {
        setError(result.message);
        setPartners([]);
        return;
      }
      setPartners(result.data);
    } finally {
      setLoading(false);
    }
  };

  const submit = async (event: SubmitEvent) => {
    event.preventDefault();
    if (!name().trim()) {
      setError('Partner name is required.');
      return;
    }
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const result = isEditing()
        ? await updatePartner({
            id: editingPartnerId(),
            name: name().trim(),
            contactName: contactName().trim(),
            contactEmail: contactEmail().trim(),
            notes: notes().trim(),
            partnerType: partnerType(),
            supportedProductTypes: parseCapabilityList(supportedProductTypes()),
            supportedRegions: parseCapabilityList(supportedRegions()),
            slaDays: Math.max(0, Number.parseInt(slaDays(), 10) || 0),
            routingPriority: Math.max(0, Number.parseInt(routingPriority(), 10) || 0),
            baseFulfillmentCost: baseFulfillmentCost().trim(),
            shippingCostRules: parseShippingCostRules(shippingCostRules()),
          })
        : await createPartner({
            tenantId: params().tenantId,
            code: code().trim(),
            name: name().trim(),
            contactName: contactName().trim(),
            contactEmail: contactEmail().trim(),
            notes: notes().trim(),
            partnerType: partnerType(),
            supportedProductTypes: parseCapabilityList(supportedProductTypes()),
            supportedRegions: parseCapabilityList(supportedRegions()),
            slaDays: Math.max(0, Number.parseInt(slaDays(), 10) || 0),
            routingPriority: Math.max(0, Number.parseInt(routingPriority(), 10) || 0),
            baseFulfillmentCost: baseFulfillmentCost().trim(),
            shippingCostRules: parseShippingCostRules(shippingCostRules()),
          });
      if (!result.success) {
        setError(result.message);
        return;
      }
      const actionLabel = isEditing() ? 'Updated' : 'Created';
      resetForm();
      setMessage(`${actionLabel} print partner ${result.data.name}.`);
      await loadPartners();
    } finally {
      setSaving(false);
    }
  };

  const toggleStatus = async (partner: PartnerInfo) => {
    setError('');
    setMessage('');
    const nextStatus = partner.status === 'active' ? 'inactive' : 'active';
    const result = await updatePartnerStatus(partner.id, nextStatus);
    if (!result.success) {
      setError(result.message);
      return;
    }
    setMessage(
      `${result.data.name} is now ${result.data.status === 'active' ? 'active' : 'inactive'}.`
    );
    await loadPartners();
  };

  const startEdit = (partner: PartnerInfo) => {
    setEditingPartnerId(partner.id);
    setName(partner.name);
    setCode(partner.code);
    setContactName(partner.contactName || '');
    setContactEmail(partner.contactEmail || '');
    setNotes(partner.notes || '');
    setPartnerType(partner.partnerType || 'print_on_demand');
    setSupportedProductTypes(joinCapabilityList(partner.supportedProductTypes || []));
    setSupportedRegions(joinCapabilityList(partner.supportedRegions || []));
    setSlaDays(String(partner.slaDays || 0));
    setRoutingPriority(String(partner.routingPriority || 0));
    setBaseFulfillmentCost(partner.baseFulfillmentCost || '');
    setShippingCostRules(joinShippingCostRules(partner.shippingCostRules));
    setError('');
    setMessage(`Editing print partner ${partner.name}.`);
  };

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  onMount(() => {
    void loadPartners();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Print Partners"
          title={`Execution partners for store ${params().tenantId}`}
          copy="Manage the real backend-backed partner records that this store can use later for production and fulfillment workflows."
        />
      </Card>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <InfoAlert>
        This is the real partner record layer behind the demo flow. Seed and reset from store home only affect local product and order mock data.
      </InfoAlert>

      <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <Card class="space-y-4">
          <SectionTitle
            title={isEditing() ? 'Edit print partner' : 'Add print partner'}
            subtitle={
              isEditing()
                ? 'Update partner details without leaving the store workspace.'
                : 'Create a partner record for production, fulfillment, or future sourced-product workflows.'
            }
          />

          <form class="space-y-4" onSubmit={submit}>
            <InputField
              label="Partner name"
              value={name()}
              placeholder="Acme Print Lab"
              onInput={(event) => setName(event.currentTarget.value)}
            />
            <InputField
              label="Partner code"
              value={code()}
              placeholder="acme-print"
              onInput={(event) => setCode(event.currentTarget.value)}
            />
            <Show when={isEditing()}>
              <InfoAlert>
                Partner code is currently locked during edit so external references stay stable.
              </InfoAlert>
            </Show>
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Contact name"
                value={contactName()}
                placeholder="Linh Tran"
                onInput={(event) => setContactName(event.currentTarget.value)}
              />
              <InputField
                label="Contact email"
                value={contactEmail()}
                placeholder="ops@acmeprint.com"
                onInput={(event) => setContactEmail(event.currentTarget.value)}
              />
            </div>
            <SelectField
              label="Partner type"
              value={partnerType()}
              options={partnerTypeOptions}
              onChange={(event) => setPartnerType(event.currentTarget.value)}
            />
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Supported product types"
                value={supportedProductTypes()}
                placeholder="tshirt, hoodie, tote"
                onInput={(event) =>
                  setSupportedProductTypes(event.currentTarget.value)
                }
              />
              <InputField
                label="Supported regions"
                value={supportedRegions()}
                placeholder="us, eu, uk"
                onInput={(event) =>
                  setSupportedRegions(event.currentTarget.value)
                }
              />
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="SLA days"
                value={slaDays()}
                placeholder="3"
                onInput={(event) => setSlaDays(event.currentTarget.value)}
              />
              <InputField
                label="Routing priority"
                value={routingPriority()}
                placeholder="100"
                onInput={(event) =>
                  setRoutingPriority(event.currentTarget.value)
                }
              />
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Base fulfillment cost"
                value={baseFulfillmentCost()}
                placeholder="$8.50"
                onInput={(event) =>
                  setBaseFulfillmentCost(event.currentTarget.value)
                }
              />
              <InputField
                label="Shipping cost rules"
                value={shippingCostRules()}
                placeholder="us:$4.50, eu:$6.00"
                onInput={(event) =>
                  setShippingCostRules(event.currentTarget.value)
                }
              />
            </div>
            <InfoAlert>
              Shipping cost rules use `region:cost` pairs separated by commas.
              Example: `us:$4.50, eu:$6.00, sea:$7.25`.
            </InfoAlert>
            <TextareaField
              label="Notes"
              value={notes()}
              rows={4}
              onInput={(event) => setNotes(event.currentTarget.value)}
            />
            <div class="flex flex-wrap gap-3">
              <Button type="submit" loading={saving()}>
                {isEditing() ? 'Save changes' : 'Create partner'}
              </Button>
              <Show when={isEditing()}>
                <Button
                  type="button"
                  color="light"
                  onClick={() => {
                    resetForm();
                    setMessage('Partner edit canceled.');
                  }}
                >
                  Cancel edit
                </Button>
              </Show>
              <Button
                type="button"
                color="alternative"
                onClick={() => {
                  void loadPartners();
                }}
              >
                Reload partners
              </Button>
            </div>
          </form>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Partner list"
            subtitle="Active and inactive partners available to this store."
          />

          <div class="grid gap-4 md:grid-cols-2">
            <SelectField
              label="Filter by partner type"
              value={filterPartnerType()}
              options={partnerTypeOptions}
              onChange={(event) => setFilterPartnerType(event.currentTarget.value)}
            />
            <SelectField
              label="Filter by status"
              value={filterStatus()}
              options={partnerStatusOptions}
              onChange={(event) => setFilterStatus(event.currentTarget.value)}
            />
          </div>

          <div class="flex flex-wrap gap-3">
            <Button
              type="button"
              color="alternative"
              onClick={() => {
                void loadPartners();
              }}
            >
              Apply filters
            </Button>
            <Button
              type="button"
              color="light"
              onClick={() => {
                setFilterPartnerType('');
                setFilterStatus('');
                void loadPartners();
              }}
            >
              Reset filters
            </Button>
          </div>

          <Show when={loading()}>
            <LoadingInline label="Loading partners..." />
          </Show>

          <Show
            when={!loading() && partners().length > 0}
            fallback={
              <EmptyBlock
                title="No partners yet"
                copy="Add your first print or fulfillment partner to start shaping the execution side of this store."
              />
            }
          >
            <div class="space-y-3">
              <For each={partners()}>
                {(partner) => (
                  <div class="rounded-2xl border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">{partner.name}</p>
                        <p class="mt-1 text-sm text-gray-500">
                          {partner.code} · {partnerTypeLabel(partner.partnerType)}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          Priority {partner.routingPriority || 0} · SLA{' '}
                          {partner.slaDays || 0}d
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          Fulfillment {partner.baseFulfillmentCost || 'TBD'} ·
                          Shipping{' '}
                          {joinShippingCostRules(partner.shippingCostRules) ||
                            'No region rules'}
                        </p>
                        <Show when={partner.contactEmail || partner.contactName}>
                          <p class="mt-1 text-sm text-gray-500">
                            {partner.contactName || 'No contact name'} ·{' '}
                            {partner.contactEmail || 'No contact email'}
                          </p>
                        </Show>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge content={partner.status} color={badgeColorForStatus(partner.status)} />
                        <Button
                          size="xs"
                          color="alternative"
                          href={`/t/${params().tenantId}/partners/${partner.id}`}
                        >
                          View
                        </Button>
                        <Button
                          size="xs"
                          color="light"
                          onClick={() => {
                            startEdit(partner);
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          size="xs"
                          color={partner.status === 'active' ? 'alternative' : 'green'}
                          onClick={() => {
                            void toggleStatus(partner);
                          }}
                        >
                          {partner.status === 'active' ? 'Deactivate' : 'Activate'}
                        </Button>
                      </div>
                    </div>
                    <Show when={partner.notes}>
                      <p class="mt-3 text-sm text-gray-600">{partner.notes}</p>
                    </Show>
                    <p class="mt-3 text-sm text-gray-600">
                      Products: {joinCapabilityList(partner.supportedProductTypes || []) || 'Any'} ·
                      Regions: {joinCapabilityList(partner.supportedRegions || []) || 'Any'}
                    </p>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Card>
      </div>
    </PageShell>
  );
}
