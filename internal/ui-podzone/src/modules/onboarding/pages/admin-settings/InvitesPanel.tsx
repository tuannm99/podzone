import { For, Show } from 'solid-js';
import { tokenStorage } from '@/services/tokenStorage';
import { EmptyBlock, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback';
import { Badge, Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { membershipStatusColor, roleOptions } from './presentation';
import { useAdminSettings } from './context';

export function InvitesPanel() {
  const vm = useAdminSettings();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Workspace invites"
        subtitle="Create email invites, track pending team access, and revoke old invite links."
      />
      <Show when={!vm.canManageMembers()}>
        <InfoAlert>
          Workspace invites require access to manage team permissions for this workspace.
        </InfoAlert>
      </Show>
      <form class="space-y-4" onSubmit={vm.submitInvite}>
        <Show when={vm.tenantOptions().length > 0}>
          <SelectField
            label="Workspace"
            value={vm.memberTenantId()}
            options={vm.tenantOptions()}
            onChange={(event) => vm.setMemberTenantId(event.currentTarget.value)}
          />
        </Show>
        <div class="grid gap-4 md:grid-cols-2">
          <InputField
            label="Invite email"
            value={vm.inviteEmail()}
            placeholder="owner@shop.com"
            onInput={(event) => vm.setInviteEmail(event.currentTarget.value)}
          />
          <SelectField
            label="Role"
            value={vm.inviteRoleName()}
            options={roleOptions}
            onChange={(event) => vm.setInviteRoleName(event.currentTarget.value)}
          />
        </div>
        <div class="flex flex-wrap gap-3">
          <Button
            type="button"
            color="light"
            disabled={!tokenStorage.getUser()?.email}
            onClick={() => vm.setInviteEmail(tokenStorage.getUser()?.email || '')}
          >
            Use my email
          </Button>
          <Button
            type="submit"
            loading={vm.savingInvite()}
            disabled={!vm.canManageMembers()}
          >
            Create workspace invite
          </Button>
          <Button
            type="button"
            color="alternative"
            disabled={!vm.canManageMembers()}
            onClick={() => void vm.loadTenantInvites()}
          >
            Reload invites
          </Button>
        </div>
      </form>

      <Show when={vm.latestInviteAcceptURL()}>
        <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-700">
          <p class="font-semibold text-gray-900">Latest join link</p>
          <p class="mt-2 break-all">{vm.latestInviteAcceptURL()}</p>
        </div>
      </Show>

      <Show when={vm.loadingInvites()}>
        <LoadingInline label="Loading workspace invites..." />
      </Show>

      <Show
        when={!vm.loadingInvites() && vm.invites().length > 0}
        fallback={
          <EmptyBlock
            title="No workspace invites loaded"
            copy="Create an invite or reload invites for the selected workspace."
          />
        }
      >
        <div class="space-y-3">
          <For each={vm.invites()}>
            {(invite) => (
              <div class="rounded-lg border border-gray-200 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="font-semibold text-gray-900">{invite.email}</p>
                    <p class="mt-1 text-sm text-gray-500">
                      workspace {invite.tenantId} · {invite.roleName} · expires{' '}
                      {invite.expiresAt || 'unknown'}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge
                      content={invite.status}
                      color={membershipStatusColor(invite.status)}
                    />
                    <Button
                      color="red"
                      size="xs"
                      disabled={!vm.canManageMembers() || invite.status !== 'pending'}
                      onClick={() => {
                        void vm.handleRevokeInvite(
                          invite.id,
                          invite.tenantId,
                          invite.email
                        );
                      }}
                    >
                      Revoke
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
