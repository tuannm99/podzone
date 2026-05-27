import { For, Show } from 'solid-js';
import { EmptyBlock } from '../../../components/common/Feedback';
import { IamStatementBuilder } from '../../../components/common/IamStatementBuilder';
import { Badge, Button, InputField, SelectField } from '../../../components/common/Primitives';
import { SectionTitle } from '../../../components/common/SectionTitle';
import {
  useAdminIamPrincipal,
  type PrincipalMode,
} from './principal-context';

export function PrincipalPoliciesPanel() {
  const principal = useAdminIamPrincipal();

  return (
    <>
      <SectionTitle
        title="Principal policies"
        subtitle="Manage direct policies, inline policies, and permission boundaries for platform and tenant users."
      />
      <div class="grid gap-3 md:grid-cols-2">
        <SelectField
          label="Principal mode"
          value={principal.principalMode()}
          options={[
            { name: 'Platform user', value: 'platform' },
            { name: 'Tenant user', value: 'tenant' },
          ]}
          onChange={(e) =>
            principal.setPrincipalMode(e.currentTarget.value as PrincipalMode)
          }
        />
        <Show
          when={principal.principalMode() === 'platform'}
          fallback={
            <SelectField
              label="Tenant"
              value={principal.principalTenantId()}
              options={principal.tenantOptions()}
              onChange={(e) => principal.setPrincipalTenantId(e.currentTarget.value)}
            />
          }
        >
          <InputField
            label="Platform user id"
            value={principal.principalPlatformUserId()}
            onInput={(e) => principal.setPrincipalPlatformUserId(e.currentTarget.value)}
          />
        </Show>
      </div>
      <Show when={principal.principalMode() === 'tenant'}>
        <InputField
          label="Tenant user id"
          value={principal.principalTenantUserId()}
          onInput={(e) => principal.setPrincipalTenantUserId(e.currentTarget.value)}
        />
      </Show>

      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
        <InputField
          label="Managed policy name"
          value={principal.principalManagedPolicyName()}
          onInput={(e) => principal.setPrincipalManagedPolicyName(e.currentTarget.value)}
        />
        <Button
          size="sm"
          onClick={principal.handleAttachPrincipalManagedPolicy}
          disabled={!principal.principalManagedPolicyName().trim()}
        >
          Attach policy
        </Button>
      </div>

      <div class="grid gap-4 lg:grid-cols-2">
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Managed policies</p>
          <p class="text-xs text-gray-500">
            Direct attachments only affect this principal and stack with group,
            role, boundary, and SCP evaluation.
          </p>
          <Show
            when={principal.currentManagedPolicies().length > 0}
            fallback={
              <EmptyBlock
                title="No direct policies"
                copy="Attach a managed policy to scope direct user access."
              />
            }
          >
            <div class="flex flex-wrap gap-2">
              <For each={principal.currentManagedPolicies()}>
                {(policy) => (
                  <button
                    class="inline-flex"
                    type="button"
                    onClick={() =>
                      principal.handleDetachPrincipalManagedPolicy(policy.name)
                    }
                  >
                    <Badge content={`${policy.name} ×`} color="blue" />
                  </button>
                )}
              </For>
            </div>
          </Show>
        </div>
        <div class="space-y-3">
          <p class="text-sm font-semibold text-gray-900">Permission boundary</p>
          <p class="text-xs text-gray-500">
            Boundary policies cap the maximum access this principal can exercise,
            even when identity policies allow more.
          </p>
          <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-end">
            <InputField
              label="Boundary policy"
              value={principal.principalBoundaryPolicyName()}
              onInput={(e) =>
                principal.setPrincipalBoundaryPolicyName(e.currentTarget.value)
              }
            />
            <Button
              size="sm"
              color="dark"
              onClick={principal.handleSavePrincipalBoundary}
              disabled={!principal.principalBoundaryPolicyName().trim()}
            >
              Save
            </Button>
            <Button
              size="sm"
              color="red"
              onClick={principal.handleDeletePrincipalBoundary}
            >
              Delete
            </Button>
          </div>
          <Show when={principal.currentBoundary()}>
            {(boundary) => (
              <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600">
                <div class="flex flex-wrap gap-2">
                  <Badge content="boundary" color="pink" />
                  <Badge content={boundary().policyName} color="blue" />
                </div>
              </div>
            )}
          </Show>
        </div>
      </div>

      <div class="space-y-3">
        <div class="grid gap-3 md:grid-cols-2">
          <InputField
            label="Inline policy name"
            value={principal.principalInlinePolicyName()}
            onInput={(e) =>
              principal.setPrincipalInlinePolicyName(e.currentTarget.value)
            }
          />
          <InputField
            label="Inline policy description"
            value={principal.principalInlinePolicyDescription()}
            onInput={(e) =>
              principal.setPrincipalInlinePolicyDescription(e.currentTarget.value)
            }
          />
        </div>
        <IamStatementBuilder
          label="Inline policy statements"
          value={principal.principalInlinePolicyJson()}
          onChange={principal.setPrincipalInlinePolicyJson}
        />
        <Button
          size="sm"
          color="dark"
          onClick={principal.handleSavePrincipalInlinePolicy}
          disabled={!principal.principalInlinePolicyName().trim()}
        >
          Save inline policy
        </Button>
        <Show
          when={principal.currentInlinePolicies().length > 0}
          fallback={
            <EmptyBlock
              title="No inline policies"
              copy="Create an inline policy when this permission should live only on one principal."
            />
          }
        >
          <div class="space-y-3">
            <For each={principal.currentInlinePolicies()}>
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
                      onClick={() =>
                        principal.handleDeletePrincipalInlinePolicy(policy.name)
                      }
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
