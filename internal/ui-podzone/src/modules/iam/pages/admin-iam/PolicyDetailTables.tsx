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
import { createClientPagination } from '@/solid/pagination'
import { useAdminIamPolicy } from './policy-context'

export function PolicyDetailTables() {
  const policy = useAdminIamPolicy()
  const versionsPage = createClientPagination(policy.policyVersions, 6)
  const attachmentsPage = createClientPagination(policy.policyAttachments, 6)

  return (
    <div class="grid gap-6 xl:grid-cols-2">
      <section class="min-w-0 space-y-3">
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
          <DataTable>
            <TableHead>
              <TableRow>
                <TableHeaderCell>Version</TableHeaderCell>
                <TableHeaderCell>Created</TableHeaderCell>
                <TableHeaderCell class="text-right">Actions</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <For each={versionsPage.pageItems()}>
                {(version) => (
                  <TableRow>
                    <TableCell class="font-semibold text-gray-900">
                      {version.version}
                    </TableCell>
                    <TableCell class="text-gray-600">
                      {version.createdAt || 'Unknown'}
                    </TableCell>
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
                                  policy.handleSetDefaultVersion(
                                    version.version
                                  )
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
            page={versionsPage.page()}
            pageSize={versionsPage.pageSize}
            total={versionsPage.total()}
            onPageChange={versionsPage.setPage}
          />
        </Show>
      </section>

      <section class="min-w-0 space-y-3">
        <p class="text-sm font-semibold text-gray-900">Attachments</p>
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
              <For each={attachmentsPage.pageItems()}>
                {(attachment) => (
                  <TableRow>
                    <TableCell>
                      <Badge
                        content={attachment.attachmentType}
                        color={policy.attachmentColor(
                          attachment.attachmentType
                        )}
                      />
                    </TableCell>
                    <TableCell class="text-gray-700">
                      {attachment.roleName ||
                        attachment.groupName ||
                        (attachment.userId
                          ? `User ${attachment.userId}`
                          : 'Organization')}
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
            page={attachmentsPage.page()}
            pageSize={attachmentsPage.pageSize}
            total={attachmentsPage.total()}
            onPageChange={attachmentsPage.setPage}
          />
        </Show>
      </section>
    </div>
  )
}
