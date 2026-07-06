import { For, Show, createEffect, createMemo, createSignal, createUniqueId, onCleanup, onMount } from 'solid-js'
import { classes } from '../../shared/utils'
import { FieldLabel, Spinner } from './Primitives'

export type SearchSelectOption = {
    value: string
    label: string
    description?: string
}

export function SearchSelectField(props: {
    label: string
    value: string
    options: SearchSelectOption[]
    onChange: (value: string) => void
    onSearch: (query: string) => void
    placeholder?: string
    loading?: boolean
    error?: string
    emptyText?: string
}) {
    let container!: HTMLDivElement
    let searchTimer: ReturnType<typeof setTimeout> | undefined
    const listboxId = createUniqueId()
    const [open, setOpen] = createSignal(false)
    const [query, setQuery] = createSignal('')
    const [activeIndex, setActiveIndex] = createSignal(-1)
    const selected = createMemo(() => props.options.find((option) => option.value === props.value))

    createEffect(() => {
        const option = selected()
        if (option && !open()) setQuery(option.label)
        if (!props.value && !open()) setQuery('')
    })

    createEffect(() => {
        const optionCount = props.options.length
        setActiveIndex((current) => (current < optionCount ? current : -1))
    })

    onMount(() => {
        const closeOnOutsideClick = (event: PointerEvent) => {
            if (!container.contains(event.target as Node)) setOpen(false)
        }
        document.addEventListener('pointerdown', closeOnOutsideClick)
        onCleanup(() => document.removeEventListener('pointerdown', closeOnOutsideClick))
    })

    onCleanup(() => {
        if (searchTimer) clearTimeout(searchTimer)
    })

    const search = (value: string) => {
        setQuery(value)
        setOpen(true)
        setActiveIndex(-1)
        if (props.value) props.onChange('')
        if (searchTimer) clearTimeout(searchTimer)
        searchTimer = setTimeout(() => props.onSearch(value), 250)
    }

    const choose = (option: SearchSelectOption) => {
        props.onChange(option.value)
        setQuery(option.label)
        setOpen(false)
    }

    const moveActiveOption = (offset: number) => {
        if (props.options.length === 0) return
        setOpen(true)
        setActiveIndex((current) => {
            const next = current + offset
            if (next < 0) return props.options.length - 1
            if (next >= props.options.length) return 0
            return next
        })
    }

    return (
        <div ref={(element) => (container = element)} class="relative">
            <FieldLabel label={props.label}>
                <div class="relative">
                    <input
                        role="combobox"
                        aria-expanded={open()}
                        aria-autocomplete="list"
                        aria-controls={listboxId}
                        aria-activedescendant={activeIndex() >= 0 ? `${listboxId}-option-${activeIndex()}` : undefined}
                        class={classes(
                            'block h-10 w-full rounded-md border bg-white px-3 pr-10 text-sm text-gray-900 outline-none transition',
                            props.error
                                ? 'border-red-300 focus:border-red-500 focus:ring-2 focus:ring-red-100'
                                : 'border-gray-300 focus:border-gray-950 focus:ring-2 focus:ring-gray-100'
                        )}
                        value={query()}
                        placeholder={props.placeholder || 'Search...'}
                        onFocus={() => {
                            setOpen(true)
                            props.onSearch('')
                        }}
                        onInput={(event) => search(event.currentTarget.value)}
                        onKeyDown={(event) => {
                            if (event.key === 'ArrowDown') {
                                event.preventDefault()
                                moveActiveOption(1)
                            } else if (event.key === 'ArrowUp') {
                                event.preventDefault()
                                moveActiveOption(-1)
                            } else if (event.key === 'Enter' && open()) {
                                const option = props.options[activeIndex()]
                                if (option) {
                                    event.preventDefault()
                                    choose(option)
                                }
                            } else if (event.key === 'Escape') {
                                setOpen(false)
                            }
                        }}
                    />
                    <Show when={props.loading}>
                        <Spinner class="absolute right-3 top-3 size-4 text-gray-500" />
                    </Show>
                </div>
            </FieldLabel>

            <Show when={props.error}>
                <p class="mt-1 text-xs font-medium text-red-600">{props.error}</p>
            </Show>

            <Show when={open()}>
                <div
                    id={listboxId}
                    role="listbox"
                    class="absolute z-30 mt-1 max-h-64 w-full overflow-y-auto rounded-md border border-gray-200 bg-white p-1 shadow-lg"
                >
                    <Show
                        when={props.options.length > 0}
                        fallback={
                            <p class="px-3 py-4 text-center text-sm text-gray-500">
                                {props.loading ? 'Loading...' : props.emptyText || 'No matching results'}
                            </p>
                        }
                    >
                        <For each={props.options}>
                            {(option, index) => (
                                <button
                                    id={`${listboxId}-option-${index()}`}
                                    type="button"
                                    role="option"
                                    aria-selected={option.value === props.value}
                                    class={classes(
                                        'block w-full rounded px-3 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none',
                                        (option.value === props.value || index() === activeIndex()) && 'bg-gray-100'
                                    )}
                                    onMouseDown={(event) => event.preventDefault()}
                                    onClick={() => choose(option)}
                                >
                                    <span class="block truncate text-sm font-medium text-gray-900">{option.label}</span>
                                    <Show when={option.description}>
                                        <span class="block truncate text-xs text-gray-500">{option.description}</span>
                                    </Show>
                                </button>
                            )}
                        </For>
                    </Show>
                </div>
            </Show>
        </div>
    )
}
