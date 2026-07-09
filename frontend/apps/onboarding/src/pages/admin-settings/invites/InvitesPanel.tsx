import { Show, createEffect, on } from 'solid-js'
import { ErrorAlert, InfoAlert } from '@podzone/shared/ui/components/common/Feedback'
import { Button, Card } from '@podzone/shared/ui/components/common/Primitives'
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { FormInputField, FormSelectField, createFormStore, email, required } from '@podzone/shared/ui/forms'
import { useAdminSettings } from '../context'
import type { TenantInviteFormValues } from '../forms'
import { roleOptions } from '../presentation'
import { InvitesTable } from './InvitesTable'

export function InvitesPanel() {
    const { invites } = useAdminSettings()
    const { access } = invites
    const inviteForm = createFormStore<TenantInviteFormValues>({
        initialValues: {
            tenantId: access.selectedTenantID(),
            email: invites.email(),
            roleName: invites.roleName(),
        },
        validators: {
            tenantId: [required('Choose a workspace.')],
            email: [required('Invite email is required.'), email('Enter a valid email.')],
            roleName: [required('Choose a role.')],
        },
    })
    createEffect(on(access.selectedTenantID, (tenantID) => inviteForm.setValue('tenantId', tenantID), { defer: true }))

    const submitInvite = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!inviteForm.validate()) return
        inviteForm.setSubmitting(true)
        try {
            await invites.create({ ...inviteForm.values })
        } finally {
            inviteForm.setSubmitting(false)
        }
    }

    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Workspace invites"
                subtitle="Create email invites, track pending team access, and revoke old invite links."
            />
            <Show when={invites.error()}>
                <ErrorAlert>{invites.error()}</ErrorAlert>
            </Show>
            <Show when={invites.message()}>
                <InfoAlert>{invites.message()}</InfoAlert>
            </Show>
            <Show when={!access.loadingAccess() && !access.canManage()}>
                <InfoAlert>Workspace invites require access to manage team permissions for this workspace.</InfoAlert>
            </Show>
            <form class="space-y-4" onSubmit={submitInvite}>
                <Show when={access.tenantOptions().length > 0}>
                    <FormSelectField
                        form={inviteForm}
                        name="tenantId"
                        label="Workspace"
                        options={access.tenantOptions()}
                        onValueChange={access.setSelectedTenantID}
                    />
                </Show>
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField
                        form={inviteForm}
                        name="email"
                        label="Invite email"
                        placeholder="owner@shop.com"
                        onValueInput={invites.setEmail}
                    />
                    <FormSelectField
                        form={inviteForm}
                        name="roleName"
                        label="Role"
                        options={roleOptions}
                        onValueChange={invites.setRoleName}
                    />
                </div>
                <div class="flex flex-wrap gap-3">
                    <Button
                        type="button"
                        color="light"
                        disabled={!invites.currentUserEmail()}
                        onClick={() => {
                            const emailAddress = invites.currentUserEmail()
                            inviteForm.setValue('email', emailAddress)
                            invites.setEmail(emailAddress)
                        }}
                    >
                        Use my email
                    </Button>
                    <Button type="submit" loading={invites.saving()} disabled={!access.canManage()}>
                        Create workspace invite
                    </Button>
                    <Button
                        type="button"
                        color="alternative"
                        disabled={!access.canManage()}
                        onClick={() => void invites.reload()}
                    >
                        Reload invites
                    </Button>
                </div>
            </form>

            <Show when={invites.latestAcceptURL()}>
                <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-700">
                    <p class="font-semibold text-gray-900">Latest join link</p>
                    <p class="mt-2 break-all">{invites.latestAcceptURL()}</p>
                </div>
            </Show>
            <InvitesTable />
        </Card>
    )
}
