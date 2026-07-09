import { useParams } from '@tanstack/solid-router'
import { Show } from 'solid-js'
import { EmptyBlock, ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { PageShell } from '@/solid/components/common/PageShell'
import { Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives'
import { SectionLead } from '@/solid/components/common/SectionLead'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useTenantWorkspace } from '@/modules/shell/workspace/context'
import { PartnerEditorForm } from './PartnerEditorForm'
import { PartnerTable } from './PartnerTable'
import { createPartnersViewModel } from './createPartnersViewModel'
import { partnerStatusOptions, partnerTypeOptions } from './presentation'

export function TenantPartnersView() {
    const params = useParams({ from: '/t/$tenantId/partners' })
    const workspace = useTenantWorkspace()
    const currentStoreID = () => workspace?.currentStoreId() || ''
    const currentStore = () => workspace?.currentStore()
    const workspaceReady = () => !workspace || currentStoreID().trim().length > 0
    const storeLabel = () => currentStore()?.name || currentStoreID() || 'selected store'
    const partners = createPartnersViewModel({
        tenantID: () => params().tenantId,
        storeID: currentStoreID,
        workspaceReady,
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

            <Show when={partners.error()}>
                <ErrorAlert>{partners.error()}</ErrorAlert>
            </Show>
            <Show when={partners.message()}>
                <InfoAlert>{partners.message()}</InfoAlert>
            </Show>
            <InfoAlert>
                This is the real partner record layer behind the demo flow. The store switcher keeps the active
                workspace explicit even while partner records are still tenant-owned underneath.
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
                        form={partners.form}
                        isEditing={partners.isEditing}
                        onSubmit={partners.submit}
                        onCancel={partners.cancelEdit}
                        onReload={() => void partners.reload()}
                    />
                </Card>

                <Card class="space-y-4">
                    <SectionTitle
                        title="Partner list"
                        subtitle="Active and inactive partners available to this store."
                    />
                    <div class="flex items-end gap-3">
                        <InputField
                            label="Search partners"
                            value={partners.search()}
                            placeholder="Name, code, email, type, or status"
                            onInput={(event) => partners.setSearch(event.currentTarget.value)}
                        />
                        <Button type="button" color="alternative" onClick={partners.applySearch}>
                            Search
                        </Button>
                    </div>
                    <div class="grid gap-4 md:grid-cols-2">
                        <SelectField
                            label="Filter by partner type"
                            value={partners.filterPartnerType()}
                            options={partnerTypeOptions}
                            onChange={(event) => partners.setFilterPartnerType(event.currentTarget.value)}
                        />
                        <SelectField
                            label="Filter by status"
                            value={partners.filterStatus()}
                            options={partnerStatusOptions}
                            onChange={(event) => partners.setFilterStatus(event.currentTarget.value)}
                        />
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <Button type="button" color="alternative" onClick={partners.applyFilters}>
                            Apply filters
                        </Button>
                        <Button type="button" color="light" onClick={partners.resetFilters}>
                            Reset filters
                        </Button>
                    </div>
                    <PartnerTable
                        tenantID={params().tenantId}
                        partners={partners.partners}
                        pageInfo={partners.pageInfo}
                        page={partners.page}
                        loading={partners.loading}
                        statusChangingPartnerID={partners.statusChangingPartnerID}
                        onPageChange={partners.setPage}
                        onEdit={partners.edit}
                        onToggleStatus={(partner) => void partners.toggleStatus(partner)}
                    />
                </Card>
            </div>
        </PageShell>
    )
}
