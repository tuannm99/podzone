import { For, Show } from 'solid-js'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@/solid/components/common/DataTable'
import { EmptyBlock } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button } from '@/solid/components/common/Primitives'
import { CollectionControls } from '@/solid/components/common/CollectionControls'
import { useAdminIamPrincipal } from './context'

export function PrincipalManagedPoliciesTable() {
    const principal = useAdminIamPrincipal()

    return (
        <div class="space-y-3">
            <CollectionControls
                query={principal.managedPoliciesQuery}
                loading={principal.managedPoliciesLoading}
                error={principal.managedPoliciesError}
                searchPlaceholder="Search managed policy"
                sortOptions={[
                    { label: 'Created', value: 'createdAt' },
                    { label: 'Name', value: 'name' },
                    { label: 'Scope', value: 'scope' },
                ]}
                filterFields={[
                    {
                        label: 'Scope',
                        value: 'scope',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                    },
                    {
                        label: 'Name',
                        value: 'name',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
                    },
                ]}
                updateQuery={principal.updateManagedPoliciesQuery}
            />
            <Show
                when={principal.currentManagedPolicies().length > 0}
                fallback={
                    <EmptyBlock title="No direct policies" copy="No managed policies are attached to this principal." />
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Policy</TableHeaderCell>
                            <TableHeaderCell>Scope</TableHeaderCell>
                            <TableHeaderCell class="text-right">Action</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={principal.currentManagedPolicies()}>
                            {(policy) => (
                                <TableRow>
                                    <TableCell class="font-medium text-gray-900">{policy.name}</TableCell>
                                    <TableCell>
                                        <Badge content={policy.scope} color="blue" />
                                    </TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            size="xs"
                                            color="red"
                                            onClick={() => principal.handleDetachPrincipalManagedPolicy(policy.name)}
                                        >
                                            Detach
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={principal.managedPoliciesPageInfo().page}
                    pageSize={principal.managedPoliciesPageInfo().pageSize}
                    total={principal.managedPoliciesPageInfo().total}
                    loading={principal.managedPoliciesLoading()}
                    onPageChange={(page) => principal.updateManagedPoliciesQuery({ page })}
                />
            </Show>
        </div>
    )
}

export function PrincipalInlinePoliciesTable() {
    const principal = useAdminIamPrincipal()

    return (
        <div class="space-y-3">
            <CollectionControls
                query={principal.inlinePoliciesQuery}
                loading={principal.inlinePoliciesLoading}
                error={principal.inlinePoliciesError}
                searchPlaceholder="Search inline policy"
                sortOptions={[
                    { label: 'Created', value: 'createdAt' },
                    { label: 'Updated', value: 'updatedAt' },
                    { label: 'Name', value: 'name' },
                ]}
                filterFields={[
                    {
                        label: 'Name',
                        value: 'name',
                        operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
                    },
                ]}
                updateQuery={principal.updateInlinePoliciesQuery}
            />
            <Show
                when={principal.currentInlinePolicies().length > 0}
                fallback={
                    <EmptyBlock title="No inline policies" copy="No inline policies are attached to this principal." />
                }
            >
                <DataTable>
                    <TableHead>
                        <TableRow>
                            <TableHeaderCell>Policy</TableHeaderCell>
                            <TableHeaderCell>Description</TableHeaderCell>
                            <TableHeaderCell>Statements</TableHeaderCell>
                            <TableHeaderCell class="text-right">Action</TableHeaderCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        <For each={principal.currentInlinePolicies()}>
                            {(policy) => (
                                <TableRow>
                                    <TableCell class="font-medium text-gray-900">{policy.name}</TableCell>
                                    <TableCell class="text-gray-600">
                                        {policy.description || 'No description'}
                                    </TableCell>
                                    <TableCell class="text-gray-600">{policy.statements?.length || 0}</TableCell>
                                    <TableCell class="text-right">
                                        <Button
                                            size="xs"
                                            color="red"
                                            onClick={() => principal.handleDeletePrincipalInlinePolicy(policy.name)}
                                        >
                                            Delete
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            )}
                        </For>
                    </TableBody>
                </DataTable>
                <Pagination
                    page={principal.inlinePoliciesPageInfo().page}
                    pageSize={principal.inlinePoliciesPageInfo().pageSize}
                    total={principal.inlinePoliciesPageInfo().total}
                    loading={principal.inlinePoliciesLoading()}
                    onPageChange={(page) => principal.updateInlinePoliciesQuery({ page })}
                />
            </Show>
        </div>
    )
}
