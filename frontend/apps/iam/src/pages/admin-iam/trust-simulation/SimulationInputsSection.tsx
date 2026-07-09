import { Show } from 'solid-js'
import { InfoAlert } from '@/solid/components/common/Feedback'
import { IamKeyValueBuilder } from '@/modules/iam/components'
import { IamStatementBuilder } from '../shared/IamStatementBuilder'
import { Card, InputField, SelectField } from '@/solid/components/common/Primitives'
import { SearchSelectField } from '@/solid/components/common/SearchSelectField'
import { useAdminIamTrustSim } from './context'
import { SimulationPresetButtons } from './SimulationPresetButtons'

export function SimulationInputsSection() {
    const trust = useAdminIamTrustSim()
    const simulationActionOptions = () => [
        { name: 'Choose a permission', value: '' },
        ...trust.permissionOptions(),
        { name: 'Custom action', value: '__custom__' },
    ]
    const simulationActionValue = () => {
        if (!trust.simAction()) return ''
        return trust.permissionOptions().some((option) => option.value === trust.simAction())
            ? trust.simAction()
            : '__custom__'
    }

    return (
        <>
            <div class="grid gap-3 md:grid-cols-2">
                <SelectField
                    label="Simulation scope"
                    value={trust.simScope()}
                    options={trust.policyScopeOptions}
                    onChange={(e) => trust.setSimScope(e.currentTarget.value)}
                />
                <SearchSelectField
                    label="Target user"
                    value={trust.simTargetUserId()}
                    options={trust.simScope() === 'tenant' ? trust.tenantUserOptions() : trust.platformUserOptions()}
                    loading={trust.simScope() === 'tenant' ? trust.tenantUsersLoading() : trust.platformUsersLoading()}
                    error={trust.simScope() === 'tenant' ? trust.tenantUsersError() : trust.platformUsersError()}
                    onSearch={trust.simScope() === 'tenant' ? trust.searchTenantUsers : trust.searchPlatformUsers}
                    onChange={trust.setSimTargetUserId}
                    placeholder="Search name, username, or email"
                />
            </div>
            <InfoAlert>
                Simulation evaluates identity policies, group policies, trust, boundaries, session scope-down, and SCP
                layers together.
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
                <div class="space-y-3">
                    <SelectField
                        label="Permission"
                        value={simulationActionValue()}
                        options={simulationActionOptions()}
                        onChange={(event) =>
                            trust.setSimAction(
                                event.currentTarget.value === '__custom__' ? '' : event.currentTarget.value
                            )
                        }
                    />
                    <Show when={simulationActionValue() === '__custom__'}>
                        <InputField
                            label="Custom action"
                            value={trust.simAction()}
                            onInput={(e) => trust.setSimAction(e.currentTarget.value)}
                        />
                    </Show>
                </div>
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
            <SimulationPresetButtons />
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
                    actionOptions={trust.permissionOptions()}
                    value={trust.simSessionPolicyJson()}
                    onChange={trust.setSimSessionPolicyJson}
                />
                <AssumedRoleSessionCard />
            </div>
        </>
    )
}

function AssumedRoleSessionCard() {
    const trust = useAdminIamTrustSim()

    return (
        <Card class="space-y-4 border border-gray-200 bg-gray-50 p-4 shadow-none">
            <div>
                <p class="text-sm font-medium text-gray-700">Assumed role session</p>
                <p class="mt-1 text-xs text-gray-500">
                    Provide a session snapshot when you want to simulate access through an already assumed role.
                </p>
                <p class="mt-1 text-xs text-gray-500">
                    Filling an assumed role id enables the assumed-role branch for this simulation.
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
    )
}
