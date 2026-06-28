import { useParams } from '@tanstack/solid-router'
import { Show, createEffect, createSignal } from 'solid-js'
import {
  createPartner,
  listPartners,
  updatePartner,
  updatePartnerStatus,
  type PartnerInfo,
} from '@/services/partner'
import { tenantStorage } from '@/services/tenantStorage'
import {
  ErrorAlert,
  InfoAlert,
  EmptyBlock,
} from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Button, Card, SelectField } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { createFormStore, email, required } from '@/solid/forms'
import { useTenantWorkspace } from '@/solid/workspace/context'
import {
  nonNegativeInteger,
  partnerInitialValues,
  shippingRules,
} from './partners/forms'
import { PartnerEditorForm } from './partners/PartnerEditorForm'
import { PartnerTable } from './partners/PartnerTable'
import {
  joinCapabilityList,
  joinShippingCostRules,
  parseCapabilityList,
  parseShippingCostRules,
  partnerStatusOptions,
  partnerTypeOptions,
} from './partners/presentation'

export default function TenantPartnersPage() {
  const params = useParams({ from: '/t/$tenantId/partners' })
  const workspace = useTenantWorkspace()

  const [partners, setPartners] = createSignal<PartnerInfo[]>([])
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')
  const [message, setMessage] = createSignal('')

  const partnerForm = createFormStore({
    initialValues: partnerInitialValues,
    validators: {
      name: [required('Enter a partner name.')],
      code: [required('Enter a stable partner code.')],
      contactEmail: [email('Enter a valid contact email.')],
      slaDays: [nonNegativeInteger('SLA days must be a whole number.')],
      routingPriority: [
        nonNegativeInteger('Routing priority must be a whole number.'),
      ],
      shippingCostRules: [
        shippingRules('Use comma-separated region:cost pairs.'),
      ],
    },
  })
  const [editingPartnerId, setEditingPartnerId] = createSignal('')
  const [filterPartnerType, setFilterPartnerType] = createSignal('')
  const [filterStatus, setFilterStatus] = createSignal('')
  const currentStoreId = () => workspace?.currentStoreId() || ''
  const currentStore = () => workspace?.currentStore()
  const workspaceReady = () => !workspace || currentStoreId().trim().length > 0
  const storeLabel = () =>
    currentStore()?.name || currentStoreId() || 'selected store'

  const isEditing = () => editingPartnerId().trim().length > 0

  const resetForm = () => {
    setEditingPartnerId('')
    partnerForm.reset()
  }

  const loadPartners = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await listPartners(
        params().tenantId,
        filterPartnerType(),
        filterStatus()
      )
      if (!result.success) {
        setError(result.message)
        setPartners([])
        return
      }
      setPartners(result.data)
    } finally {
      setLoading(false)
    }
  }

  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (!partnerForm.validate()) {
      return
    }
    partnerForm.setSubmitting(true)
    setError('')
    setMessage('')
    try {
      const result = isEditing()
        ? await updatePartner({
            id: editingPartnerId(),
            name: partnerForm.values.name.trim(),
            contactName: partnerForm.values.contactName.trim(),
            contactEmail: partnerForm.values.contactEmail.trim(),
            notes: partnerForm.values.notes.trim(),
            partnerType: partnerForm.values.partnerType,
            supportedProductTypes: parseCapabilityList(
              partnerForm.values.supportedProductTypes
            ),
            supportedRegions: parseCapabilityList(
              partnerForm.values.supportedRegions
            ),
            slaDays: Number.parseInt(partnerForm.values.slaDays, 10),
            routingPriority: Number.parseInt(
              partnerForm.values.routingPriority,
              10
            ),
            baseFulfillmentCost: partnerForm.values.baseFulfillmentCost.trim(),
            shippingCostRules: parseShippingCostRules(
              partnerForm.values.shippingCostRules
            ),
          })
        : await createPartner({
            tenantId: params().tenantId,
            code: partnerForm.values.code.trim(),
            name: partnerForm.values.name.trim(),
            contactName: partnerForm.values.contactName.trim(),
            contactEmail: partnerForm.values.contactEmail.trim(),
            notes: partnerForm.values.notes.trim(),
            partnerType: partnerForm.values.partnerType,
            supportedProductTypes: parseCapabilityList(
              partnerForm.values.supportedProductTypes
            ),
            supportedRegions: parseCapabilityList(
              partnerForm.values.supportedRegions
            ),
            slaDays: Number.parseInt(partnerForm.values.slaDays, 10),
            routingPriority: Number.parseInt(
              partnerForm.values.routingPriority,
              10
            ),
            baseFulfillmentCost: partnerForm.values.baseFulfillmentCost.trim(),
            shippingCostRules: parseShippingCostRules(
              partnerForm.values.shippingCostRules
            ),
          })
      if (!result.success) {
        setError(result.message)
        return
      }
      const actionLabel = isEditing() ? 'Updated' : 'Created'
      resetForm()
      setMessage(`${actionLabel} print partner ${result.data.name}.`)
      await loadPartners()
    } finally {
      partnerForm.setSubmitting(false)
    }
  }

  const toggleStatus = async (partner: PartnerInfo) => {
    setError('')
    setMessage('')
    const nextStatus = partner.status === 'active' ? 'inactive' : 'active'
    const result = await updatePartnerStatus(partner.id, nextStatus)
    if (!result.success) {
      setError(result.message)
      return
    }
    setMessage(
      `${result.data.name} is now ${result.data.status === 'active' ? 'active' : 'inactive'}.`
    )
    await loadPartners()
  }

  const startEdit = (partner: PartnerInfo) => {
    setEditingPartnerId(partner.id)
    partnerForm.reset({
      name: partner.name,
      code: partner.code,
      contactName: partner.contactName || '',
      contactEmail: partner.contactEmail || '',
      notes: partner.notes || '',
      partnerType: partner.partnerType || 'print_on_demand',
      supportedProductTypes: joinCapabilityList(
        partner.supportedProductTypes || []
      ),
      supportedRegions: joinCapabilityList(partner.supportedRegions || []),
      slaDays: String(partner.slaDays || 0),
      routingPriority: String(partner.routingPriority || 0),
      baseFulfillmentCost: partner.baseFulfillmentCost || '',
      shippingCostRules: joinShippingCostRules(partner.shippingCostRules),
    })
    setError('')
    setMessage(`Editing print partner ${partner.name}.`)
  }

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId)
    if (!workspaceReady()) {
      return
    }
    void loadPartners()
  })

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Print Partners"
          title={`Execution partners for ${storeLabel()}`}
          copy="Manage partner records from the current tenant workspace while selecting the store context that will use them for production and fulfillment workflows."
        />
      </Card>

      <Show when={error()}>
        <ErrorAlert>{error()}</ErrorAlert>
      </Show>

      <Show when={message()}>
        <InfoAlert>{message()}</InfoAlert>
      </Show>

      <InfoAlert>
        This is the real partner record layer behind the demo flow. The store
        switcher keeps the active workspace explicit even while partner records
        are still tenant-owned underneath.
      </InfoAlert>

      <Show when={!workspaceReady()}>
        <EmptyBlock
          title="Choose a store first"
          copy="Use the workspace store switcher before opening partner operations from the seller shell."
        />
      </Show>

      <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <Card class="space-y-4">
          <PartnerEditorForm
            form={partnerForm}
            isEditing={isEditing}
            onSubmit={submit}
            onCancel={() => {
              resetForm()
              setMessage('Partner edit canceled.')
            }}
            onReload={() => {
              void loadPartners()
            }}
          />
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
              onChange={(event) =>
                setFilterPartnerType(event.currentTarget.value)
              }
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
                void loadPartners()
              }}
            >
              Apply filters
            </Button>
            <Button
              type="button"
              color="light"
              onClick={() => {
                setFilterPartnerType('')
                setFilterStatus('')
                void loadPartners()
              }}
            >
              Reset filters
            </Button>
          </div>

          <PartnerTable
            tenantID={params().tenantId}
            partners={partners}
            loading={loading}
            onEdit={startEdit}
            onToggleStatus={(partner) => {
              void toggleStatus(partner)
            }}
          />
        </Card>
      </div>
    </PageShell>
  )
}
