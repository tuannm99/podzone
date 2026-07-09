import { createSignal, For } from 'solid-js'
import { Button } from './Primitives'

export type CollectionSortOption = {
    label: string
    value: string
}

type CollectionToolbarProps = {
    search: string
    searchPlaceholder: string
    sortBy: string
    sortDirection: 'SORT_DIRECTION_ASC' | 'SORT_DIRECTION_DESC'
    pageSize: number
    sortOptions: CollectionSortOption[]
    onSearch: (value: string) => void
    onSortByChange: (value: string) => void
    onSortDirectionChange: (value: 'SORT_DIRECTION_ASC' | 'SORT_DIRECTION_DESC') => void
    onPageSizeChange: (value: number) => void
}

export function CollectionToolbar(props: CollectionToolbarProps) {
    const [searchDraft, setSearchDraft] = createSignal(props.search)

    const submitSearch = (event: SubmitEvent) => {
        event.preventDefault()
        props.onSearch(searchDraft().trim())
    }

    return (
        <form class="grid gap-3 lg:grid-cols-[minmax(14rem,1fr)_minmax(10rem,0.5fr)_auto_auto]" onSubmit={submitSearch}>
            <label class="space-y-1 text-sm text-gray-600">
                <span class="font-medium text-gray-800">Search</span>
                <div class="flex gap-2">
                    <input
                        type="search"
                        value={searchDraft()}
                        placeholder={props.searchPlaceholder}
                        class="min-w-0 flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm"
                        onInput={(event) => setSearchDraft(event.currentTarget.value)}
                    />
                    <Button type="submit" color="alternative" size="sm">
                        Search
                    </Button>
                </div>
            </label>

            <label class="space-y-1 text-sm text-gray-600">
                <span class="font-medium text-gray-800">Sort</span>
                <select
                    value={props.sortBy}
                    class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                    onChange={(event) => props.onSortByChange(event.currentTarget.value)}
                >
                    <For each={props.sortOptions}>
                        {(option) => <option value={option.value}>{option.label}</option>}
                    </For>
                </select>
            </label>

            <label class="space-y-1 text-sm text-gray-600">
                <span class="font-medium text-gray-800">Direction</span>
                <select
                    value={props.sortDirection}
                    class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                    onChange={(event) =>
                        props.onSortDirectionChange(
                            event.currentTarget.value as 'SORT_DIRECTION_ASC' | 'SORT_DIRECTION_DESC'
                        )
                    }
                >
                    <option value="SORT_DIRECTION_DESC">Descending</option>
                    <option value="SORT_DIRECTION_ASC">Ascending</option>
                </select>
            </label>

            <label class="space-y-1 text-sm text-gray-600">
                <span class="font-medium text-gray-800">Rows</span>
                <select
                    value={props.pageSize}
                    class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                    onChange={(event) => props.onPageSizeChange(Number.parseInt(event.currentTarget.value, 10))}
                >
                    <For each={[5, 10, 20, 50]}>{(size) => <option value={size}>{size}</option>}</For>
                </select>
            </label>
        </form>
    )
}
