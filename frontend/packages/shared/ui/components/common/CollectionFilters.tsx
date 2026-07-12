import { createMemo, createSignal, For, Show } from 'solid-js'
import { Button } from './Primitives'

type FilterOperator =
    | 'FILTER_OPERATOR_EQ'
    | 'FILTER_OPERATOR_NEQ'
    | 'FILTER_OPERATOR_CONTAINS'
    | 'FILTER_OPERATOR_STARTS_WITH'
    | 'FILTER_OPERATOR_GT'
    | 'FILTER_OPERATOR_GTE'
    | 'FILTER_OPERATOR_LT'
    | 'FILTER_OPERATOR_LTE'
    | 'FILTER_OPERATOR_IN'

type CollectionFilter = {
    field: string
    operator: FilterOperator
    values: string[]
}

export type CollectionFilterField = {
    label: string
    value: string
    operators: FilterOperator[]
}

type CollectionFiltersProps = {
    fields: CollectionFilterField[]
    filters: readonly CollectionFilter[]
    onChange: (filters: CollectionFilter[]) => void
}

const operatorLabels: Record<FilterOperator, string> = {
    FILTER_OPERATOR_EQ: 'Equals',
    FILTER_OPERATOR_NEQ: 'Not equal',
    FILTER_OPERATOR_CONTAINS: 'Contains',
    FILTER_OPERATOR_STARTS_WITH: 'Starts with',
    FILTER_OPERATOR_GT: 'Greater than',
    FILTER_OPERATOR_GTE: 'Greater or equal',
    FILTER_OPERATOR_LT: 'Less than',
    FILTER_OPERATOR_LTE: 'Less or equal',
    FILTER_OPERATOR_IN: 'In list',
}

export function CollectionFilters(props: CollectionFiltersProps) {
    const [field, setField] = createSignal(props.fields[0]?.value || '')
    const selectedField = createMemo(() => props.fields.find((item) => item.value === field()))
    const [operator, setOperator] = createSignal<FilterOperator>(selectedField()?.operators[0] || 'FILTER_OPERATOR_EQ')
    const [value, setValue] = createSignal('')

    const selectField = (nextField: string) => {
        setField(nextField)
        const next = props.fields.find((item) => item.value === nextField)
        setOperator(next?.operators[0] || 'FILTER_OPERATOR_EQ')
    }

    const addFilter = (event: SubmitEvent) => {
        event.preventDefault()
        const normalized = value().trim()
        if (!field() || !normalized) return
        const values =
            operator() === 'FILTER_OPERATOR_IN'
                ? normalized
                      .split(',')
                      .map((item) => item.trim())
                      .filter(Boolean)
                : [normalized]
        props.onChange([
            ...props.filters.filter((item) => item.field !== field()),
            { field: field(), operator: operator(), values },
        ])
        setValue('')
    }

    const removeFilter = (index: number) => {
        props.onChange(props.filters.filter((_, itemIndex) => itemIndex !== index))
    }

    return (
        <div class="space-y-3">
            <form
                class="flex flex-wrap items-end gap-3 rounded-md border border-gray-200 bg-gray-50 p-3"
                onSubmit={addFilter}
            >
                <label class="min-w-40 flex-1 space-y-1 text-sm text-gray-600">
                    <span class="font-medium text-gray-800">Filter field</span>
                    <select
                        value={field()}
                        class="w-full rounded-md border border-gray-300 bg-white px-3 py-2"
                        onChange={(event) => selectField(event.currentTarget.value)}
                    >
                        <For each={props.fields}>{(item) => <option value={item.value}>{item.label}</option>}</For>
                    </select>
                </label>

                <label class="min-w-40 flex-1 space-y-1 text-sm text-gray-600">
                    <span class="font-medium text-gray-800">Operator</span>
                    <select
                        value={operator()}
                        class="w-full rounded-md border border-gray-300 bg-white px-3 py-2"
                        onChange={(event) => setOperator(event.currentTarget.value as FilterOperator)}
                    >
                        <For each={selectedField()?.operators || []}>
                            {(item) => <option value={item}>{operatorLabels[item]}</option>}
                        </For>
                    </select>
                </label>

                <label class="min-w-48 flex-[2] space-y-1 text-sm text-gray-600">
                    <span class="font-medium text-gray-800">Value</span>
                    <input
                        value={value()}
                        placeholder={operator() === 'FILTER_OPERATOR_IN' ? 'Comma-separated values' : 'Filter value'}
                        class="w-full rounded-md border border-gray-300 bg-white px-3 py-2"
                        onInput={(event) => setValue(event.currentTarget.value)}
                    />
                </label>

                <Button type="submit" color="alternative" size="sm">
                    Add filter
                </Button>
            </form>

            <Show when={props.filters.length > 0}>
                <div class="flex flex-wrap gap-2" aria-label="Active filters">
                    <For each={props.filters}>
                        {(filter, index) => (
                            <Button type="button" color="light" size="xs" onClick={() => removeFilter(index())}>
                                {filter.field} {operatorLabels[filter.operator]} {filter.values.join(', ')} ×
                            </Button>
                        )}
                    </For>
                </div>
            </Show>
        </div>
    )
}
