import { Show, createEffect, on } from 'solid-js'
import { useAuthContext } from '@/solid/context/auth-context'
import { ErrorAlert, InfoAlert } from '@/solid/components/common/Feedback'
import { Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { FormInputField, FormSelectField, createFormStore, numberValue, required } from '@/solid/forms'
import { useAdminSettings } from '../context'
import type { TeamMemberFormValues } from '../forms'
import { roleOptions } from '../presentation'
import { TeamMembersList } from './TeamMembersList'

export function TeamAccessEditor() {
    const auth = useAuthContext()
    const { teamAccess } = useAdminSettings()
    const { access } = teamAccess
    const memberForm = createFormStore<TeamMemberFormValues>({
        initialValues: {
            tenantId: access.selectedTenantID(),
            userId: teamAccess.userId(),
            roleName: teamAccess.roleName(),
            identity: teamAccess.identity(),
        },
        validators: {
            tenantId: [required('Choose a workspace.')],
            userId: [numberValue('User id must be a number.')],
            roleName: [required('Choose a role.')],
            identity: [
                (value, values) =>
                    !value.trim() && !values.userId.trim() ? 'Enter teammate email/username or user id.' : undefined,
            ],
        },
    })
    createEffect(on(access.selectedTenantID, (tenantID) => memberForm.setValue('tenantId', tenantID), { defer: true }))

    const submitMember = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!memberForm.validate()) return
        memberForm.setSubmitting(true)
        try {
            await teamAccess.save({ ...memberForm.values })
        } finally {
            memberForm.setSubmitting(false)
        }
    }

    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Team access"
                subtitle="List, add, update, or remove workspace teammates. Start from one of your workspaces instead of typing technical IDs by hand."
            />
            <Show when={teamAccess.error()}>
                <ErrorAlert>{teamAccess.error()}</ErrorAlert>
            </Show>
            <Show when={teamAccess.message()}>
                <InfoAlert>{teamAccess.message()}</InfoAlert>
            </Show>
            <form class="space-y-4" onSubmit={submitMember}>
                <Show when={access.tenantOptions().length > 0}>
                    <FormSelectField
                        form={memberForm}
                        name="tenantId"
                        label="Workspace"
                        options={access.tenantOptions()}
                        onValueChange={access.setSelectedTenantID}
                    />
                </Show>
                <FormInputField
                    form={memberForm}
                    name="tenantId"
                    label="Workspace id override"
                    placeholder="workspace id"
                    onValueInput={access.setSelectedTenantID}
                />
                <div class="grid gap-4 md:grid-cols-2">
                    <FormInputField form={memberForm} name="userId" label="User id" placeholder="42" />
                    <FormSelectField form={memberForm} name="roleName" label="Role" options={roleOptions} />
                </div>
                <FormInputField
                    form={memberForm}
                    name="identity"
                    label="Teammate email or username"
                    placeholder="ops@workspace.com or store_operator"
                />
                <div class="flex flex-wrap gap-3">
                    <Button
                        type="button"
                        color="light"
                        disabled={!teamAccess.userID}
                        onClick={() => memberForm.setValue('userId', String(teamAccess.userID))}
                    >
                        Use my user id
                    </Button>
                    <Button
                        type="button"
                        color="light"
                        disabled={!auth.getUserEmail()}
                        onClick={() => memberForm.setValue('identity', auth.getUserEmail())}
                    >
                        Use my email
                    </Button>
                    <Button type="submit" loading={teamAccess.saving()} disabled={!access.canManage()}>
                        Save access
                    </Button>
                    <Button
                        type="button"
                        color="alternative"
                        disabled={!access.canRead()}
                        onClick={() => void teamAccess.reload()}
                    >
                        Reload team
                    </Button>
                </div>
            </form>
            <TeamMembersList />
        </Card>
    )
}
