import { For, Show } from 'solid-js';
import { EmptyBlock, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback';
import { Badge, Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { membershipStatusColor, platformRoleOptions } from './presentation';
import { useAdminSettings } from './context';

export function PlatformAdminPanel() {
  const vm = useAdminSettings();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Platform administration"
        subtitle="Assign or revoke platform-wide roles such as workspace creation and admin governance."
      />
      <Show when={!vm.canManagePlatformRoles()}>
        <InfoAlert>
          Platform administration requires dedicated platform access.
        </InfoAlert>
      </Show>
      <form class="space-y-4" onSubmit={vm.submitPlatformRole}>
        <div class="grid gap-4 md:grid-cols-2">
          <InputField
            label="Target user id"
            value={vm.platformUserId()}
            placeholder="1"
            onInput={(event) => vm.setPlatformUserId(event.currentTarget.value)}
          />
          <SelectField
            label="Platform admin role"
            value={vm.platformRoleName()}
            options={platformRoleOptions}
            onChange={(event) => vm.setPlatformRoleName(event.currentTarget.value)}
          />
        </div>
        <div class="flex flex-wrap gap-3">
          <Button
            type="button"
            color="light"
            disabled={!vm.userID}
            onClick={() => vm.setPlatformUserId(String(vm.userID))}
          >
            Use my user id
          </Button>
          <Button
            type="submit"
            loading={vm.savingPlatformRole()}
            disabled={!vm.canManagePlatformRoles()}
          >
            Save admin role
          </Button>
          <Button
            type="button"
            color="alternative"
            disabled={!vm.canManagePlatformRoles()}
            onClick={() => void vm.loadPlatformRoleAssignments()}
          >
            Reload admin roles
          </Button>
        </div>
      </form>

      <Show when={vm.loadingPlatformRoles()}>
        <LoadingInline label="Loading admin roles..." />
      </Show>

      <Show
        when={!vm.loadingPlatformRoles() && vm.platformRoles().length > 0}
        fallback={
          <EmptyBlock
            title="No admin roles loaded"
            copy="Choose a target user to inspect platform-level administration access."
          />
        }
      >
        <div class="space-y-3">
          <For each={vm.platformRoles()}>
            {(membership) => (
              <div class="rounded-lg border border-gray-200 p-4">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p class="font-semibold text-gray-900">
                      user {membership.userId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      {membership.roleName}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge
                      content={membership.status}
                      color={membershipStatusColor(membership.status)}
                    />
                    <Button
                      color="red"
                      size="xs"
                      disabled={!vm.canManagePlatformRoles()}
                      onClick={() => {
                        void vm.handleRemovePlatformRole(
                          membership.userId,
                          membership.roleName
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
