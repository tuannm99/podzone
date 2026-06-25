import { For, Show } from 'solid-js';
import { EmptyBlock, InfoAlert } from '@/solid/components/common/Feedback';
import { IamKeyValueBuilder } from '@/solid/components/common/IamKeyValueBuilder';
import { IamStatementBuilder } from '@/solid/components/common/IamStatementBuilder';
import { IamTrustPolicyBuilder } from '@/solid/components/common/IamTrustPolicyBuilder';
import { Badge, Button, Card, InputField, SelectField } from '@/solid/components/common/Primitives';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import {
  FormInputField,
  createFormStore,
  jsonArray,
  required,
} from '@/solid/forms';
import { classes } from '@/solid/shared/utils';
import type {
  RoleBoundaryFormValues,
  TrustPolicyFormValues,
} from './trust-forms';
import { useAdminIamTrustSim } from './trust-sim-context';

export function TrustSimulationPanel() {
  const trust = useAdminIamTrustSim();
  const trustPolicyForm = createFormStore<TrustPolicyFormValues>({
    initialValues: {
      roleName: trust.trustRoleName(),
      trustJson: trust.trustJson(),
    },
    validators: {
      roleName: [required('Enter a role name.')],
      trustJson: [jsonArray('Trust policy must be a JSON array.')],
    },
  });
  const roleBoundaryForm = createFormStore<RoleBoundaryFormValues>({
    initialValues: {
      policyName: trust.trustBoundaryPolicyName(),
    },
    validators: {
      policyName: [required('Enter a role boundary policy name.')],
    },
  });

  const saveTrustPolicy = async () => {
    if (!trustPolicyForm.validate()) return;
    trustPolicyForm.setSubmitting(true);
    trust.setTrustRoleName(trustPolicyForm.values.roleName);
    trust.setTrustJson(trustPolicyForm.values.trustJson);
    await trust.handleSaveTrustPolicy();
    trustPolicyForm.setSubmitting(false);
  };

  const saveRoleBoundary = async () => {
    if (!roleBoundaryForm.validate()) return;
    roleBoundaryForm.setSubmitting(true);
    trust.setTrustBoundaryPolicyName(roleBoundaryForm.values.policyName);
    await trust.handleSaveRoleBoundary();
    roleBoundaryForm.setSubmitting(false);
  };

  return (
    <>
      <SectionTitle
        title="Role trust and access simulation"
        subtitle="Edit trust policies, test service principals, session tags, conditions, boundaries, and SCP outcomes."
      />
      <div class="grid gap-3 md:grid-cols-2">
        <FormInputField
          form={trustPolicyForm}
          name="roleName"
          label="Role name"
        />
        <InfoAlert>
          Use the trust policy builder to define which user, role, or service
          principals may assume this role.
        </InfoAlert>
      </div>
      <div class="flex flex-wrap gap-3">
        <Button size="sm" color="light" onClick={trust.loadTrustPolicy}>
          Load trust policy
        </Button>
        <Button
          size="sm"
          onClick={saveTrustPolicy}
          loading={trustPolicyForm.isSubmitting()}
        >
          Save trust policy
        </Button>
      </div>
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-end">
        <FormInputField
          form={roleBoundaryForm}
          name="policyName"
          label="Role boundary policy"
        />
        <Button
          size="sm"
          color="dark"
          onClick={saveRoleBoundary}
          loading={roleBoundaryForm.isSubmitting()}
        >
          Save boundary
        </Button>
        <Button size="sm" color="red" onClick={trust.handleDeleteRoleBoundary}>
          Delete boundary
        </Button>
      </div>
      <Show when={trust.roleBoundary()}>
        {(boundary) => (
          <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600">
            <div class="flex flex-wrap gap-2">
              <Badge content="role boundary" color="pink" />
              <Badge content={boundary().policyName} color="blue" />
            </div>
          </div>
        )}
      </Show>
      <IamTrustPolicyBuilder
        label="Trust policy"
        value={trustPolicyForm.values.trustJson}
        onChange={(value) => trustPolicyForm.setValue('trustJson', value)}
      />
      <Show when={trustPolicyForm.hasError('trustJson')}>
        <p class="text-xs font-medium text-red-600">
          {trustPolicyForm.error('trustJson')}
        </p>
      </Show>

      <div class="grid gap-3 md:grid-cols-2">
        <SelectField
          label="Simulation scope"
          value={trust.simScope()}
          options={trust.policyScopeOptions}
          onChange={(e) => trust.setSimScope(e.currentTarget.value)}
        />
        <InputField
          label="Target user id"
          value={trust.simTargetUserId()}
          onInput={(e) => trust.setSimTargetUserId(e.currentTarget.value)}
        />
      </div>
      <InfoAlert>
        Simulation evaluates identity policies, group policies, trust,
        boundaries, session scope-down, and SCP layers together.
      </InfoAlert>
      <div class="grid gap-3 md:grid-cols-2">
        <Show when={trust.simScope() === 'tenant'}>
          <SelectField
            label="Simulation tenant"
            value={trust.simTenantId()}
            options={trust.tenantOptions()}
            onChange={(e) => trust.setSimTenantId(e.currentTarget.value)}
          />
        </Show>
        <InputField
          label="Action"
          value={trust.simAction()}
          onInput={(e) => trust.setSimAction(e.currentTarget.value)}
        />
      </div>
      <InputField
        label="Resource"
        value={trust.simResource()}
        onInput={(e) => trust.setSimResource(e.currentTarget.value)}
      />
      <InputField
        label="Service principal"
        value={trust.simServicePrincipal()}
        onInput={(e) => trust.setSimServicePrincipal(e.currentTarget.value)}
      />
      <div class="flex flex-wrap gap-2">
        <Button size="xs" color="light" onClick={trust.applyServiceAssumePreset}>
          Preset: service assume
        </Button>
        <Button size="xs" color="light" onClick={trust.applyTenantAssumePreset}>
          Preset: tenant admin assume
        </Button>
        <Button size="xs" color="light" onClick={trust.applyScopeDownDenyPreset}>
          Preset: scope-down deny
        </Button>
      </div>
      <div class="grid gap-3 lg:grid-cols-2">
        <IamKeyValueBuilder
          label="Attributes"
          helper="Pass request attributes used by policy conditions, such as lane, region, or source identity hints."
          value={trust.simAttributesJson()}
          emptyKeyPlaceholder="lane"
          emptyValuePlaceholder="priority"
          badgeLabel="attributes"
          addLabel="Add attribute"
          onChange={trust.setSimAttributesJson}
        />
        <IamKeyValueBuilder
          label="Session tags"
          helper="Model AWS-style session tags that scope policy conditions during simulation."
          value={trust.simSessionTagsJson()}
          emptyKeyPlaceholder="team"
          emptyValuePlaceholder="ops"
          badgeLabel="tags"
          addLabel="Add tag"
          onChange={trust.setSimSessionTagsJson}
        />
      </div>
      <div class="grid gap-3 lg:grid-cols-2">
        <IamStatementBuilder
          label="Session policy"
          value={trust.simSessionPolicyJson()}
          onChange={trust.setSimSessionPolicyJson}
        />
        <Card class="space-y-4 border border-gray-200 bg-gray-50 p-4 shadow-none">
          <div>
            <p class="text-sm font-medium text-gray-700">Assumed role session</p>
            <p class="mt-1 text-xs text-gray-500">
              Provide a session snapshot when you want to simulate access through
              an already assumed role.
            </p>
            <p class="mt-1 text-xs text-gray-500">
              Filling an assumed role id enables the assumed-role branch for this
              simulation.
            </p>
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <InputField
              label="Assumed role id"
              value={trust.simAssumedRoleId()}
              placeholder="7"
              onInput={(e) => trust.setSimAssumedRoleId(e.currentTarget.value)}
            />
            <SelectField
              label="Assumed role scope"
              value={trust.simAssumedRoleScope()}
              options={trust.policyScopeOptions}
              onChange={(e) => trust.setSimAssumedRoleScope(e.currentTarget.value)}
            />
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <InputField
              label="Assumed role name"
              value={trust.simAssumedRoleName()}
              placeholder="tenant_admin"
              onInput={(e) => trust.setSimAssumedRoleName(e.currentTarget.value)}
            />
            <InputField
              label="Assumed role tenant"
              value={trust.simAssumedRoleTenantId()}
              placeholder="t_demo"
              onInput={(e) => trust.setSimAssumedRoleTenantId(e.currentTarget.value)}
            />
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <InputField
              label="Session name"
              value={trust.simAssumedRoleSessionName()}
              placeholder="ops-review"
              onInput={(e) => trust.setSimAssumedRoleSessionName(e.currentTarget.value)}
            />
            <InputField
              label="Source identity"
              value={trust.simAssumedRoleSourceIdentity()}
              placeholder="backoffice-admin"
              onInput={(e) => trust.setSimAssumedRoleSourceIdentity(e.currentTarget.value)}
            />
          </div>
          <div class="grid gap-3 md:grid-cols-2">
            <InputField
              label="Service principal"
              value={trust.simAssumedRoleServicePrincipal()}
              placeholder="backoffice.podzone.internal"
              onInput={(e) => trust.setSimAssumedRoleServicePrincipal(e.currentTarget.value)}
            />
            <InputField
              label="Expires at"
              value={trust.simAssumedRoleExpiresAt()}
              placeholder="2026-05-19T18:30:00Z"
              onInput={(e) => trust.setSimAssumedRoleExpiresAt(e.currentTarget.value)}
            />
          </div>
        </Card>
      </div>
      <Button size="sm" color="dark" onClick={trust.handleSimulate}>
        Simulate access
      </Button>

      <Show
        when={trust.simulation()}
        fallback={
          <EmptyBlock
            title="No simulation yet"
            copy="Run a simulation to inspect why a request is allowed or denied across identity, boundaries, session policy, and SCP layers."
          />
        }
      >
        {(result) => (
          <div class="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4">
            <div class="flex flex-wrap items-center gap-3">
              <Badge
                content={result().allowed ? 'allowed' : 'denied'}
                color={result().allowed ? 'green' : 'red'}
              />
              <Badge
                content={result().decisionSource}
                color={trust.simulationSourceColor(result().decisionSource)}
              />
              <Badge
                content={`${result().layers?.length || 0} layers`}
                color="dark"
              />
              <Badge
                content={`${result().matchedStatements?.length || 0} top matches`}
                color="blue"
              />
            </div>
            <p class="text-sm text-gray-600">{result().reason}</p>
            <Show when={(result().matchedStatements || []).length > 0}>
              <div class="rounded-lg border border-gray-200 bg-white p-4">
                <div class="flex flex-wrap items-center gap-2">
                  <Badge content="decision matches" color="dark" />
                  <Badge
                    content={
                      result().matchedStatements?.some(
                        (statement) => statement.effect.toLowerCase() === 'deny'
                      )
                        ? 'explicit deny present'
                        : 'allow path'
                    }
                    color={
                      result().matchedStatements?.some(
                        (statement) => statement.effect.toLowerCase() === 'deny'
                      )
                        ? 'red'
                        : 'green'
                    }
                  />
                </div>
                <div class="mt-3 space-y-2">
                  <For each={result().matchedStatements || []}>
                    {(statement) => (
                      <div class="rounded-md bg-gray-50 p-3 text-xs text-gray-600">
                        <div class="flex flex-wrap items-center gap-2">
                          <Badge
                            content={statement.effect}
                            color={
                              statement.effect.toLowerCase() === 'deny'
                                ? 'red'
                                : 'green'
                            }
                          />
                          <Badge
                            content={trust.statementSourceLabel(statement.source)}
                            color={trust.simulationSourceColor(statement.source)}
                          />
                          <Show when={statement.policyName}>
                            <Badge
                              content={statement.policyName || 'inline'}
                              color="dark"
                            />
                          </Show>
                        </div>
                        <p class="mt-2">
                          {statement.actionPattern} on {statement.resourcePattern}
                        </p>
                        <Show when={(statement.conditions || []).length > 0}>
                          <p class="mt-2 text-[11px] text-gray-500">
                            Conditions:{' '}
                            {(statement.conditions || [])
                              .map(
                                (condition) =>
                                  `${condition.operator} ${condition.key}=${condition.value}`
                              )
                              .join(' · ')}
                          </p>
                        </Show>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </Show>
            <div class="space-y-3">
              <For each={result().layers || []}>
                {(layer) => (
                  <div
                    class={classes(
                      'rounded-lg border p-4',
                      trust.simulationLayerTone(layer.allowed, layer.reason)
                    )}
                  >
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge
                        content={layer.layer}
                        color={trust.simulationSourceColor(layer.layer)}
                      />
                      <Badge
                        content={layer.allowed ? 'allowed' : 'denied'}
                        color={layer.allowed ? 'green' : 'red'}
                      />
                      <Show when={layer.reason.toLowerCase().includes('deny')}>
                        <Badge content="explicit deny" color="red" />
                      </Show>
                      <Show when={layer.reason.toLowerCase().includes('boundary')}>
                        <Badge content="boundary gate" color="pink" />
                      </Show>
                      <Show when={layer.reason.toLowerCase().includes('scp')}>
                        <Badge content="scp gate" color="yellow" />
                      </Show>
                      <Show
                        when={layer.reason.toLowerCase().includes('session policy')}
                      >
                        <Badge content="session scope-down" color="indigo" />
                      </Show>
                    </div>
                    <p class="mt-2 text-sm text-gray-600">{layer.reason}</p>
                    <Show when={(layer.matchedStatements || []).length > 0}>
                      <div class="mt-3 space-y-2">
                        <For each={layer.matchedStatements || []}>
                          {(statement) => (
                            <div class="rounded-md bg-gray-50 p-3 text-xs text-gray-600">
                              <div class="flex flex-wrap items-center gap-2">
                                <Badge
                                  content={statement.effect}
                                  color={
                                    statement.effect.toLowerCase() === 'deny'
                                      ? 'red'
                                      : 'green'
                                  }
                                />
                                <Badge
                                  content={trust.statementSourceLabel(statement.source)}
                                  color={trust.simulationSourceColor(statement.source)}
                                />
                                <Badge
                                  content={statement.policyName || 'inline'}
                                  color="dark"
                                />
                              </div>
                              <p class="mt-1">
                                {statement.actionPattern} on {statement.resourcePattern}
                              </p>
                              <Show when={(statement.conditions || []).length > 0}>
                                <p class="mt-2 text-[11px] text-gray-500">
                                  Conditions:{' '}
                                  {(statement.conditions || [])
                                    .map(
                                      (condition) =>
                                        `${condition.operator} ${condition.key}=${condition.value}`
                                    )
                                    .join(' · ')}
                                </p>
                              </Show>
                            </div>
                          )}
                        </For>
                      </div>
                    </Show>
                  </div>
                )}
              </For>
            </div>
          </div>
        )}
      </Show>
    </>
  );
}
