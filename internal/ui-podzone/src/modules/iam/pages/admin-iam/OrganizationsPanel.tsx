import { For, Show, type Accessor, type Setter } from 'solid-js'
import type { OrganizationInfo, PolicyInfo } from '@/services/iam'
import {
  DataTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import {
  Badge,
  Button,
  InputField,
  SelectField,
  type SelectOption,
} from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { createClientPagination } from '@/solid/pagination'

type OrganizationsPanelProps = {
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
  organizations: Accessor<OrganizationInfo[]>
  orgPolicies: Accessor<PolicyInfo[]>
  handleDetachTenantFromOrg: (tenantID: string) => void
  handleDetachScp: (policyName: string) => void
}

export function OrganizationsPanel(props: OrganizationsPanelProps) {
  const organizationsPage = createClientPagination(props.organizations, 6)

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

      <Show
        when={props.organizations().length > 0}
        fallback={
          <EmptyBlock
            title="No organizations"
            copy="Create the first organization to start applying SCP guardrails."
          />
        }
      >
        <DataTable>
          <TableHead>
            <TableRow>
              <TableHeaderCell>Organization</TableHeaderCell>
              <TableHeaderCell>Slug</TableHeaderCell>
              <TableHeaderCell>Status</TableHeaderCell>
              <TableHeaderCell class="text-right">Actions</TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <For each={organizationsPage.pageItems()}>
              {(organization) => (
                <TableRow>
                  <TableCell class="font-semibold text-gray-900">
                    {organization.name}
                  </TableCell>
                  <TableCell class="text-gray-600">
                    {organization.slug}
                  </TableCell>
                  <TableCell>
                    <Badge
                      content={
                        organization.id === props.selectedOrgId()
                          ? 'selected'
                          : 'organization'
                      }
                      color={
                        organization.id === props.selectedOrgId()
                          ? 'blue'
                          : 'dark'
                      }
                    />
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-wrap justify-end gap-2">
                      <Show
                        when={
                          organization.id === props.selectedOrgId() &&
                          props.orgTenantId().trim()
                        }
                      >
                        <Button
                          size="xs"
                          color="light"
                          onClick={() =>
                            props.handleDetachTenantFromOrg(
                              props.orgTenantId().trim()
                            )
                          }
                        >
                          Detach selected workspace
                        </Button>
                      </Show>
                      <For
                        each={
                          organization.id === props.selectedOrgId()
                            ? props.orgPolicies()
                            : []
                        }
                      >
                        {(policy) => (
                          <Button
                            size="xs"
                            color="alternative"
                            onClick={() => props.handleDetachScp(policy.name)}
                          >
                            Detach SCP {policy.name}
                          </Button>
                        )}
                      </For>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </For>
          </TableBody>
        </DataTable>
        <Pagination
          page={organizationsPage.page()}
          pageSize={organizationsPage.pageSize}
          total={organizationsPage.total()}
          onPageChange={organizationsPage.setPage}
        />
      </Show>
    </div>
  )
}
