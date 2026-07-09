import { Show } from 'solid-js'
import { InfoAlert } from '@podzone/shared/ui/components/common/Feedback'
import { IamTrustPolicyBuilder } from '@/modules/iam/components'
import { Badge, Button } from '@podzone/shared/ui/components/common/Primitives'
import { FormInputField, FormSelectField, type FormStore } from '@podzone/shared/ui/forms'
import type { RoleBoundaryFormValues, TrustPolicyFormValues } from './forms'
import { useAdminIamTrustSim } from './context'

export function TrustPolicySection(props: {
    trustPolicyForm: FormStore<TrustPolicyFormValues>
    roleBoundaryForm: FormStore<RoleBoundaryFormValues>
    onSaveTrustPolicy: () => Promise<void>
    onSaveRoleBoundary: () => Promise<void>
}) {
    const trust = useAdminIamTrustSim()

    return (
        <>
            <div class="grid gap-3 md:grid-cols-2">
                <FormInputField form={props.trustPolicyForm} name="roleName" label="Role name" />
                <InfoAlert>
                    Use the trust policy builder to define which user, role, or service principals may assume this role.
                </InfoAlert>
            </div>
            <div class="flex flex-wrap gap-3">
                <Button size="sm" color="light" onClick={trust.loadTrustPolicy}>
                    Load trust policy
                </Button>
                <Button size="sm" onClick={props.onSaveTrustPolicy} loading={props.trustPolicyForm.isSubmitting()}>
                    Save trust policy
                </Button>
            </div>
            <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-end">
                <FormSelectField
                    form={props.roleBoundaryForm}
                    name="policyName"
                    label="Role boundary policy"
                    options={trust.boundaryPolicyOptions()}
                />
                <Button
                    size="sm"
                    color="dark"
                    onClick={props.onSaveRoleBoundary}
                    loading={props.roleBoundaryForm.isSubmitting()}
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
                value={props.trustPolicyForm.values.trustJson}
                onChange={(value: string) => props.trustPolicyForm.setValue('trustJson', value)}
            />
            <Show when={props.trustPolicyForm.hasError('trustJson')}>
                <p class="text-xs font-medium text-red-600">{props.trustPolicyForm.error('trustJson')}</p>
            </Show>
        </>
    )
}
