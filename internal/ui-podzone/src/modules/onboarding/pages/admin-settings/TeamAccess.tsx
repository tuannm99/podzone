import { For, Show } from 'solid-js';
import { tokenStorage } from '@/services/tokenStorage';
import { EmptyBlock, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback';
import { Badge, Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { membershipStatusColor, roleOptions } from './presentation';
import { useAdminSettings } from './context';

export function TeamAccess() {
  return (
    <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
      <MyWorkspaceAccess />
      <TeamAccessEditor />
    </div>
  );
}

function MyWorkspaceAccess() {
  const vm = useAdminSettings();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="My workspace access"
        subtitle="Workspaces this account can access right now."
      />
      <Show when={vm.loadingTenants()}>
        <LoadingInline label="Loading workspace access..." />
      </Show>
      <Show
        when={!vm.loadingTenants() && vm.memberships().length > 0}
        fallback={
          <EmptyBlock
            title="No workspace access yet"
            copy="Create or join a workspace to see your working spaces here."
          />
        }
      >
        <div class="space-y-3">
          <For each={vm.memberships()}>
            {(membership) => (
              <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
                <div class="flex flex-wrap items-center gap-2">
                  <Badge content={membership.roleName} color="blue" />
                  <Badge content={membership.status} color="green" />
                </div>
                <p class="mt-3 font-semibold text-gray-900">
                  {membership.tenantId}
                </p>
                <p class="mt-1 text-sm text-gray-500">
                  user {membership.userId}
                </p>
                <div class="mt-3">
                  <Button
                    size="sm"
                    color="alternative"
                    onClick={() => {
                      vm.setMemberTenantId(membership.tenantId);
                      void vm.loadTenantMembers(membership.tenantId);
                    }}
                  >
                    Open team access
                  </Button>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </Card>
  );
}

function TeamAccessEditor() {
  const vm = useAdminSettings();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Team access"
        subtitle="List, add, update, or remove workspace teammates. Start from one of your workspaces instead of typing technical IDs by hand."
      />

      <form class="space-y-4" onSubmit={vm.submitMember}>
        <Show when={vm.tenantOptions().length > 0}>
          <SelectField
            label="Workspace"
            value={vm.memberTenantId()}
            options={vm.tenantOptions()}
            onChange={(event) => vm.setMemberTenantId(event.currentTarget.value)}
          />
        </Show>
        <InputField
          label="Workspace id override"
          value={vm.memberTenantId()}
          placeholder="workspace id"
          onInput={(event) => vm.setMemberTenantId(event.currentTarget.value)}
        />
        <div class="grid gap-4 md:grid-cols-2">
          <InputField
            label="User id"
            value={vm.memberUserId()}
            placeholder="42"
            onInput={(event) => vm.setMemberUserId(event.currentTarget.value)}
          />
          <SelectField
            label="Role"
            value={vm.roleName()}
            options={roleOptions}
            onChange={(event) => vm.setRoleName(event.currentTarget.value)}
          />
        </div>
        <InputField
          label="Teammate email or username"
          value={vm.memberIdentity()}
          placeholder="ops@workspace.com or store_operator"
          onInput={(event) => vm.setMemberIdentity(event.currentTarget.value)}
        />
        <div class="flex flex-wrap gap-3">
          <Button
            type="button"
            color="light"
            disabled={!vm.userID}
            onClick={() => vm.setMemberUserId(String(vm.userID))}
          >
            Use my user id
          </Button>
          <Button
            type="button"
            color="light"
            disabled={!tokenStorage.getUser()?.email}
            onClick={() => vm.setMemberIdentity(tokenStorage.getUser()?.email || '')}
          >
            Use my email
          </Button>
          <Button
            type="submit"
            loading={vm.savingMember()}
            disabled={!vm.canManageMembers()}
          >
            Save access
          </Button>
          <Button
            type="button"
            color="alternative"
            disabled={!vm.canReadTenant()}
            onClick={() => void vm.loadTenantMembers()}
          >
            Reload team
          </Button>
        </div>
      </form>

      <Show when={vm.loadingMembers()}>
        <LoadingInline label="Loading workspace team..." />
      </Show>

      <Show
        when={!vm.loadingMembers() && vm.members().length > 0}
        fallback={
          <EmptyBlock
            title="No team members loaded"
            copy="Choose a workspace and reload team access to inspect who can operate in that workspace."
          />
        }
      >
        <Show when={!vm.canReadTenant()}>
          <EmptyBlock
            title="No workspace access"
            copy="You do not currently have permission to inspect team access for this workspace."
          />
        </Show>
        <Show when={vm.canReadTenant() && !vm.canManageMembers()}>
          <InfoAlert>
            You can inspect this workspace, but only authorized workspace owners or admins can manage team access.
          </InfoAlert>
        </Show>
        <div class="space-y-3">
          <For each={vm.members()}>
            {(membership) => (
              <div class="rounded-lg border border-gray-200 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="font-semibold text-gray-900">
                      teammate {membership.userId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      {membership.tenantId}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge content={membership.roleName} color="blue" />
                    <Badge
                      content={membership.status}
                      color={membershipStatusColor(membership.status)}
                    />
                    <Button
                      color="red"
                      size="xs"
                      disabled={!vm.canManageMembers()}
                      onClick={() => {
                        void vm.handleRemoveMember(
                          membership.tenantId,
                          membership.userId
                        );
                      }}
                    >
                      Remove
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </Card>
  );
}
