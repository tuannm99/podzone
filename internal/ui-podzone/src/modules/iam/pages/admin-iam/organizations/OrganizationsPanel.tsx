import { Show, type Accessor, type Setter } from 'solid-js'
import type { OrganizationMembership } from '@/services/iam'
import type { CollectionQuery, PageInfo } from '@/services/collection'
import {
  Button,
  InputField,
  SelectField,
  type SelectOption,
} from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import type { SearchSelectOption } from '@/solid/components/common/SearchSelectField'
import {
  OrganizationsCollection,
  type OrganizationsCollectionProps,
} from './OrganizationsCollection'
import { OrganizationMembersPanel } from './OrganizationMembersPanel'

type OrganizationsPanelProps = OrganizationsCollectionProps & {
  canManagePlatform: Accessor<boolean>
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
  organizationMembers: Accessor<OrganizationMembership[]>
  organizationMembersQuery: CollectionQuery
  organizationMembersPageInfo: Accessor<PageInfo>
  organizationMembersLoading: Accessor<boolean>
  organizationMembersError: Accessor<string>
  updateOrganizationMembersQuery: (patch: Partial<CollectionQuery>) => void
  handleAddOrganizationMember: (
    userID: number,
    roleName: string
  ) => Promise<void>
  handleRemoveOrganizationMember: (userID: string) => Promise<void>
  userOptions: Accessor<SearchSelectOption[]>
  usersLoading: Accessor<boolean>
  usersError: Accessor<string>
  searchUsers: (search: string) => void
}

export function OrganizationsPanel(props: OrganizationsPanelProps) {
  return (
    <div class="space-y-5">
      <SectionTitle
        title="Organizations and SCP"
        subtitle="Organization hierarchy and service control policy guardrails."
      />
      <Show when={props.canManagePlatform()}>
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
      </Show>

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

      <Show when={props.canManagePlatform()}>
        <div class="grid gap-3 md:grid-cols-2">
          <SelectField
            label="Workspace to attach"
            value={props.orgTenantId()}
            options={props.tenantOptions()}
            onChange={(event) =>
              props.setOrgTenantId(event.currentTarget.value)
            }
          />
          <InputField
            label="SCP policy name"
            value={props.orgPolicyName()}
            onInput={(event) =>
              props.setOrgPolicyName(event.currentTarget.value)
            }
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
      </Show>

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
        canManagePlatform={props.canManagePlatform}
      />
      <OrganizationMembersPanel
        organizationId={props.selectedOrgId}
        members={props.organizationMembers}
        query={props.organizationMembersQuery}
        pageInfo={props.organizationMembersPageInfo}
        loading={props.organizationMembersLoading}
        error={props.organizationMembersError}
        updateQuery={props.updateOrganizationMembersQuery}
        addMember={props.handleAddOrganizationMember}
        removeMember={props.handleRemoveOrganizationMember}
        userOptions={props.userOptions}
        usersLoading={props.usersLoading}
        usersError={props.usersError}
        searchUsers={props.searchUsers}
      />
    </div>
  )
}
