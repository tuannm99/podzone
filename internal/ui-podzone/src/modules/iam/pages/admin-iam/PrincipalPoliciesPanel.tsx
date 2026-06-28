import { Show } from 'solid-js'
import { IamStatementBuilder } from '@/solid/components/common/IamStatementBuilder'
import {
  Badge,
  Button,
  InputField,
  SelectField,
} from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  FormInputField,
  createFormStore,
  jsonArray,
  required,
} from '@/solid/forms'
import type {
  PrincipalBoundaryFormValues,
  PrincipalInlinePolicyFormValues,
  PrincipalManagedPolicyFormValues,
} from './principal-forms'
import { useAdminIamPrincipal, type PrincipalMode } from './principal-context'
import {
  PrincipalInlinePoliciesTable,
  PrincipalManagedPoliciesTable,
} from './PrincipalPolicyTables'

export function PrincipalPoliciesPanel() {
  const principal = useAdminIamPrincipal()
  const managedPolicyForm = createFormStore<PrincipalManagedPolicyFormValues>({
    initialValues: {
      policyName: principal.principalManagedPolicyName(),
    },
    validators: {
      policyName: [required('Enter a managed policy name.')],
    },
  })
  const boundaryForm = createFormStore<PrincipalBoundaryFormValues>({
    initialValues: {
      policyName: principal.principalBoundaryPolicyName(),
    },
    validators: {
      policyName: [required('Enter a boundary policy name.')],
    },
  })
  const inlinePolicyForm = createFormStore<PrincipalInlinePolicyFormValues>({
    initialValues: {
      name: principal.principalInlinePolicyName(),
      description: principal.principalInlinePolicyDescription(),
      statementsJson: principal.principalInlinePolicyJson(),
    },
    validators: {
      name: [required('Enter an inline policy name.')],
      statementsJson: [
        jsonArray('Inline policy statements must be a JSON array.'),
      ],
    },
  })

  const attachManagedPolicy = async () => {
    if (!managedPolicyForm.validate()) return
    managedPolicyForm.setSubmitting(true)
    await principal.attachPrincipalManagedPolicyFromForm({
      ...managedPolicyForm.values,
    })
    managedPolicyForm.setSubmitting(false)
  }

  const saveBoundary = async () => {
    if (!boundaryForm.validate()) return
    boundaryForm.setSubmitting(true)
    await principal.savePrincipalBoundaryFromForm({ ...boundaryForm.values })
    boundaryForm.setSubmitting(false)
  }

  const saveInlinePolicy = async () => {
    if (!inlinePolicyForm.validate()) return
    inlinePolicyForm.setSubmitting(true)
    await principal.savePrincipalInlinePolicyFromForm({
      ...inlinePolicyForm.values,
    })
    inlinePolicyForm.setSubmitting(false)
  }

  return (
    <>
      <SectionTitle
        title="Principal policies"
        subtitle="Manage direct policies, inline policies, and permission boundaries for platform and tenant users."
      />
      <div class="grid gap-3 md:grid-cols-2">
        <SelectField
          label="Principal mode"
          value={principal.principalMode()}
          options={[
            { name: 'Platform user', value: 'platform' },
            { name: 'Tenant user', value: 'tenant' },
          ]}
          onChange={(e) =>
            principal.setPrincipalMode(e.currentTarget.value as PrincipalMode)
          }
        />
        <Show
          when={principal.principalMode() === 'platform'}
          fallback={
            <SelectField
              label="Tenant"
              value={principal.principalTenantId()}
              options={principal.tenantOptions()}
              onChange={(e) =>
                principal.setPrincipalTenantId(e.currentTarget.value)
              }
            />
          }
        >
          <InputField
            label="Platform user id"
            value={principal.principalPlatformUserId()}
            onInput={(e) =>
              principal.setPrincipalPlatformUserId(e.currentTarget.value)
            }
          />
        </Show>
      </div>
      <Show when={principal.principalMode() === 'tenant'}>
        <InputField
          label="Tenant user id"
          value={principal.principalTenantUserId()}
          onInput={(e) =>
            principal.setPrincipalTenantUserId(e.currentTarget.value)
          }
        />
      </Show>

      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
        <FormInputField
          form={managedPolicyForm}
          name="policyName"
          label="Managed policy name"
        />
        <Button
          size="sm"
          onClick={attachManagedPolicy}
          disabled={!managedPolicyForm.values.policyName.trim()}
          loading={managedPolicyForm.isSubmitting()}
        >
          Attach policy
        </Button>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Managed policies</p>
          <p class="text-xs text-gray-500">
            Direct attachments only affect this principal and stack with group,
            role, boundary, and SCP evaluation.
          </p>
          <PrincipalManagedPoliciesTable />
        </div>
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Permission boundary</p>
          <p class="text-xs text-gray-500">
            Boundary policies cap the maximum access this principal can
            exercise, even when identity policies allow more.
          </p>
          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-end">
            <FormInputField
              form={boundaryForm}
              name="policyName"
              label="Boundary policy"
            />
            <Button
              size="sm"
              color="dark"
              onClick={saveBoundary}
              disabled={!boundaryForm.values.policyName.trim()}
              loading={boundaryForm.isSubmitting()}
            >
              Save
            </Button>
            <Button
              size="sm"
              color="red"
              onClick={principal.handleDeletePrincipalBoundary}
            >
              Delete
            </Button>
          </div>
          <Show when={principal.currentBoundary()}>
            {(boundary) => (
              <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600">
                <div class="flex flex-wrap gap-2">
                  <Badge content="boundary" color="pink" />
                  <Badge content={boundary().policyName} color="blue" />
                </div>
              </div>
            )}
          </Show>
        </div>
      </div>

      <div class="space-y-3">
        <div class="grid gap-3 md:grid-cols-2">
          <FormInputField
            form={inlinePolicyForm}
            name="name"
            label="Inline policy name"
          />
          <FormInputField
            form={inlinePolicyForm}
            name="description"
            label="Inline policy description"
          />
        </div>
        <IamStatementBuilder
          label="Inline policy statements"
          value={inlinePolicyForm.values.statementsJson}
          onChange={(value) =>
            inlinePolicyForm.setValue('statementsJson', value)
          }
        />
        <Show when={inlinePolicyForm.hasError('statementsJson')}>
          <p class="text-xs font-medium text-red-600">
            {inlinePolicyForm.error('statementsJson')}
          </p>
        </Show>
        <Button
          size="sm"
          color="dark"
          onClick={saveInlinePolicy}
          disabled={!inlinePolicyForm.values.name.trim()}
          loading={inlinePolicyForm.isSubmitting()}
        >
          Save inline policy
        </Button>
        <PrincipalInlinePoliciesTable />
      </div>
    </>
  )
}
