import { Show } from 'solid-js'
import { IamStatementBuilder } from '@/solid/components/common/IamStatementBuilder'
import {
  Badge,
  Button,
  SelectField,
} from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  FormInputField,
  FormSelectField,
  createFormStore,
  jsonArray,
  required,
} from '@/solid/forms'
import type {
  CreatePolicyFormValues,
  CreatePolicyVersionFormValues,
} from './policy-forms'
import { useAdminIamPolicy } from './policy-context'
import { PolicyDetailTables } from './PolicyDetailTables'

export function PoliciesPanel() {
  const policy = useAdminIamPolicy()
  const createPolicyForm = createFormStore<CreatePolicyFormValues>({
    initialValues: {
      scope: policy.policyScope(),
      name: policy.policyName(),
      description: policy.policyDescription(),
      statementsJson: policy.policyStatementsJson(),
    },
    validators: {
      scope: [required('Choose a policy scope.')],
      name: [required('Enter a policy name.')],
      statementsJson: [jsonArray('Policy statements must be a JSON array.')],
    },
  })
  const versionForm = createFormStore<CreatePolicyVersionFormValues>({
    initialValues: {
      statementsJson: policy.policyVersionJson(),
    },
    validators: {
      statementsJson: [
        jsonArray('Policy version statements must be a JSON array.'),
      ],
    },
  })

  const submitCreatePolicy = async (event: SubmitEvent) => {
    event.preventDefault()
    if (!createPolicyForm.validate()) {
      return
    }
    createPolicyForm.setSubmitting(true)
    await policy.createPolicyFromForm({ ...createPolicyForm.values })
    createPolicyForm.setSubmitting(false)
  }
  const submitCreatePolicyVersion = async () => {
    if (!versionForm.validate()) {
      return
    }
    versionForm.setSubmitting(true)
    await policy.createPolicyVersionFromForm({ ...versionForm.values })
    versionForm.setSubmitting(false)
  }

  return (
    <>
      <SectionTitle
        title="Policies and versions"
        subtitle="Create managed policies, inspect attachments, and roll default versions."
      />
      <form class="space-y-3" onSubmit={submitCreatePolicy}>
        <div class="grid gap-3 md:grid-cols-2">
          <FormSelectField
            form={createPolicyForm}
            name="scope"
            label="Policy scope"
            options={policy.policyScopeOptions}
          />
          <FormInputField
            form={createPolicyForm}
            name="name"
            label="Policy name"
          />
        </div>
        <FormInputField
          form={createPolicyForm}
          name="description"
          label="Description"
        />
        <IamStatementBuilder
          label="Statements"
          value={createPolicyForm.values.statementsJson}
          onChange={(value) =>
            createPolicyForm.setValue('statementsJson', value)
          }
        />
        <Show when={createPolicyForm.hasError('statementsJson')}>
          <p class="text-xs font-medium text-red-600">
            {createPolicyForm.error('statementsJson')}
          </p>
        </Show>
        <Button
          type="submit"
          size="sm"
          loading={createPolicyForm.isSubmitting()}
        >
          Create policy
        </Button>
      </form>

      <Show when={policy.policyOptions().length > 0}>
        <SelectField
          label="Inspect policy"
          value={policy.selectedPolicyName()}
          options={policy.policyOptions()}
          onChange={(e) => policy.setSelectedPolicyName(e.currentTarget.value)}
        />
      </Show>

      <Show when={policy.policyDetail()}>
        {(detail) => (
          <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600">
            <p class="font-semibold text-gray-900">{detail().name}</p>
            <p class="mt-1">{detail().description || 'No description'}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <Badge content={detail().scope} color="blue" />
              <Badge
                content={`default ${detail().defaultVersion || 'v1'}`}
                color="green"
              />
              <Show when={detail().isSystem}>
                <Badge content="system" color="dark" />
              </Show>
            </div>
          </div>
        )}
      </Show>

      <IamStatementBuilder
        label="New version statements"
        value={versionForm.values.statementsJson}
        onChange={(value) => versionForm.setValue('statementsJson', value)}
      />
      <Show when={versionForm.hasError('statementsJson')}>
        <p class="text-xs font-medium text-red-600">
          {versionForm.error('statementsJson')}
        </p>
      </Show>
      <div class="flex flex-wrap gap-3">
        <Button
          size="sm"
          color="dark"
          onClick={submitCreatePolicyVersion}
          disabled={!policy.selectedPolicyName()}
          loading={versionForm.isSubmitting()}
        >
          Create version
        </Button>
        <Button
          size="sm"
          color="red"
          onClick={policy.handleDeletePolicy}
          disabled={!policy.selectedPolicyName()}
        >
          Delete policy
        </Button>
      </div>

      <PolicyDetailTables />
    </>
  )
}
