import { Show } from 'solid-js';
import { Badge, Button, Card, InputField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { useAdminHome } from './context';

export function WorkspaceSetup() {
  const vm = useAdminHome();

  return (
    <Show when={vm.canCreateTenant() || vm.canBootstrapFirstWorkspace()}>
      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create workspace"
            subtitle="Create the tenant workspace that will own your stores and IAM memberships."
          />
          <form class="space-y-4" onSubmit={vm.submitCreateTenant}>
            <InputField
              label="Workspace name"
              value={vm.tenantName()}
              placeholder="Urban Finds"
              onInput={(event) => {
                const value = event.currentTarget.value;
                vm.setTenantName(value);
                if (!vm.tenantSlug().trim()) {
                  vm.setTenantSlug(vm.slugify(value));
                }
              }}
            />
            <InputField
              label="Workspace slug"
              value={vm.tenantSlug()}
              placeholder="urban-finds"
              onInput={(event) =>
                vm.setTenantSlug(vm.slugify(event.currentTarget.value))
              }
            />
            <div class="flex flex-wrap gap-3">
              <Button
                type="submit"
                loading={vm.creatingTenant()}
                disabled={
                  !vm.tenantName().trim() ||
                  (!vm.canCreateTenant() && !vm.canBootstrapFirstWorkspace())
                }
              >
                Create workspace
              </Button>
              <Badge
                content={
                  vm.tenantSlug().trim()
                    ? `slug ${vm.tenantSlug().trim()}`
                    : 'slug pending'
                }
                color={vm.tenantSlug().trim() ? 'indigo' : 'dark'}
              />
            </div>
          </form>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Current workspace"
            subtitle="The workspace you are preparing to enter."
          />
          <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
            <p class="text-sm font-semibold text-gray-950">
              {vm.selectedWorkspace()?.tenantId || 'No workspace selected'}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              {vm.selectedWorkspace()?.roleName || 'Select a workspace above'}
            </p>
          </div>
          <Button
            color="alternative"
            disabled={!vm.selectedWorkspaceId()}
            onClick={() => {
              void vm.prepareTenant(vm.selectedWorkspaceId());
            }}
          >
            Reload workspace
          </Button>
        </Card>
      </div>
    </Show>
  );
}
