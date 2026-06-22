import { For, Show } from 'solid-js';
import { EmptyBlock } from '@/solid/components/common/Feedback';
import { IamStatementBuilder } from '@/solid/components/common/IamStatementBuilder';
import { Badge, Button, InputField, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { useAdminIamPolicy } from './policy-context';

export function PoliciesPanel() {
  const policy = useAdminIamPolicy();

  return (
    <>
      <SectionTitle
        title="Policies and versions"
        subtitle="Create managed policies, inspect attachments, and roll default versions."
      />
      <form class="space-y-3" onSubmit={policy.submitCreatePolicy}>
        <div class="grid gap-3 md:grid-cols-2">
          <SelectField
            label="Policy scope"
            value={policy.policyScope()}
            options={policy.policyScopeOptions}
            onChange={(e) => policy.setPolicyScope(e.currentTarget.value)}
          />
          <InputField
            label="Policy name"
            value={policy.policyName()}
            onInput={(e) => policy.setPolicyName(e.currentTarget.value)}
          />
        </div>
        <InputField
          label="Description"
          value={policy.policyDescription()}
          onInput={(e) => policy.setPolicyDescription(e.currentTarget.value)}
        />
        <IamStatementBuilder
          label="Statements"
          value={policy.policyStatementsJson()}
          onChange={policy.setPolicyStatementsJson}
        />
        <Button type="submit" size="sm">
          Create policy
        </Button>
      </form>

      <Show when={policy.policyOptions().length > 0}>
        <SelectField
          label="Inspect policy"
          value={policy.selectedPolicyName()}
          options={policy.policyOptions()}
          onChange={(e) => policy.setSelectedPolicyName(e.currentTarget.value)}
        />
      </Show>

      <Show when={policy.policyDetail()}>
        {(detail) => (
          <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600">
            <p class="font-semibold text-gray-900">{detail().name}</p>
            <p class="mt-1">{detail().description || 'No description'}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <Badge content={detail().scope} color="blue" />
              <Badge
                content={`default ${detail().defaultVersion || 'v1'}`}
                color="green"
              />
              <Show when={detail().isSystem}>
                <Badge content="system" color="dark" />
              </Show>
            </div>
          </div>
        )}
      </Show>

      <IamStatementBuilder
        label="New version statements"
        value={policy.policyVersionJson()}
        onChange={policy.setPolicyVersionJson}
      />
      <div class="flex flex-wrap gap-3">
        <Button
          size="sm"
          color="dark"
          onClick={policy.handleCreatePolicyVersion}
          disabled={!policy.selectedPolicyName()}
        >
          Create version
        </Button>
        <Button
          size="sm"
          color="red"
          onClick={policy.handleDeletePolicy}
          disabled={!policy.selectedPolicyName()}
        >
          Delete policy
        </Button>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Versions</p>
          <Show
            when={policy.policyVersions().length > 0}
            fallback={
              <EmptyBlock
                title="No versions"
                copy="Select a policy to inspect its version history."
              />
            }
          >
            <div class="space-y-3">
              <For each={policy.policyVersions()}>
                {(version) => (
                  <div class="rounded-lg border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">{version.version}</p>
                        <p class="text-sm text-gray-500">
                          {version.createdAt || 'unknown time'}
                        </p>
                      </div>
                      <Show
                        when={version.isDefault}
                        fallback={
                          <div class="flex gap-2">
                            <Button
                              size="xs"
                              color="light"
                              onClick={() =>
                                policy.handleSetDefaultVersion(version.version)
                              }
                            >
                              Set default
                            </Button>
                            <Button
                              size="xs"
                              color="red"
                              onClick={() =>
                                policy.handleDeleteVersion(version.version)
                              }
                            >
                              Delete
                            </Button>
                          </div>
                        }
                      >
                        <Badge content="default" color="green" />
                      </Show>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </div>

        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Attachments</p>
          <Show
            when={policy.policyAttachments().length > 0}
            fallback={
              <EmptyBlock
                title="No attachments"
                copy="This policy is not currently attached to any role, user, group, boundary, or SCP."
              />
            }
          >
            <div class="space-y-3">
              <For each={policy.policyAttachments()}>
                {(attachment) => (
                  <div class="rounded-lg border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={attachment.attachmentType}
                        color={policy.attachmentColor(attachment.attachmentType)}
                      />
                      <Show when={attachment.roleName}>
                        <Badge content={attachment.roleName || ''} color="dark" />
                      </Show>
                      <Show when={attachment.groupName}>
                        <Badge content={attachment.groupName || ''} color="green" />
                      </Show>
                    </div>
                    <p class="mt-2 text-sm text-gray-600">
                      {attachment.tenantId || attachment.scope || 'platform'}
                      <Show when={attachment.userId}>
                        <> · user {attachment.userId}</>
                      </Show>
                    </p>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </div>
      </div>
    </>
  );
}
