import { For } from 'solid-js'
import { classes } from '../../shared/utils'

export type TabItem<T extends string = string> = {
    value: T
    label: string
}

export function Tabs<T extends string>(props: {
    value: T
    items: Array<TabItem<T>>
    onChange: (value: T) => void
    variant?: 'pills' | 'underline'
    ariaLabel?: string
    class?: string
}) {
    const isUnderline = () => props.variant === 'underline'
    const handleKeyDown = (event: KeyboardEvent & { currentTarget: HTMLButtonElement }, index: number) => {
        const lastIndex = props.items.length - 1
        let nextIndex: number | undefined

        if (event.key === 'ArrowRight') nextIndex = index === lastIndex ? 0 : index + 1
        if (event.key === 'ArrowLeft') nextIndex = index === 0 ? lastIndex : index - 1
        if (event.key === 'Home') nextIndex = 0
        if (event.key === 'End') nextIndex = lastIndex
        if (nextIndex === undefined) return

        event.preventDefault()
        const nextItem = props.items[nextIndex]
        if (!nextItem) return
        props.onChange(nextItem.value)
        const tablist = event.currentTarget
            .closest('[role="tablist"]')
            ?.querySelectorAll<HTMLButtonElement>('[role="tab"]')
        tablist?.[nextIndex]?.focus()
    }

    return (
        <div
            role="tablist"
            aria-label={props.ariaLabel}
            class={classes(
                isUnderline() ? 'overflow-x-auto border-b border-gray-200' : 'flex flex-wrap items-center gap-2',
                props.class
            )}
        >
            <div class={classes(isUnderline() ? 'flex min-w-max gap-6' : 'contents')}>
                <For each={props.items}>
                    {(item, index) => {
                        const selected = () => props.value === item.value
                        return (
                            <button
                                type="button"
                                role="tab"
                                aria-selected={selected()}
                                tabIndex={selected() ? 0 : -1}
                                class={classes(
                                    'text-sm font-medium transition',
                                    isUnderline() ? 'border-b-2 px-1 py-3' : 'rounded-md px-3 py-2',
                                    selected() && isUnderline()
                                        ? 'border-gray-950 text-gray-950'
                                        : selected()
                                          ? 'bg-gray-950 text-white shadow-sm'
                                          : isUnderline()
                                            ? 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-900'
                                            : 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50'
                                )}
                                onClick={() => props.onChange(item.value)}
                                onKeyDown={(event) => handleKeyDown(event, index())}
                            >
                                {item.label}
                            </button>
                        )
                    }}
                </For>
            </div>
        </div>
    )
}
