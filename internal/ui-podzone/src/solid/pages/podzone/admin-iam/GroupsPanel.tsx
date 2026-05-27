import { For, Show } from 'solid-js';
import { EmptyBlock } from '../../../components/common/Feedback';
import { IamStatementBuilder } from '../../../components/common/IamStatementBuilder';
import { Badge, Button, InputField, SelectField } from '../../../components/common/Primitives';
import { SectionTitle } from '../../../components/common/SectionTitle';
import { useAdminIamGroup } from './group-context';

export function GroupsPanel() {
  const group = useAdminIamGroup();

  return (
    <>
      <SectionTitle
        title="Groups"
        subtitle="Create platform or tenant groups, add members, and attach managed policies."
      />
      <form class="space-y-3" onSubmit={group.submitCreateGroup}>
        <div class="grid gap-3 md:grid-cols-2">
          <SelectField
            label="Group scope"
            value={group.groupScope()}
            options={group.groupScopeOptions}
            onChange={(e) => group.setGroupScope(e.currentTarget.value)}
          />
          <Show when={group.groupScope() === 'tenant'}>
            <SelectField
              label="Tenant"
              value={group.groupTenantId()}
              options={group.tenantOptions()}
              onChange={(e) => group.setGroupTenantId(e.currentTarget.value)}
            />
          </Show>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <InputField
            label="Group name"
            value={group.groupName()}
            onInput={(e) => group.setGroupName(e.currentTarget.value)}
          />
          <InputField
            label="Description"
            value={group.groupDescription()}
            onInput={(e) => group.setGroupDescription(e.currentTarget.value)}
          />
        </div>
        <Button type="submit" size="sm">
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
        <InputField
          label="Add member user id"
          value={group.groupMemberUserId()}
          onInput={(e) => group.setGroupMemberUserId(e.currentTarget.value)}
        />
        <InputField
          label="Attach policy name"
          value={group.groupPolicyName()}
          onInput={(e) => group.setGroupPolicyName(e.currentTarget.value)}
        />
      </div>

      <div class="flex flex-wrap gap-3">
        <Button
          size="sm"
          onClick={group.handleAddGroupMember}
          disabled={!group.selectedGroupId() || !group.groupMemberUserId().trim()}
        >
          Add member
        </Button>
        <Button
          size="sm"
          color="dark"
          onClick={group.handleAttachGroupPolicy}
          disabled={!group.selectedGroupId() || !group.groupPolicyName().trim()}
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
          <InputField
            label="Inline policy name"
            value={group.groupInlinePolicyName()}
            onInput={(e) => group.setGroupInlinePolicyName(e.currentTarget.value)}
          />
          <InputField
            label="Description"
            value={group.groupInlinePolicyDescription()}
            onInput={(e) =>
              group.setGroupInlinePolicyDescription(e.currentTarget.value)
            }
          />
        </div>
        <IamStatementBuilder
          label="Statements"
          value={group.groupInlinePolicyJson()}
          onChange={group.setGroupInlinePolicyJson}
        />
        <Button
          size="sm"
          color="dark"
          onClick={group.handleSaveGroupInlinePolicy}
          disabled={!group.selectedGroupId() || !group.groupInlinePolicyName().trim()}
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
