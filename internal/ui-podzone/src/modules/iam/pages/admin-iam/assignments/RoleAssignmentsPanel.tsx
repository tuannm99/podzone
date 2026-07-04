import { Show, type Accessor, type Setter } from 'solid-js'
import {
  Button,
  SelectField,
  type SelectOption,
} from '@/solid/components/common/Primitives'
import {
  SearchSelectField,
  type SearchSelectOption,
} from '@/solid/components/common/SearchSelectField'
import { SectionTitle } from '@/solid/components/common/SectionTitle'

type RoleAssignmentsPanelProps = {
  canManagePlatform: Accessor<boolean>
  shortcutPlatformUserId: Accessor<string>
  setShortcutPlatformUserId: Setter<string>
  platformUserOptions: Accessor<SearchSelectOption[]>
  platformUsersLoading: Accessor<boolean>
  platformUsersError: Accessor<string>
  searchPlatformUsers: (search: string) => void
  shortcutPlatformRoleName: Accessor<string>
  setShortcutPlatformRoleName: Setter<string>
  platformRoleOptions: SelectOption[]
  handleAssignPlatformRole: () => void
  handleRemovePlatformRoleShortcut: () => void
  shortcutTenantId: Accessor<string>
  setShortcutTenantId: Setter<string>
  shortcutTenantUserId: Accessor<string>
  setShortcutTenantUserId: Setter<string>
  tenantUserOptions: Accessor<SearchSelectOption[]>
  tenantUsersLoading: Accessor<boolean>
  tenantUsersError: Accessor<string>
  searchTenantUsers: (search: string) => void
  shortcutTenantRoleName: Accessor<string>
  setShortcutTenantRoleName: Setter<string>
  tenantOptions: Accessor<SelectOption[]>
  tenantRoleOptions: SelectOption[]
  handleAssignTenantRole: () => void
  handleRemoveTenantMembershipShortcut: () => void
}

export function RoleAssignmentsPanel(props: RoleAssignmentsPanelProps) {
  return (
    <div class="space-y-5">
      <SectionTitle
        title="Role assignments"
        subtitle="Platform roles and workspace memberships."
      />
      <div
        class="grid gap-6"
        classList={{ 'lg:grid-cols-2': props.canManagePlatform() }}
      >
        <Show when={props.canManagePlatform()}>
          <section class="space-y-3 border-b border-gray-200 pb-6 lg:border-b-0 lg:border-r lg:pb-0 lg:pr-6">
            <p class="text-sm font-semibold text-gray-900">Platform role</p>
            <SearchSelectField
              label="User"
              value={props.shortcutPlatformUserId()}
              options={props.platformUserOptions()}
              loading={props.platformUsersLoading()}
              error={props.platformUsersError()}
              onSearch={props.searchPlatformUsers}
              onChange={props.setShortcutPlatformUserId}
              placeholder="Search name, username, or email"
            />
            <SelectField
              label="Platform role"
              value={props.shortcutPlatformRoleName()}
              options={props.platformRoleOptions}
              onChange={(event) =>
                props.setShortcutPlatformRoleName(event.currentTarget.value)
              }
            />
            <div class="flex flex-wrap gap-3">
              <Button
                size="sm"
                onClick={props.handleAssignPlatformRole}
                disabled={!props.shortcutPlatformUserId().trim()}
              >
                Assign role
              </Button>
              <Button
                size="sm"
                color="red"
                onClick={props.handleRemovePlatformRoleShortcut}
                disabled={!props.shortcutPlatformUserId().trim()}
              >
                Remove role
              </Button>
            </div>
          </section>
        </Show>

        <section class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">
            Workspace membership
          </p>
          <SelectField
            label="Workspace"
            value={props.shortcutTenantId()}
            options={props.tenantOptions()}
            onChange={(event) =>
              props.setShortcutTenantId(event.currentTarget.value)
            }
          />
          <SearchSelectField
            label="User"
            value={props.shortcutTenantUserId()}
            options={props.tenantUserOptions()}
            loading={props.tenantUsersLoading()}
            error={props.tenantUsersError()}
            onSearch={props.searchTenantUsers}
            onChange={props.setShortcutTenantUserId}
            placeholder="Search name, username, or email"
          />
          <SelectField
            label="Workspace role"
            value={props.shortcutTenantRoleName()}
            options={props.tenantRoleOptions}
            onChange={(event) =>
              props.setShortcutTenantRoleName(event.currentTarget.value)
            }
          />
          <div class="flex flex-wrap gap-3">
            <Button
              size="sm"
              onClick={props.handleAssignTenantRole}
              disabled={
                !props.shortcutTenantId().trim() ||
                !props.shortcutTenantUserId().trim()
              }
            >
              Assign membership
            </Button>
            <Button
              size="sm"
              color="red"
              onClick={props.handleRemoveTenantMembershipShortcut}
              disabled={
                !props.shortcutTenantId().trim() ||
                !props.shortcutTenantUserId().trim()
              }
            >
              Remove membership
            </Button>
          </div>
        </section>
      </div>
    </div>
  )
}
