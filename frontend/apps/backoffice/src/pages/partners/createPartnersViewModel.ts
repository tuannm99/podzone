import { useAuthContext } from '@podzone/shared/auth'
import { createEffect, createSignal, type Accessor } from 'solid-js'
import { createPartner, listPartners, updatePartner, updatePartnerStatus, type PartnerInfo } from '@podzone/shared/services/partner'

import { createFormStore, email, required } from '@podzone/shared/ui/forms'
import { createPaginatedResource } from '@podzone/shared/ui/pagination'
import { nonNegativeInteger, partnerInitialValues, shippingRules } from './forms'
import { joinCapabilityList, joinShippingCostRules, parseCapabilityList, parseShippingCostRules } from './presentation'

type PartnersViewModelParams = {
    tenantID: Accessor<string>
    storeID: Accessor<string>
    workspaceReady: Accessor<boolean>
}

export function createPartnersViewModel(params: PartnersViewModelParams) {
    const auth = useAuthContext()
    const [mutationError, setMutationError] = createSignal('')
    const [message, setMessage] = createSignal('')
    const [editingPartnerID, setEditingPartnerID] = createSignal('')
    const [statusChangingPartnerID, setStatusChangingPartnerID] = createSignal('')
    const [filterPartnerType, setFilterPartnerType] = createSignal('')
    const [filterStatus, setFilterStatus] = createSignal('')
    const [search, setSearch] = createSignal('')
    const form = createFormStore({
        initialValues: partnerInitialValues,
        validators: {
            name: [required('Enter a partner name.')],
            code: [required('Enter a stable partner code.')],
            contactEmail: [email('Enter a valid contact email.')],
            slaDays: [nonNegativeInteger('SLA days must be a whole number.')],
            routingPriority: [nonNegativeInteger('Routing priority must be a whole number.')],
            shippingCostRules: [shippingRules('Use comma-separated region:cost pairs.')],
        },
    })
    const isEditing = () => Boolean(editingPartnerID().trim())
    const list = createPaginatedResource<PartnerInfo>(
        {
            page: 1,
            pageSize: 8,
            sortBy: 'routingPriority',
            sortDirection: 'SORT_DIRECTION_DESC',
        },
        async (query) => {
            const result = await listPartners(params.tenantID(), query)
            if (!result.success) throw new Error(result.message)
            return result.data
        },
        {
            enabled: params.workspaceReady,
            dependency: () => `${params.tenantID()}|${params.storeID()}`,
        }
    )
    const partners = list.items
    const loading = list.loading
    const error = () => mutationError() || list.error()
    const reload = list.reload

    const resetForm = () => {
        setEditingPartnerID('')
        form.reset()
    }

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        form.setSubmitting(true)
        setMutationError('')
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
                      supportedProductTypes: parseCapabilityList(form.values.supportedProductTypes),
                      supportedRegions: parseCapabilityList(form.values.supportedRegions),
                      slaDays: Number.parseInt(form.values.slaDays, 10),
                      routingPriority: Number.parseInt(form.values.routingPriority, 10),
                      baseFulfillmentCost: form.values.baseFulfillmentCost.trim(),
                      shippingCostRules: parseShippingCostRules(form.values.shippingCostRules),
                  })
                : await createPartner({
                      tenantId: params.tenantID(),
                      code: form.values.code.trim(),
                      name: form.values.name.trim(),
                      contactName: form.values.contactName.trim(),
                      contactEmail: form.values.contactEmail.trim(),
                      notes: form.values.notes.trim(),
                      partnerType: form.values.partnerType,
                      supportedProductTypes: parseCapabilityList(form.values.supportedProductTypes),
                      supportedRegions: parseCapabilityList(form.values.supportedRegions),
                      slaDays: Number.parseInt(form.values.slaDays, 10),
                      routingPriority: Number.parseInt(form.values.routingPriority, 10),
                      baseFulfillmentCost: form.values.baseFulfillmentCost.trim(),
                      shippingCostRules: parseShippingCostRules(form.values.shippingCostRules),
                  })
            if (!result.success) {
                setMutationError(result.message)
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
        if (statusChangingPartnerID()) return
        setStatusChangingPartnerID(partner.id)
        setMutationError('')
        setMessage('')
        try {
            const nextStatus = partner.status === 'active' ? 'inactive' : 'active'
            const result = await updatePartnerStatus(partner.id, nextStatus)
            if (!result.success) {
                setMutationError(result.message)
                return
            }
            setMessage(`${result.data.name} is now ${result.data.status === 'active' ? 'active' : 'inactive'}.`)
            await reload()
        } finally {
            setStatusChangingPartnerID('')
        }
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
            supportedProductTypes: joinCapabilityList(partner.supportedProductTypes || []),
            supportedRegions: joinCapabilityList(partner.supportedRegions || []),
            slaDays: String(partner.slaDays || 0),
            routingPriority: String(partner.routingPriority || 0),
            baseFulfillmentCost: partner.baseFulfillmentCost || '',
            shippingCostRules: joinShippingCostRules(partner.shippingCostRules),
        })
        setMutationError('')
        setMessage(`Editing print partner ${partner.name}.`)
    }

    const applyFilters = () => {
        const partnerType = filterPartnerType()
        const status = filterStatus()
        list.updateQuery({
            filters: [
                ...(partnerType
                    ? [
                          {
                              field: 'partnerType',
                              operator: 'FILTER_OPERATOR_EQ' as const,
                              values: [partnerType],
                          },
                      ]
                    : []),
                ...(status
                    ? [
                          {
                              field: 'status',
                              operator: 'FILTER_OPERATOR_EQ' as const,
                              values: [status],
                          },
                      ]
                    : []),
            ],
        })
    }

    const resetFilters = () => {
        setFilterPartnerType('')
        setFilterStatus('')
        list.updateQuery({ filters: [] })
    }

    const applySearch = () => list.updateQuery({ search: search().trim() })

    createEffect(() => {
        auth.setActiveTenantId(params.tenantID())
    })

    return {
        partners,
        loading,
        pageInfo: list.pageInfo,
        page: () => list.query.page,
        setPage: (page: number) => list.updateQuery({ page }),
        error,
        message,
        form,
        isEditing,
        filterPartnerType,
        setFilterPartnerType,
        filterStatus,
        setFilterStatus,
        statusChangingPartnerID,
        search,
        setSearch,
        applySearch,
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
