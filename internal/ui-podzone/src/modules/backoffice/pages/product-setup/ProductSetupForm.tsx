import type { Accessor } from 'solid-js'
import { Button } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  FormCheckboxField,
  FormInputField,
  FormSelectField,
  FormTextareaField,
  type FormStore,
} from '@/solid/forms'
import type { ProductSetupFormValues } from './forms'
import { channelOptions, statusOptions } from './presentation'

type ProductSetupFormProps = {
  form: FormStore<ProductSetupFormValues>
  saving: Accessor<boolean>
  onSubmit: (event: SubmitEvent) => void
  onReset: () => void
}

export function ProductSetupForm(props: ProductSetupFormProps) {
  return (
    <>
      <SectionTitle
        title="New setup draft"
        subtitle="Capture the commercial and execution basics for a POD product before deeper automation exists."
      />
      <form class="space-y-4" onSubmit={props.onSubmit}>
        <FormInputField
          form={props.form}
          name="name"
          label="Product name"
          placeholder="Signature Tee"
        />
        <FormInputField
          form={props.form}
          name="partner"
          label="Preferred print partner"
          placeholder="Acme Print Lab"
        />
        <div class="grid gap-4 md:grid-cols-2">
          <FormInputField
            form={props.form}
            name="baseCost"
            label="Base cost"
            placeholder="$8.20"
          />
          <FormInputField
            form={props.form}
            name="retailPrice"
            label="Retail price"
            placeholder="$24.00"
          />
        </div>
        <FormSelectField
          form={props.form}
          name="status"
          label="Draft status"
          options={statusOptions}
        />
        <FormSelectField
          form={props.form}
          name="channel"
          label="Mock publish channel"
          options={channelOptions}
        />
        <div class="grid gap-4 md:grid-cols-2">
          <FormInputField
            form={props.form}
            name="variantColor"
            label="Primary color"
            placeholder="Black"
          />
          <FormInputField
            form={props.form}
            name="variantSize"
            label="Primary size"
            placeholder="M"
          />
        </div>
        <div class="rounded-lg border border-gray-200 p-4">
          <p class="text-sm font-semibold text-gray-900">Artwork readiness</p>
          <div class="mt-3 grid gap-3 md:grid-cols-2">
            <FormCheckboxField
              form={props.form}
              name="hasFrontArtwork"
              label="Front artwork prepared"
            />
            <FormCheckboxField
              form={props.form}
              name="hasBackArtwork"
              label="Back artwork prepared"
            />
            <FormCheckboxField
              form={props.form}
              name="mockupReady"
              label="Mockups exported"
            />
            <FormCheckboxField
              form={props.form}
              name="printSpecChecked"
              label="Print specs checked"
            />
          </div>
        </div>
        <FormTextareaField
          form={props.form}
          name="notes"
          label="Setup notes"
          rows={4}
        />
        <div class="flex flex-wrap gap-3">
          <Button type="submit" loading={props.saving()}>
            Save setup draft
          </Button>
          <Button type="button" color="alternative" onClick={props.onReset}>
            Clear form
          </Button>
        </div>
      </form>
    </>
  )
}
