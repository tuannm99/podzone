import { Show } from 'solid-js';
import { Button, Card, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { useAdminHome } from './context';

export function StoreChooser() {
  const vm = useAdminHome();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Choose store"
        subtitle="Pick a workspace, then choose a store inside it. If the workspace has no store yet, create one right here."
      />
      <div class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr]">
        <div class="space-y-3">
          <SelectField
            label="Workspace"
            value={vm.selectedWorkspaceId()}
            options={vm.selectedWorkspaceOptions()}
            onChange={(event) => {
              vm.setSelectedWorkspaceId(event.currentTarget.value);
            }}
          />
          <p class="text-sm text-gray-600">{vm.currentSelectionLabel()}</p>
        </div>

        <div class="space-y-3">
          <Show
            when={vm.selectedWorkspace() && vm.selectedStoreOptions().length > 0}
            fallback={<CreateFirstStore />}
          >
            <SelectField
              label="Store"
              value={vm.selectedStoreId()}
              options={vm.selectedStoreOptions()}
              onChange={(event) => vm.setSelectedStoreId(event.currentTarget.value)}
            />
            <div class="flex flex-wrap gap-3">
              <Button
                disabled={!vm.selectedWorkspaceId() || !vm.selectedStoreId()}
                loading={vm.switchingTenant()}
                onClick={() => {
                  void vm.openStore(vm.selectedWorkspaceId(), vm.selectedStoreId());
                }}
              >
                Open selected store
              </Button>
              <Button
                color="light"
                disabled={!vm.selectedWorkspaceId() || !vm.selectedStoreId()}
                onClick={() => {
                  const current = vm.selectedWorkspace();
                  const store = current?.stores.find(
                    (item: { id: string }) => item.id === vm.selectedStoreId()
                  );
                  if (!current || !store) return;
                  vm.setTenantMessage(
                    `Selected ${store.name} in ${current.tenantId}.`
                  );
                }}
              >
                Confirm selection
              </Button>
            </div>
          </Show>
        </div>
      </div>
    </Card>
  );
}

function CreateFirstStore() {
  const vm = useAdminHome();

  return (
    <div class="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4">
      <p class="text-sm font-semibold text-gray-950">
        No store in this workspace yet
      </p>
      <p class="text-sm text-gray-600">
        Create the first store below, then select it from this workspace.
      </p>
      <div class="flex flex-col gap-3 sm:flex-row">
        <input
          class="block h-10 min-w-0 flex-1 rounded-md border border-gray-300 bg-white px-3 text-sm text-gray-900 outline-none transition focus:border-gray-950 focus:ring-2 focus:ring-gray-100"
          value={vm.storeNameByTenant()[vm.selectedWorkspaceId()] || ''}
          placeholder="New store name"
          onInput={(event) =>
            vm.setDraftStoreName(
              vm.selectedWorkspaceId(),
              event.currentTarget.value
            )
          }
        />
        <Button
          size="sm"
          color="dark"
          loading={vm.creatingStoreTenantId() === vm.selectedWorkspaceId()}
          disabled={
            !vm.selectedWorkspaceId() ||
            vm.creatingStoreTenantId() === vm.selectedWorkspaceId() ||
            !(vm.storeNameByTenant()[vm.selectedWorkspaceId()] || '').trim()
          }
          onClick={() => {
            void vm.submitCreateStore(vm.selectedWorkspaceId());
          }}
        >
          Create store
        </Button>
      </div>
    </div>
  );
}
