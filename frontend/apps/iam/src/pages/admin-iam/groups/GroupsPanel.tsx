import { Show } from 'solid-js'
import { IamStatementBuilder } from '../shared/IamStatementBuilder'
import { Button, SelectField } from '@podzone/shared/ui/components/common/Primitives'
import { SearchSelectField } from '@podzone/shared/ui/components/common/SearchSelectField'
import { SectionTitle } from '@podzone/shared/ui/components/common/SectionTitle'
import { FormInputField, FormSelectField, createFormStore, jsonArray, required } from '@podzone/shared/ui/forms'
import type {
    CreateGroupFormValues,
    GroupInlinePolicyFormValues,
    GroupMemberFormValues,
    GroupPolicyAttachmentFormValues,
} from './forms'
import { useAdminIamGroup } from './context'
import { GroupsCollection } from './GroupsCollection'
import { GroupAccessTables, GroupInlinePoliciesTable } from './GroupResourceTables'

export function GroupsPanel() {
    const group = useAdminIamGroup()
    const createGroupForm = createFormStore<CreateGroupFormValues>({
        initialValues: {
            scope: group.groupScope(),
            tenantId: group.groupTenantId(),
            name: group.groupName(),
            description: group.groupDescription(),
        },
        validators: {
            scope: [required('Choose a group scope.')],
            tenantId: [
                (value, values) =>
                    values.scope === 'tenant' && !value.trim() ? 'Choose a tenant for tenant groups.' : undefined,
            ],
            name: [required('Enter a group name.')],
        },
    })
    const memberForm = createFormStore<GroupMemberFormValues>({
        initialValues: { userId: group.groupMemberUserId() },
        validators: {
            userId: [required('Choose a user.')],
        },
    })
    const policyAttachmentForm = createFormStore<GroupPolicyAttachmentFormValues>({
        initialValues: { policyName: group.groupPolicyName() },
        validators: {
            policyName: [required('Choose a managed policy.')],
        },
    })
    const inlinePolicyForm = createFormStore<GroupInlinePolicyFormValues>({
        initialValues: {
            name: group.groupInlinePolicyName(),
            description: group.groupInlinePolicyDescription(),
            statementsJson: group.groupInlinePolicyJson(),
        },
        validators: {
            name: [required('Enter an inline policy name.')],
            statementsJson: [jsonArray('Inline policy statements must be a JSON array.')],
        },
    })

    const submitCreateGroup = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!createGroupForm.validate()) return
        createGroupForm.setSubmitting(true)
        try {
            await group.createGroupFromForm({ ...createGroupForm.values })
        } finally {
            createGroupForm.setSubmitting(false)
        }
    }

    const addGroupMember = async () => {
        if (!memberForm.validate()) return
        memberForm.setSubmitting(true)
        try {
            await group.addGroupMemberFromForm({ ...memberForm.values })
        } finally {
            memberForm.setSubmitting(false)
        }
    }

    const attachGroupPolicy = async () => {
        if (!policyAttachmentForm.validate()) return
        policyAttachmentForm.setSubmitting(true)
        try {
            await group.attachGroupPolicyFromForm({
                ...policyAttachmentForm.values,
            })
        } finally {
            policyAttachmentForm.setSubmitting(false)
        }
    }

    const saveGroupInlinePolicy = async () => {
        if (!inlinePolicyForm.validate()) return
        inlinePolicyForm.setSubmitting(true)
        try {
            await group.saveGroupInlinePolicyFromForm({ ...inlinePolicyForm.values })
        } finally {
            inlinePolicyForm.setSubmitting(false)
        }
    }

    return (
        <>
            <SectionTitle
                title="Groups"
                subtitle="Create platform or tenant groups, add members, and attach managed policies."
            />
            <form class="space-y-3" onSubmit={submitCreateGroup}>
                <div class="grid gap-3 md:grid-cols-2">
                    <FormSelectField
                        form={createGroupForm}
                        name="scope"
                        label="Group scope"
                        options={group.groupScopeOptions()}
                    />
                    <Show when={createGroupForm.values.scope === 'tenant'}>
                        <FormSelectField
                            form={createGroupForm}
                            name="tenantId"
                            label="Tenant"
                            options={group.tenantOptions()}
                        />
                    </Show>
                </div>
                <div class="grid gap-3 md:grid-cols-2">
                    <FormInputField form={createGroupForm} name="name" label="Group name" />
                    <FormInputField form={createGroupForm} name="description" label="Description" />
                </div>
                <Button type="submit" size="sm" loading={createGroupForm.isSubmitting()}>
                    Create group
                </Button>
            </form>

            <GroupsCollection />

            <Show when={group.groupOptions().length > 0}>
                <SelectField
                    label="Selected group"
                    value={group.selectedGroupId()}
                    options={group.groupOptions()}
                    onChange={(e) => group.setSelectedGroupId(e.currentTarget.value)}
                />
            </Show>

            <div class="grid gap-3 md:grid-cols-2">
                <SearchSelectField
                    label="Add member"
                    value={memberForm.values.userId}
                    options={group.memberUserOptions()}
                    loading={group.memberUsersLoading()}
                    error={memberForm.error('userId') || group.memberUsersError()}
                    onSearch={group.searchMemberUsers}
                    onChange={(value) => memberForm.setValue('userId', value)}
                    placeholder="Search name, username, or email"
                />
                <FormSelectField
                    form={policyAttachmentForm}
                    name="policyName"
                    label="Managed policy"
                    options={group.managedPolicyOptions()}
                />
            </div>

            <div class="flex flex-wrap gap-3">
                <Button
                    size="sm"
                    onClick={addGroupMember}
                    disabled={!group.selectedGroupId() || !memberForm.values.userId.trim()}
                    loading={memberForm.isSubmitting()}
                >
                    Add member
                </Button>
                <Button
                    size="sm"
                    color="dark"
                    onClick={attachGroupPolicy}
                    disabled={!group.selectedGroupId() || !policyAttachmentForm.values.policyName.trim()}
                    loading={policyAttachmentForm.isSubmitting()}
                >
                    Attach policy
                </Button>
                <Button size="sm" color="red" onClick={group.handleDeleteGroup} disabled={!group.selectedGroupId()}>
                    Delete group
                </Button>
            </div>

            <GroupAccessTables />

            <div class="space-y-3 border-t border-gray-200 pt-4">
                <p class="text-sm font-semibold text-gray-900">Group inline policies</p>
                <div class="grid gap-3 md:grid-cols-2">
                    <FormInputField form={inlinePolicyForm} name="name" label="Inline policy name" />
                    <FormInputField form={inlinePolicyForm} name="description" label="Description" />
                </div>
                <IamStatementBuilder
                    label="Statements"
                    actionOptions={group.permissionOptions()}
                    value={inlinePolicyForm.values.statementsJson}
                    onChange={(value) => inlinePolicyForm.setValue('statementsJson', value)}
                />
                <Show when={inlinePolicyForm.hasError('statementsJson')}>
                    <p class="text-xs font-medium text-red-600">{inlinePolicyForm.error('statementsJson')}</p>
                </Show>
                <Button
                    size="sm"
                    color="dark"
                    onClick={saveGroupInlinePolicy}
                    disabled={!group.selectedGroupId() || !inlinePolicyForm.values.name.trim()}
                    loading={inlinePolicyForm.isSubmitting()}
                >
                    Save group inline policy
                </Button>
                <GroupInlinePoliciesTable />
            </div>
        </>
    )
}
