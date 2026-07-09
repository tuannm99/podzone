import { For, Show } from 'solid-js'
import {
    DataTable,
    TableBody,
    TableCell,
    TableHead,
    TableHeaderCell,
    TableRow,
} from '@podzone/shared/ui/components/common/DataTable'
import { EmptyBlock } from '@podzone/shared/ui/components/common/Feedback'
import { Pagination } from '@podzone/shared/ui/components/common/Pagination'
import { Badge, Button } from '@podzone/shared/ui/components/common/Primitives'
import { CollectionControls } from '@podzone/shared/ui/components/common/CollectionControls'
import { useAdminIamPolicy } from './context'

export function PolicyDetailTables() {
    const policy = useAdminIamPolicy()

    return (
        <div class="grid gap-6 xl:grid-cols-2">
            <section class="min-w-0 space-y-3">
                <p class="text-sm font-semibold text-gray-900">Versions</p>
                <CollectionControls
                    query={policy.policyVersionsQuery}
                    loading={policy.policyVersionsLoading}
                    error={policy.policyVersionsError}
                    searchPlaceholder="Search version"
                    sortOptions={[
                        { label: 'Created', value: 'createdAt' },
                        { label: 'Version', value: 'version' },
                    ]}
                    filterFields={[
                        {
                            label: 'Version',
                            value: 'version',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_CONTAINS'],
                        },
                        {
                            label: 'Default',
                            value: 'isDefault',
                            operators: ['FILTER_OPERATOR_EQ'],
                        },
                    ]}
                    updateQuery={policy.updatePolicyVersionsQuery}
                />
                <Show
                    when={policy.policyVersions().length > 0}
                    fallback={<EmptyBlock title="No versions" copy="Select a policy to inspect its version history." />}
                >
                    <DataTable>
                        <TableHead>
                            <TableRow>
                                <TableHeaderCell>Version</TableHeaderCell>
                                <TableHeaderCell>Created</TableHeaderCell>
                                <TableHeaderCell class="text-right">Actions</TableHeaderCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <For each={policy.policyVersions()}>
                                {(version) => (
                                    <TableRow>
                                        <TableCell class="font-semibold text-gray-900">{version.version}</TableCell>
                                        <TableCell class="text-gray-600">{version.createdAt || 'Unknown'}</TableCell>
                                        <TableCell>
                                            <div class="flex justify-end gap-2">
                                                <Show
                                                    when={version.isDefault}
                                                    fallback={
                                                        <>
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
                                                        </>
                                                    }
                                                >
                                                    <Badge content="default" color="green" />
                                                </Show>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                )}
                            </For>
                        </TableBody>
                    </DataTable>
                    <Pagination
                        page={policy.policyVersionsPageInfo().page}
                        pageSize={policy.policyVersionsPageInfo().pageSize}
                        total={policy.policyVersionsPageInfo().total}
                        loading={policy.policyVersionsLoading()}
                        onPageChange={(page) => policy.updatePolicyVersionsQuery({ page })}
                    />
                </Show>
            </section>

            <section class="min-w-0 space-y-3">
                <p class="text-sm font-semibold text-gray-900">Attachments</p>
                <CollectionControls
                    query={policy.policyAttachmentsQuery}
                    loading={policy.policyAttachmentsLoading}
                    error={policy.policyAttachmentsError}
                    searchPlaceholder="Search type, scope, or principal"
                    sortOptions={[
                        { label: 'Created', value: 'createdAt' },
                        { label: 'Type', value: 'attachmentType' },
                        { label: 'Scope', value: 'scope' },
                    ]}
                    filterFields={[
                        {
                            label: 'Type',
                            value: 'attachmentType',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                        },
                        {
                            label: 'Scope',
                            value: 'scope',
                            operators: ['FILTER_OPERATOR_EQ', 'FILTER_OPERATOR_IN'],
                        },
                        {
                            label: 'Tenant ID',
                            value: 'tenantId',
                            operators: ['FILTER_OPERATOR_EQ'],
                        },
                    ]}
                    updateQuery={policy.updatePolicyAttachmentsQuery}
                />
                <Show
                    when={policy.policyAttachments().length > 0}
                    fallback={
                        <EmptyBlock
                            title="No attachments"
                            copy="This policy is not attached to another IAM resource."
                        />
                    }
                >
                    <DataTable>
                        <TableHead>
                            <TableRow>
                                <TableHeaderCell>Type</TableHeaderCell>
                                <TableHeaderCell>Principal</TableHeaderCell>
                                <TableHeaderCell>Scope</TableHeaderCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <For each={policy.policyAttachments()}>
                                {(attachment) => (
                                    <TableRow>
                                        <TableCell>
                                            <Badge
                                                content={attachment.attachmentType}
                                                color={policy.attachmentColor(attachment.attachmentType)}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <span class="block text-gray-700">
                                                {attachment.roleName ||
                                                    attachment.groupName ||
                                                    (attachment.userId
                                                        ? policy.identityForUser(attachment.userId).label
                                                        : 'Organization')}
                                            </span>
                                            <Show when={attachment.userId}>
                                                {(userID) => (
                                                    <span class="block text-xs text-gray-500">
                                                        {policy.identityForUser(userID()).description}
                                                    </span>
                                                )}
                                            </Show>
                                        </TableCell>
                                        <TableCell class="text-gray-600">
                                            {attachment.tenantId || attachment.scope || 'platform'}
                                        </TableCell>
                                    </TableRow>
                                )}
                            </For>
                        </TableBody>
                    </DataTable>
                    <Pagination
                        page={policy.policyAttachmentsPageInfo().page}
                        pageSize={policy.policyAttachmentsPageInfo().pageSize}
                        total={policy.policyAttachmentsPageInfo().total}
                        loading={policy.policyAttachmentsLoading()}
                        onPageChange={(page) => policy.updatePolicyAttachmentsQuery({ page })}
                    />
                </Show>
            </section>
        </div>
    )
}
