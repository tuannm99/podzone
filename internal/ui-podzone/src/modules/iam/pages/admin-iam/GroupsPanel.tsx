import { For, Show } from 'solid-js';
import { EmptyBlock } from '@/solid/components/common/Feedback';
import { IamStatementBuilder } from '@/solid/components/common/IamStatementBuilder';
import { Badge, Button, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import {
  FormInputField,
  FormSelectField,
  createFormStore,
  jsonArray,
  numberValue,
  required,
} from '@/solid/forms';
import type {
  CreateGroupFormValues,
  GroupInlinePolicyFormValues,
  GroupMemberFormValues,
  GroupPolicyAttachmentFormValues,
} from './group-forms';
import { useAdminIamGroup } from './group-context';

export function GroupsPanel() {
  const group = useAdminIamGroup();
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
          values.scope === 'tenant' && !value.trim()
            ? 'Choose a tenant for tenant groups.'
            : undefined,
      ],
      name: [required('Enter a group name.')],
    },
  });
  const memberForm = createFormStore<GroupMemberFormValues>({
    initialValues: { userId: group.groupMemberUserId() },
    validators: {
      userId: [
        required('Enter a user id.'),
        numberValue('User id must be a number.'),
      ],
    },
  });
  const policyAttachmentForm =
    createFormStore<GroupPolicyAttachmentFormValues>({
      initialValues: { policyName: group.groupPolicyName() },
      validators: {
        policyName: [required('Enter a policy name.')],
      },
    });
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
  });

  const submitCreateGroup = async (event: SubmitEvent) => {
    event.preventDefault();
    if (!createGroupForm.validate()) return;
    createGroupForm.setSubmitting(true);
    await group.createGroupFromForm({ ...createGroupForm.values });
    createGroupForm.setSubmitting(false);
  };

  const addGroupMember = async () => {
    if (!memberForm.validate()) return;
    memberForm.setSubmitting(true);
    await group.addGroupMemberFromForm({ ...memberForm.values });
    memberForm.setSubmitting(false);
  };

  const attachGroupPolicy = async () => {
    if (!policyAttachmentForm.validate()) return;
    policyAttachmentForm.setSubmitting(true);
    await group.attachGroupPolicyFromForm({
      ...policyAttachmentForm.values,
    });
    policyAttachmentForm.setSubmitting(false);
  };

  const saveGroupInlinePolicy = async () => {
    if (!inlinePolicyForm.validate()) return;
    inlinePolicyForm.setSubmitting(true);
    await group.saveGroupInlinePolicyFromForm({ ...inlinePolicyForm.values });
    inlinePolicyForm.setSubmitting(false);
  };

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
            options={group.groupScopeOptions}
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
          <FormInputField
            form={createGroupForm}
            name="name"
            label="Group name"
          />
          <FormInputField
            form={createGroupForm}
            name="description"
            label="Description"
          />
        </div>
        <Button
          type="submit"
          size="sm"
          loading={createGroupForm.isSubmitting()}
        >
          Create group
        </Button>
      </form>

      <Show when={group.groupOptions().length > 0}>
        <SelectField
          label="Selected group"
          value={group.selectedGroupId()}
          options={group.groupOptions()}
          onChange={(e) => group.setSelectedGroupId(e.currentTarget.value)}
        />
      </Show>

      <div class="grid gap-3 md:grid-cols-2">
        <FormInputField
          form={memberForm}
          name="userId"
          label="Add member user id"
        />
        <FormInputField
          form={policyAttachmentForm}
          name="policyName"
          label="Attach policy name"
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
          disabled={
            !group.selectedGroupId() ||
            !policyAttachmentForm.values.policyName.trim()
          }
          loading={policyAttachmentForm.isSubmitting()}
        >
          Attach policy
        </Button>
        <Button
          size="sm"
          color="red"
          onClick={group.handleDeleteGroup}
          disabled={!group.selectedGroupId()}
        >
          Delete group
        </Button>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Members</p>
          <Show
            when={group.groupMembers().length > 0}
            fallback={
              <EmptyBlock
                title="No group members"
                copy="Select a group and add users to start deriving permissions from group membership."
              />
            }
          >
            <div class="flex flex-wrap gap-2">
              <For each={group.groupMembers()}>
                {(userId) => (
                  <button
                    class="inline-flex"
                    type="button"
                    onClick={() => group.handleRemoveGroupMember(userId)}
                  >
                    <Badge content={`user ${userId} ×`} color="green" />
                  </button>
                )}
              </For>
            </div>
          </Show>
        </div>
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Attached policies</p>
          <Show
            when={group.groupPolicies().length > 0}
            fallback={
              <EmptyBlock
                title="No group policies"
                copy="Attach a managed policy to use this group as a reusable permission bundle."
              />
            }
          >
            <div class="flex flex-wrap gap-2">
              <For each={group.groupPolicies()}>
                {(policy) => (
                  <button
                    class="inline-flex"
                    type="button"
                    onClick={() => group.handleDetachGroupPolicy(policy.name)}
                  >
                    <Badge content={`${policy.name} ×`} color="blue" />
                  </button>
                )}
              </For>
            </div>
          </Show>
        </div>
      </div>

      <div class="space-y-3 border-t border-gray-200 pt-4">
        <p class="text-sm font-semibold text-gray-900">Group inline policies</p>
        <div class="grid gap-3 md:grid-cols-2">
          <FormInputField
            form={inlinePolicyForm}
            name="name"
            label="Inline policy name"
          />
          <FormInputField
            form={inlinePolicyForm}
            name="description"
            label="Description"
          />
        </div>
        <IamStatementBuilder
          label="Statements"
          value={inlinePolicyForm.values.statementsJson}
          onChange={(value) =>
            inlinePolicyForm.setValue('statementsJson', value)
          }
        />
        <Show when={inlinePolicyForm.hasError('statementsJson')}>
          <p class="text-xs font-medium text-red-600">
            {inlinePolicyForm.error('statementsJson')}
          </p>
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
        <Show
          when={group.groupInlinePolicies().length > 0}
          fallback={
            <EmptyBlock
              title="No group inline policies"
              copy="Use inline policies when permissions should live only on one group instead of a shared managed policy."
            />
          }
        >
          <div class="space-y-3">
            <For each={group.groupInlinePolicies()}>
              {(policy) => (
                <div class="rounded-lg border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">{policy.name}</p>
                      <p class="text-sm text-gray-500">
                        {policy.description || 'No description'}
                      </p>
                    </div>
                    <Button
                      size="xs"
                      color="red"
                      onClick={() => group.handleDeleteGroupInlinePolicy(policy.name)}
                    >
                      Delete
                    </Button>
                  </div>
                  <p class="mt-3 text-xs text-gray-500">
                    {policy.statements?.length || 0} statements
                  </p>
                </div>
              )}
            </For>
          </div>
        </Show>
      </div>
    </>
  );
}
