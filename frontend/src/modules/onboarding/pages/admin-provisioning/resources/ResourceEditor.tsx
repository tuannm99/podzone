import { createEffect } from 'solid-js'
import { ErrorAlert } from '@/solid/components/common/Feedback'
import { Drawer } from '@/solid/components/common/Overlay'
import { Button } from '@/solid/components/common/Primitives'
import { createFormStore, FormInputField, FormTextareaField, jsonObject, required } from '@/solid/forms'
import type {
    DatabaseClusterResource,
    KubernetesClusterResource,
    RuntimePoolResource,
} from '@/services/onboarding/provisioning'
import type { ResourceEditor as EditorState, ResourcesViewModel } from './createResourcesViewModel'

type ResourceDocument = DatabaseClusterResource | KubernetesClusterResource | RuntimePoolResource

type EditorForm = {
    name: string
    document: string
}

const labels: Record<EditorState['kind'], string> = {
    'database-clusters': 'database cluster',
    'kubernetes-clusters': 'Kubernetes cluster',
    'runtime-pools': 'runtime pool',
}

function defaultDocument(kind: EditorState['kind']): ResourceDocument {
    const timestamps = { created_at: '', updated_at: '' }
    if (kind === 'database-clusters') {
        return {
            name: '',
            engine: 'postgres',
            region: '',
            placement_db: 'podzone_tenants',
            max_tenants: 100,
            current_tenants: 0,
            max_schemas: 100,
            current_schemas: 0,
            max_connections: 500,
            current_connections: 0,
            status: 'active',
            healthy: true,
            ...timestamps,
        }
    }
    if (kind === 'kubernetes-clusters') {
        return {
            name: '',
            region: '',
            namespaces: [
                {
                    name: 'default',
                    max_tenants: 50,
                    current_tenants: 0,
                    cpu_milli: 1000,
                    memory_mi: 2048,
                    status: 'active',
                    healthy: true,
                },
            ],
            status: 'active',
            healthy: true,
            ...timestamps,
        }
    }
    return {
        name: '',
        kind: 'docker',
        max_tenants: 50,
        current_tenants: 0,
        status: 'active',
        healthy: true,
        ...timestamps,
    }
}

function editableDocument(editor: EditorState) {
    const value = editor.value || defaultDocument(editor.kind)
    const config = Object.fromEntries(
        Object.entries(value).filter(([key]) => !['name', 'created_at', 'updated_at'].includes(key))
    )
    return JSON.stringify(config, null, 2)
}

export function ResourceEditor(props: { editor?: EditorState; vm: ResourcesViewModel }) {
    const form = createFormStore<EditorForm>({
        initialValues: { name: '', document: '{}' },
        validators: {
            name: [required('Resource name is required.')],
            document: [jsonObject('Configuration must be a JSON object.')],
        },
    })

    createEffect(() => {
        const editor = props.editor
        if (!editor) return
        form.reset({
            name: editor.value?.name || '',
            document: editableDocument(editor),
        })
    })

    const submit = async (event: SubmitEvent) => {
        event.preventDefault()
        const editor = props.editor
        if (!editor || !form.validate()) return
        const resource = {
            ...JSON.parse(form.values.document),
            name: form.values.name.trim(),
        } as ResourceDocument
        if (editor.kind === 'database-clusters') {
            await props.vm.saveDatabaseCluster(resource as DatabaseClusterResource)
        } else if (editor.kind === 'kubernetes-clusters') {
            await props.vm.saveKubernetesCluster(resource as KubernetesClusterResource)
        } else {
            await props.vm.saveRuntimePool(resource as RuntimePoolResource)
        }
    }

    return (
        <Drawer
            open={Boolean(props.editor)}
            title={`${props.editor?.value ? 'Edit' : 'Add'} ${props.editor ? labels[props.editor.kind] : 'resource'}`}
            class="max-w-2xl"
            onClose={() => props.vm.setEditor()}
        >
            <form class="space-y-5" onSubmit={submit}>
                <FormInputField
                    form={form}
                    name="name"
                    label="Resource key"
                    disabled={Boolean(props.editor?.value)}
                    placeholder="Unique resource name"
                />
                <FormTextareaField form={form} name="document" label="Configuration document" rows={24} />
                <ErrorAlert>
                    Capacity counters are operational state. Changing them affects the next placement decision
                    immediately.
                </ErrorAlert>
                <div class="flex justify-end gap-3">
                    <Button color="alternative" onClick={() => props.vm.setEditor()}>
                        Cancel
                    </Button>
                    <Button type="submit" loading={props.vm.saving()}>
                        Save resource
                    </Button>
                </div>
            </form>
        </Drawer>
    )
}
