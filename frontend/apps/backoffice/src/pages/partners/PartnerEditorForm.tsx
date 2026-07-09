import { Show, type Accessor } from 'solid-js'
import { InfoAlert } from '@podzone/shared/ui/components/common/Feedback'
import { Button } from '@podzone/shared/ui/components/common/Primitives'
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { FormInputField, FormSelectField, FormTextareaField, type FormStore } from '@podzone/shared/ui/forms'
import type { PartnerFormValues } from './forms'
import { partnerTypeOptions } from './presentation'

type PartnerEditorFormProps = {
    form: FormStore<PartnerFormValues>
    isEditing: Accessor<boolean>
    onSubmit: (event: SubmitEvent) => void
    onCancel: () => void
    onReload: () => void
}

export function PartnerEditorForm(props: PartnerEditorFormProps) {
    return (
        <>
            <SectionTitle
                title={props.isEditing() ? 'Edit print partner' : 'Add print partner'}
                subtitle={
                    props.isEditing()
                        ? 'Update partner details without leaving the store workspace.'
                        : 'Create a partner record for production, fulfillment, or future sourced-product workflows.'
                }
            />
            <form class="space-y-4" onSubmit={props.onSubmit}>
                <FormInputField form={props.form} name="name" label="Partner name" placeholder="Acme Print Lab" />
                <FormInputField
                    form={props.form}
                    name="code"
                    label="Partner code"
                    placeholder="acme-print"
                    disabled={props.isEditing()}
                />
                <Show when={props.isEditing()}>
                    <InfoAlert>Partner code is locked during edit so external references stay stable.</InfoAlert>
                </Show>
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField form={props.form} name="contactName" label="Contact name" placeholder="Linh Tran" />
                    <FormInputField
                        form={props.form}
                        name="contactEmail"
                        label="Contact email"
                        placeholder="ops@acmeprint.com"
                    />
                </div>
                <FormSelectField
                    form={props.form}
                    name="partnerType"
                    label="Partner type"
                    options={partnerTypeOptions}
                />
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField
                        form={props.form}
                        name="supportedProductTypes"
                        label="Supported product types"
                        placeholder="tshirt, hoodie, tote"
                    />
                    <FormInputField
                        form={props.form}
                        name="supportedRegions"
                        label="Supported regions"
                        placeholder="us, eu, uk"
                    />
                </div>
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField form={props.form} name="slaDays" label="SLA days" placeholder="3" />
                    <FormInputField
                        form={props.form}
                        name="routingPriority"
                        label="Routing priority"
                        placeholder="100"
                    />
                </div>
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField
                        form={props.form}
                        name="baseFulfillmentCost"
                        label="Base fulfillment cost"
                        placeholder="$8.50"
                    />
                    <FormInputField
                        form={props.form}
                        name="shippingCostRules"
                        label="Shipping cost rules"
                        placeholder="us:$4.50, eu:$6.00"
                    />
                </div>
                <InfoAlert>
                    Shipping cost rules use `region:cost` pairs separated by commas. Example: `us:$4.50, eu:$6.00,
                    sea:$7.25`.
                </InfoAlert>
                <FormTextareaField form={props.form} name="notes" label="Notes" rows={4} />
                <div class="flex flex-wrap gap-3">
                    <Button type="submit" loading={props.form.isSubmitting()}>
                        {props.isEditing() ? 'Save changes' : 'Create partner'}
                    </Button>
                    <Show when={props.isEditing()}>
                        <Button type="button" color="light" onClick={props.onCancel}>
                            Cancel edit
                        </Button>
                    </Show>
                    <Button type="button" color="alternative" onClick={props.onReload}>
                        Reload partners
                    </Button>
                </div>
            </form>
        </>
    )
}
