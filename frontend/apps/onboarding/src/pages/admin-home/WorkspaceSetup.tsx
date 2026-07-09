import { Badge, Button, Card } from '@podzone/shared/ui/components/common/Primitives'
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { FormInputField, createFormStore, required } from '@podzone/shared/ui/forms'
import type { CreateWorkspaceFormValues } from './forms'
import { useAdminHome } from './context'

export function WorkspaceSetup() {
    const vm = useAdminHome()
    const workspaceForm = createFormStore<CreateWorkspaceFormValues>({
        initialValues: {
            name: vm.tenantName(),
            slug: vm.tenantSlug(),
        },
        validators: {
            name: [required('Enter a workspace name.')],
            slug: [required('Enter a workspace slug.')],
        },
    })

    const submitWorkspace = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!workspaceForm.validate()) return
        workspaceForm.setSubmitting(true)
        try {
            const created = await vm.createTenantFromForm({ ...workspaceForm.values })
            if (created) {
                workspaceForm.reset({ name: '', slug: '' })
            }
        } finally {
            workspaceForm.setSubmitting(false)
        }
    }

    return (
        <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
            <Card class="space-y-4">
                <SectionTitle
                    title="Create workspace"
                    subtitle="Create the tenant workspace that will own your stores and IAM memberships."
                />
                <form class="space-y-4" onSubmit={submitWorkspace}>
                    <FormInputField
                        form={workspaceForm}
                        name="name"
                        label="Workspace name"
                        placeholder="Urban Finds"
                        onValueInput={(value) => {
                            if (!workspaceForm.values.slug.trim()) {
                                workspaceForm.setValue('slug', vm.slugify(value))
                            }
                        }}
                    />
                    <FormInputField
                        form={workspaceForm}
                        name="slug"
                        label="Workspace slug"
                        placeholder="urban-finds"
                        onValueInput={(value) => workspaceForm.setValue('slug', vm.slugify(value))}
                    />
                    <div class="flex flex-wrap gap-3">
                        <Button
                            type="submit"
                            loading={vm.creatingTenant()}
                            disabled={!workspaceForm.values.name.trim()}
                        >
                            Create workspace
                        </Button>
                        <Badge
                            content={
                                workspaceForm.values.slug.trim()
                                    ? `slug ${workspaceForm.values.slug.trim()}`
                                    : 'slug pending'
                            }
                            color={workspaceForm.values.slug.trim() ? 'indigo' : 'dark'}
                        />
                    </div>
                </form>
            </Card>

            <Card class="space-y-4">
                <SectionTitle title="Current workspace" subtitle="The workspace you are preparing to enter." />
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
                        void vm.prepareTenant(vm.selectedWorkspaceId())
                    }}
                >
                    Reload workspace
                </Button>
            </Card>
        </div>
    )
}
