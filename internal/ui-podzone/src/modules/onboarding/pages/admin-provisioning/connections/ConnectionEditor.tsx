import { createEffect } from 'solid-js'
import { Drawer } from '@/solid/components/common/Overlay'
import { Button } from '@/solid/components/common/Primitives'
import {
    createFormStore,
    FormInputField,
    FormSelectField,
    FormTextareaField,
    jsonObject,
    required,
} from '@/solid/forms'
import type { InfrastructureConnection, UpsertInfrastructureConnection } from '@/services/onboarding/provisioning'
import type { ConnectionsViewModel } from './createConnectionsViewModel'

type ConnectionForm = {
    infraType: string
    name: string
    endpoint: string
    secretRef: string
    status: string
    clusterName: string
    mode: string
    dbName: string
    schemaName: string
    metaJson: string
    configJson: string
}

function initialValues(connection?: InfrastructureConnection): ConnectionForm {
    return {
        infraType: connection?.infra_type || 'postgres',
        name: connection?.name || 'primary',
        endpoint: connection?.endpoint || '',
        secretRef: connection?.secret_ref || '',
        status: connection?.status || 'active',
        clusterName: String(connection?.meta?.cluster_name || ''),
        mode: String(connection?.meta?.mode || 'schema'),
        dbName: String(connection?.meta?.db_name || ''),
        schemaName: String(connection?.meta?.schema_name || ''),
        metaJson: JSON.stringify(connection?.meta || {}, null, 2),
        configJson: JSON.stringify(connection?.config || {}, null, 2),
    }
}

export function ConnectionEditor(props: {
    open: boolean
    connection?: InfrastructureConnection
    vm: ConnectionsViewModel
}) {
    const form = createFormStore<ConnectionForm>({
        initialValues: initialValues(),
        validators: {
            infraType: [required()],
            name: [required()],
            endpoint: [required()],
            metaJson: [jsonObject('Metadata must be a JSON object.')],
            configJson: [jsonObject('Configuration must be a JSON object.')],
        },
    })

    createEffect(() => {
        if (props.open) form.reset(initialValues(props.connection))
    })

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        if (!form.validate()) return
        const request: UpsertInfrastructureConnection = {
            infra_type: form.values.infraType.trim(),
            name: form.values.name.trim(),
            endpoint: form.values.endpoint.trim(),
            secret_ref: form.values.secretRef.trim(),
            status: form.values.status.trim(),
            cluster_name: form.values.clusterName.trim(),
            mode: form.values.mode.trim(),
            db_name: form.values.dbName.trim(),
            schema_name: form.values.schemaName.trim(),
            meta: JSON.parse(form.values.metaJson),
            config: JSON.parse(form.values.configJson),
        }
        await props.vm.save(request)
    }

    return (
        <Drawer
            open={props.open}
            title={`${props.connection ? 'Edit' : 'Add'} tenant connection`}
            class="max-w-2xl"
            onClose={props.vm.closeEditor}
        >
            <form class="space-y-5" onSubmit={submit}>
                <div class="grid gap-4 sm:grid-cols-2">
                    <FormSelectField
                        form={form}
                        name="infraType"
                        label="Infrastructure type"
                        options={[
                            { name: 'PostgreSQL', value: 'postgres' },
                            { name: 'MongoDB', value: 'mongo' },
                            { name: 'Redis', value: 'redis' },
                            { name: 'Kafka', value: 'kafka' },
                            { name: 'HTTP service', value: 'http' },
                        ]}
                    />
                    <FormInputField
                        form={form}
                        name="name"
                        label="Connection key"
                        disabled={Boolean(props.connection)}
                    />
                </div>
                <FormInputField
                    form={form}
                    name="endpoint"
                    label="Endpoint"
                    placeholder="postgres://database.internal:5432"
                />
                <div class="grid gap-4 sm:grid-cols-2">
                    <FormInputField
                        form={form}
                        name="secretRef"
                        label="Secret reference"
                        placeholder="vault://path/to/secret"
                    />
                    <FormSelectField
                        form={form}
                        name="status"
                        label="Status"
                        options={[
                            { name: 'Active', value: 'active' },
                            { name: 'Disabled', value: 'disabled' },
                            { name: 'Maintenance', value: 'maintenance' },
                        ]}
                    />
                </div>
                <div class="border-t border-gray-200 pt-5">
                    <p class="mb-4 text-sm font-semibold text-gray-950">PostgreSQL placement route</p>
                    <div class="grid gap-4 sm:grid-cols-2">
                        <FormInputField form={form} name="clusterName" label="Cluster" />
                        <FormSelectField
                            form={form}
                            name="mode"
                            label="Isolation mode"
                            options={[
                                { name: 'Schema', value: 'schema' },
                                { name: 'Database', value: 'database' },
                            ]}
                        />
                        <FormInputField form={form} name="dbName" label="Database" />
                        <FormInputField form={form} name="schemaName" label="Schema" />
                    </div>
                </div>
                <div class="grid gap-4 lg:grid-cols-2">
                    <FormTextareaField form={form} name="metaJson" label="Metadata JSON" rows={10} />
                    <FormTextareaField form={form} name="configJson" label="Driver configuration JSON" rows={10} />
                </div>
                <div class="flex justify-end gap-3">
                    <Button color="alternative" onClick={props.vm.closeEditor}>
                        Cancel
                    </Button>
                    <Button type="submit" loading={props.vm.saving()}>
                        Save connection
                    </Button>
                </div>
            </form>
        </Drawer>
    )
}
