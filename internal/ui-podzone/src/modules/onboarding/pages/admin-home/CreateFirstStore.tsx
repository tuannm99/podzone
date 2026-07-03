import { Button } from '@/solid/components/common/Primitives'
import { FormInputField, createFormStore, required } from '@/solid/forms'
import { useAdminHome } from './context'
import type { CreateStoreFormValues } from './forms'

export function CreateFirstStore() {
  const vm = useAdminHome()
  const storeForm = createFormStore<CreateStoreFormValues>({
    initialValues: {
      name: vm.storeNameByTenant()[vm.selectedWorkspaceId()] || '',
    },
    validators: {
      name: [required('Enter a store name.')],
    },
  })

  const createStore = async () => {
    if (!storeForm.validate()) return
    storeForm.setSubmitting(true)
    try {
      const created = await vm.createStoreFromForm(vm.selectedWorkspaceId(), {
        ...storeForm.values,
      })
      if (created) storeForm.reset({ name: '' })
    } finally {
      storeForm.setSubmitting(false)
    }
  }

  return (
    <div class="space-y-3 border border-gray-200 bg-gray-50 p-4">
      <p class="text-sm font-semibold text-gray-950">
        No stores in this workspace
      </p>
      <p class="text-sm text-gray-600">
        Create the first store, then track provisioning below.
      </p>
      <div class="flex flex-col gap-3 sm:flex-row">
        <div class="min-w-0 flex-1">
          <FormInputField
            form={storeForm}
            name="name"
            label="Store name"
            placeholder="New store name"
            onValueInput={(value) =>
              vm.setDraftStoreName(vm.selectedWorkspaceId(), value)
            }
          />
        </div>
        <Button
          size="sm"
          color="dark"
          loading={vm.creatingStoreTenantId() === vm.selectedWorkspaceId()}
          disabled={
            !vm.selectedWorkspaceId() ||
            vm.creatingStoreTenantId() === vm.selectedWorkspaceId() ||
            !storeForm.values.name.trim()
          }
          onClick={() => void createStore()}
        >
          Create store
        </Button>
      </div>
    </div>
  )
}
