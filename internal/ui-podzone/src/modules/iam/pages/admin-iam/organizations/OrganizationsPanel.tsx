import { Show, type Accessor, type Setter } from 'solid-js'
import {
  Button,
  InputField,
  SelectField,
  type SelectOption,
} from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import {
  OrganizationsCollection,
  type OrganizationsCollectionProps,
} from './OrganizationsCollection'

type OrganizationsPanelProps = OrganizationsCollectionProps & {
  organizationOptions: Accessor<SelectOption[]>
  selectedOrgId: Accessor<string>
  setSelectedOrgId: Setter<string>
  submitCreateOrganization: (event: SubmitEvent) => void
  orgName: Accessor<string>
  setOrgName: Setter<string>
  orgSlug: Accessor<string>
  setOrgSlug: Setter<string>
  orgTenantId: Accessor<string>
  setOrgTenantId: Setter<string>
  orgPolicyName: Accessor<string>
  setOrgPolicyName: Setter<string>
  tenantOptions: Accessor<SelectOption[]>
  handleAttachTenantToOrg: () => void
  handleAttachScp: () => void
}

export function OrganizationsPanel(props: OrganizationsPanelProps) {
  return (
    <div class="space-y-5">
      <SectionTitle
        title="Organizations and SCP"
        subtitle="Organization hierarchy and service control policy guardrails."
      />
      <form
        class="grid gap-3 md:grid-cols-2"
        onSubmit={props.submitCreateOrganization}
      >
        <InputField
          label="Organization name"
          value={props.orgName()}
          onInput={(event) => props.setOrgName(event.currentTarget.value)}
        />
        <InputField
          label="Organization slug"
          value={props.orgSlug()}
          onInput={(event) => props.setOrgSlug(event.currentTarget.value)}
        />
        <div class="md:col-span-2">
          <Button type="submit" size="sm">
            Create organization
          </Button>
        </div>
      </form>

      <Show when={props.organizationOptions().length > 0}>
        <SelectField
          label="Selected organization"
          value={props.selectedOrgId()}
          options={props.organizationOptions()}
          onChange={(event) =>
            props.setSelectedOrgId(event.currentTarget.value)
          }
        />
      </Show>

      <div class="grid gap-3 md:grid-cols-2">
        <SelectField
          label="Workspace to attach"
          value={props.orgTenantId()}
          options={props.tenantOptions()}
          onChange={(event) => props.setOrgTenantId(event.currentTarget.value)}
        />
        <InputField
          label="SCP policy name"
          value={props.orgPolicyName()}
          onInput={(event) => props.setOrgPolicyName(event.currentTarget.value)}
        />
      </div>
      <div class="flex flex-wrap gap-3">
        <Button
          size="sm"
          onClick={props.handleAttachTenantToOrg}
          disabled={!props.selectedOrgId() || !props.orgTenantId()}
        >
          Attach workspace
        </Button>
        <Button
          size="sm"
          color="dark"
          onClick={props.handleAttachScp}
          disabled={!props.selectedOrgId() || !props.orgPolicyName().trim()}
        >
          Attach SCP
        </Button>
      </div>

      <OrganizationsCollection
        organizations={props.organizations}
        query={props.query}
        pageInfo={props.pageInfo}
        loading={props.loading}
        error={props.error}
        updateQuery={props.updateQuery}
        selectedOrgId={props.selectedOrgId}
        orgTenantId={props.orgTenantId}
        orgPolicies={props.orgPolicies}
        handleDetachTenantFromOrg={props.handleDetachTenantFromOrg}
        handleDetachScp={props.handleDetachScp}
      />
    </div>
  )
}
