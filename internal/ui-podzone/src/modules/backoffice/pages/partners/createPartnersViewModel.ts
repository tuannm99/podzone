import { createEffect, createSignal, on, type Accessor } from 'solid-js'
import {
  createPartner,
  listPartners,
  updatePartner,
  updatePartnerStatus,
  type PartnerInfo,
} from '@/services/partner'
import { tenantStorage } from '@/services/tenantStorage'
import { createFormStore, email, required } from '@/solid/forms'
import {
  nonNegativeInteger,
  partnerInitialValues,
  shippingRules,
} from './forms'
import {
  joinCapabilityList,
  joinShippingCostRules,
  parseCapabilityList,
  parseShippingCostRules,
} from './presentation'

type PartnersViewModelParams = {
  tenantID: Accessor<string>
  storeID: Accessor<string>
  workspaceReady: Accessor<boolean>
}

export function createPartnersViewModel(params: PartnersViewModelParams) {
  const [partners, setPartners] = createSignal<PartnerInfo[]>([])
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')
  const [message, setMessage] = createSignal('')
  const [editingPartnerID, setEditingPartnerID] = createSignal('')
  const [filterPartnerType, setFilterPartnerType] = createSignal('')
  const [filterStatus, setFilterStatus] = createSignal('')
  const [appliedPartnerType, setAppliedPartnerType] = createSignal('')
  const [appliedStatus, setAppliedStatus] = createSignal('')
  const form = createFormStore({
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
  const isEditing = () => Boolean(editingPartnerID().trim())
  let requestVersion = 0

  const reload = async () => {
    const currentRequest = ++requestVersion
    setLoading(true)
    setError('')
    try {
      const result = await listPartners(
        params.tenantID(),
        appliedPartnerType(),
        appliedStatus()
      )
      if (currentRequest !== requestVersion) return
      if (!result.success) {
        setError(result.message)
        setPartners([])
        return
      }
      setPartners(result.data)
    } finally {
      if (currentRequest === requestVersion) setLoading(false)
    }
  }

  const resetForm = () => {
    setEditingPartnerID('')
    form.reset()
  }

  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (!form.validate()) return
    form.setSubmitting(true)
    setError('')
    setMessage('')
    try {
      const result = isEditing()
        ? await updatePartner({
            id: editingPartnerID(),
            name: form.values.name.trim(),
            contactName: form.values.contactName.trim(),
            contactEmail: form.values.contactEmail.trim(),
            notes: form.values.notes.trim(),
            partnerType: form.values.partnerType,
            supportedProductTypes: parseCapabilityList(
              form.values.supportedProductTypes
            ),
            supportedRegions: parseCapabilityList(form.values.supportedRegions),
            slaDays: Number.parseInt(form.values.slaDays, 10),
            routingPriority: Number.parseInt(form.values.routingPriority, 10),
            baseFulfillmentCost: form.values.baseFulfillmentCost.trim(),
            shippingCostRules: parseShippingCostRules(
              form.values.shippingCostRules
            ),
          })
        : await createPartner({
            tenantId: params.tenantID(),
            code: form.values.code.trim(),
            name: form.values.name.trim(),
            contactName: form.values.contactName.trim(),
            contactEmail: form.values.contactEmail.trim(),
            notes: form.values.notes.trim(),
            partnerType: form.values.partnerType,
            supportedProductTypes: parseCapabilityList(
              form.values.supportedProductTypes
            ),
            supportedRegions: parseCapabilityList(form.values.supportedRegions),
            slaDays: Number.parseInt(form.values.slaDays, 10),
            routingPriority: Number.parseInt(form.values.routingPriority, 10),
            baseFulfillmentCost: form.values.baseFulfillmentCost.trim(),
            shippingCostRules: parseShippingCostRules(
              form.values.shippingCostRules
            ),
          })
      if (!result.success) {
        setError(result.message)
        return
      }
      const actionLabel = isEditing() ? 'Updated' : 'Created'
      resetForm()
      setMessage(`${actionLabel} print partner ${result.data.name}.`)
      await reload()
    } finally {
      form.setSubmitting(false)
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
    await reload()
  }

  const edit = (partner: PartnerInfo) => {
    setEditingPartnerID(partner.id)
    form.reset({
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

  const applyFilters = () => {
    const partnerType = filterPartnerType()
    const status = filterStatus()
    if (partnerType === appliedPartnerType() && status === appliedStatus()) {
      void reload()
      return
    }
    setAppliedPartnerType(partnerType)
    setAppliedStatus(status)
  }

  const resetFilters = () => {
    setFilterPartnerType('')
    setFilterStatus('')
    if (!appliedPartnerType() && !appliedStatus()) {
      void reload()
      return
    }
    setAppliedPartnerType('')
    setAppliedStatus('')
  }

  createEffect(
    on(
      () =>
        [
          params.tenantID(),
          params.storeID(),
          params.workspaceReady(),
          appliedPartnerType(),
          appliedStatus(),
        ] as const,
      ([tenantID, , ready]) => {
        tenantStorage.setTenantID(tenantID)
        if (ready) void reload()
      }
    )
  )

  return {
    partners,
    loading,
    error,
    message,
    form,
    isEditing,
    filterPartnerType,
    setFilterPartnerType,
    filterStatus,
    setFilterStatus,
    reload,
    submit,
    toggleStatus,
    edit,
    cancelEdit: () => {
      resetForm()
      setMessage('Partner edit canceled.')
    },
    applyFilters,
    resetFilters,
  }
}

export type PartnersViewModel = ReturnType<typeof createPartnersViewModel>
