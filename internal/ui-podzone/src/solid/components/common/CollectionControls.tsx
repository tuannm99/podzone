import { Show, type Accessor } from 'solid-js'
import type { CollectionQuery } from '@/services/collection'
import { CollectionFilters, type CollectionFilterField } from './CollectionFilters'
import { CollectionToolbar, type CollectionSortOption } from './CollectionToolbar'
import { ErrorAlert, LoadingInline } from './Feedback'

type CollectionControlsProps = {
    query: CollectionQuery
    loading: Accessor<boolean>
    error: Accessor<string>
    searchPlaceholder: string
    sortOptions: CollectionSortOption[]
    filterFields: CollectionFilterField[]
    updateQuery: (patch: Partial<CollectionQuery>) => void
}

export function CollectionControls(props: CollectionControlsProps) {
    return (
        <div class="space-y-3">
            <CollectionToolbar
                search={props.query.search || ''}
                searchPlaceholder={props.searchPlaceholder}
                sortBy={props.query.sortBy || 'createdAt'}
                sortDirection={props.query.sortDirection || 'SORT_DIRECTION_DESC'}
                pageSize={props.query.pageSize}
                sortOptions={props.sortOptions}
                onSearch={(search) => props.updateQuery({ search })}
                onSortByChange={(sortBy) => props.updateQuery({ sortBy })}
                onSortDirectionChange={(sortDirection) => props.updateQuery({ sortDirection })}
                onPageSizeChange={(pageSize) => props.updateQuery({ pageSize })}
            />
            <CollectionFilters
                fields={props.filterFields}
                filters={props.query.filters || []}
                onChange={(filters) => props.updateQuery({ filters })}
            />
            <Show when={props.error()}>
                <ErrorAlert>{props.error()}</ErrorAlert>
            </Show>
            <Show when={props.loading()}>
                <LoadingInline label="Loading collection..." />
            </Show>
        </div>
    )
}
