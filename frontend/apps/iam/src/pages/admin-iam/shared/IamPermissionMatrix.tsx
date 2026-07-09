import { For, Show, createMemo, createSignal } from 'solid-js'
import { Badge, InputField } from '@podzone/shared/ui/components/common/Primitives'

export type IamPermissionOption = {
    name: string
    value: string
    resource: string
    action: string
}

export type IamPermissionSelection = {
    selected: boolean
    scoped: boolean
}

type PermissionGroup = {
    resource: string
    permissions: IamPermissionOption[]
}

export function IamPermissionMatrix(props: {
    permissions: IamPermissionOption[]
    selection: (permission: IamPermissionOption) => IamPermissionSelection
    onToggle: (permission: IamPermissionOption, selected: boolean) => void
    onToggleResource: (resource: string, permissions: IamPermissionOption[], selected: boolean) => void
}) {
    const [search, setSearch] = createSignal('')
    const groups = createMemo<PermissionGroup[]>(() => {
        const query = search().trim().toLowerCase()
        const grouped = new Map<string, IamPermissionOption[]>()

        for (const permission of props.permissions) {
            const searchable =
                `${permission.resource} ${permission.action} ${permission.value} ${permission.name}`.toLowerCase()
            if (query && !searchable.includes(query)) continue
            const current = grouped.get(permission.resource) || []
            current.push(permission)
            grouped.set(permission.resource, current)
        }

        return [...grouped.entries()]
            .map(([resource, permissions]) => ({
                resource,
                permissions: permissions.sort((left, right) => left.action.localeCompare(right.action)),
            }))
            .sort((left, right) => left.resource.localeCompare(right.resource))
    })

    const selectedCount = (permissions: IamPermissionOption[]) =>
        permissions.filter((permission) => props.selection(permission).selected).length

    return (
        <div class="space-y-3">
            <div class="max-w-xl">
                <InputField
                    label="Search features and permissions"
                    value={search()}
                    placeholder="Orders, manage roles, tenant:manage_members"
                    onInput={(event) => setSearch(event.currentTarget.value)}
                />
            </div>

            <Show
                when={groups().length > 0}
                fallback={
                    <div class="border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500">
                        No permissions match this search.
                    </div>
                }
            >
                <div class="overflow-x-auto border border-gray-200">
                    <table class="min-w-full divide-y divide-gray-200 text-left text-sm">
                        <thead class="bg-gray-50 text-xs font-semibold uppercase text-gray-500">
                            <tr>
                                <th class="w-56 px-4 py-3">Feature</th>
                                <th class="min-w-[40rem] px-4 py-3">Permissions</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200 bg-white">
                            <For each={groups()}>
                                {(group) => {
                                    const count = () => selectedCount(group.permissions)
                                    const allSelected = () =>
                                        group.permissions.length > 0 && count() === group.permissions.length

                                    return (
                                        <tr class="align-top">
                                            <th class="bg-gray-50/60 px-4 py-4 font-medium text-gray-900">
                                                <label class="flex cursor-pointer items-start gap-3">
                                                    <input
                                                        type="checkbox"
                                                        class="mt-0.5 size-4 rounded border-gray-300 text-gray-950 focus:ring-gray-300"
                                                        checked={allSelected()}
                                                        onChange={(event) =>
                                                            props.onToggleResource(
                                                                group.resource,
                                                                group.permissions,
                                                                event.currentTarget.checked
                                                            )
                                                        }
                                                    />
                                                    <span class="min-w-0">
                                                        <span class="block break-words">{group.resource}</span>
                                                        <span class="mt-1 block text-xs font-normal text-gray-500">
                                                            {count()}/{group.permissions.length} selected
                                                        </span>
                                                    </span>
                                                </label>
                                            </th>
                                            <td class="px-4 py-4">
                                                <div class="grid min-w-[38rem] grid-cols-2 gap-x-5 gap-y-3 xl:grid-cols-3">
                                                    <For each={group.permissions}>
                                                        {(permission) => {
                                                            const selection = () => props.selection(permission)
                                                            return (
                                                                <label class="flex cursor-pointer items-start gap-3">
                                                                    <input
                                                                        type="checkbox"
                                                                        class="mt-0.5 size-4 rounded border-gray-300 text-gray-950 focus:ring-gray-300"
                                                                        checked={selection().selected}
                                                                        onChange={(event) =>
                                                                            props.onToggle(
                                                                                permission,
                                                                                event.currentTarget.checked
                                                                            )
                                                                        }
                                                                    />
                                                                    <span class="min-w-0">
                                                                        <span class="block break-words font-medium text-gray-800">
                                                                            {permission.action}
                                                                        </span>
                                                                        <span class="block break-all text-xs text-gray-500">
                                                                            {permission.value}
                                                                        </span>
                                                                        <Show when={selection().scoped}>
                                                                            <span class="mt-1 block">
                                                                                <Badge
                                                                                    content="Scoped in Builder"
                                                                                    color="yellow"
                                                                                />
                                                                            </span>
                                                                        </Show>
                                                                    </span>
                                                                </label>
                                                            )
                                                        }}
                                                    </For>
                                                </div>
                                            </td>
                                        </tr>
                                    )
                                }}
                            </For>
                        </tbody>
                    </table>
                </div>
            </Show>
        </div>
    )
}
