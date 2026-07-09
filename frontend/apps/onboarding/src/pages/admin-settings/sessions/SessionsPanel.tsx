import { For, Show } from 'solid-js'
import { CollectionFilters } from '@/solid/components/common/CollectionFilters'
import { CollectionToolbar } from '@/solid/components/common/CollectionToolbar'
import { EmptyBlock, ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback'
import { Pagination } from '@/solid/components/common/Pagination'
import { Badge, Button, Card } from '@/solid/components/common/Primitives'
import { SectionTitle } from '@/solid/components/common/SectionTitle'
import { useAdminSettings } from '../context'
import { sessionStatusColor } from '../presentation'
import { sessionFilterFields, sessionSortOptions } from './sessions.collection'

export function SessionsPanel() {
    const { sessions, user } = useAdminSettings()

    return (
        <Card class="space-y-4">
            <SectionTitle
                title="Sessions"
                subtitle="Review active sign-ins and revoke sessions that should no longer access your workspaces."
            />
            <div class="flex flex-wrap gap-3">
                <Badge content={`current ${sessions.currentCount()}`} color="yellow" />
                <Badge content={`other ${sessions.otherCount()}`} color="indigo" />
                <Button color="alternative" onClick={() => void sessions.reload()}>
                    Reload sessions
                </Button>
            </div>
            <CollectionToolbar
                search={sessions.query.search || ''}
                searchPlaceholder="Session, workspace, status, or role"
                sortBy={sessions.query.sortBy || 'created_at'}
                sortDirection={sessions.query.sortDirection || 'SORT_DIRECTION_DESC'}
                pageSize={sessions.query.pageSize}
                sortOptions={sessionSortOptions}
                onSearch={(search) => sessions.updateQuery({ search })}
                onSortByChange={(sortBy) => sessions.updateQuery({ sortBy })}
                onSortDirectionChange={(sortDirection) => sessions.updateQuery({ sortDirection })}
                onPageSizeChange={(pageSize) => sessions.updateQuery({ pageSize })}
            />
            <CollectionFilters
                fields={sessionFilterFields}
                filters={sessions.query.filters || []}
                onChange={(filters) => sessions.updateQuery({ filters })}
            />
            <Show when={sessions.loading()}>
                <LoadingInline label="Loading sessions..." />
            </Show>
            <Show when={sessions.error()}>
                <ErrorAlert>{sessions.error()}</ErrorAlert>
            </Show>
            <Show when={sessions.message()}>
                <InfoAlert>{sessions.message()}</InfoAlert>
            </Show>
            <div class="min-h-48">
                <Show
                    when={!sessions.error() && sessions.items().length > 0}
                    fallback={
                        <Show when={!sessions.loading() && !sessions.error()}>
                            <EmptyBlock
                                title="No sessions loaded"
                                copy="Signed-in sessions will appear here once this account starts using the backoffice."
                            />
                        </Show>
                    }
                >
                    <div class="max-h-[28rem] space-y-3 overflow-y-auto pr-1">
                        <For each={sessions.items()}>
                            {(session) => (
                                <div class="rounded-lg border border-gray-200 p-4">
                                    <div class="flex flex-wrap items-center justify-between gap-3">
                                        <div>
                                            <p class="break-all font-semibold text-gray-900">{session.id}</p>
                                            <p class="mt-1 text-sm text-gray-500">
                                                workspace {session.activeTenantId || 'not selected'} · expires{' '}
                                                {session.expiresAt || 'unknown'}
                                            </p>
                                        </div>
                                        <div class="flex flex-wrap items-center gap-2">
                                            <Badge
                                                content={session.status || 'unknown'}
                                                color={sessionStatusColor(session.status)}
                                            />
                                            <Show when={session.id === user.sessionID()}>
                                                <Badge content="current" color="yellow" />
                                            </Show>
                                            <Button
                                                color="red"
                                                size="xs"
                                                onClick={() => void sessions.revoke(session.id)}
                                            >
                                                Revoke
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            )}
                        </For>
                    </div>
                </Show>
            </div>
            <Pagination
                page={sessions.query.page}
                pageSize={sessions.query.pageSize}
                total={sessions.pageInfo().total}
                loading={sessions.loading()}
                onPageChange={(page) => sessions.updateQuery({ page })}
            />
        </Card>
    )
}
