import { For } from 'solid-js'
import { classes } from '../../shared/utils'
import { Link } from './Link'

export type BreadcrumbItem = {
    label: string
    href?: string
    current?: boolean
}

export function Breadcrumbs(props: { items: BreadcrumbItem[]; class?: string }) {
    return (
        <nav aria-label="Breadcrumb" class={props.class}>
            <ol class="flex flex-wrap items-center gap-2 text-sm text-gray-500">
                <For each={props.items}>
                    {(item, index) => {
                        const current = () => item.current || index() === props.items.length - 1

                        return (
                            <>
                                <li class="flex items-center gap-2">
                                    {item.href && !current() ? (
                                        <Link
                                            href={item.href}
                                            class="font-medium text-gray-600 transition hover:text-gray-950"
                                        >
                                            {item.label}
                                        </Link>
                                    ) : (
                                        <span
                                            class={classes(
                                                'font-medium',
                                                current() ? 'text-gray-900' : 'text-gray-600'
                                            )}
                                            aria-current={current() ? 'page' : undefined}
                                        >
                                            {item.label}
                                        </span>
                                    )}
                                </li>
                                <li
                                    class="text-gray-300"
                                    aria-hidden="true"
                                    hidden={index() === props.items.length - 1}
                                >
                                    /
                                </li>
                            </>
                        )
                    }}
                </For>
            </ol>
        </nav>
    )
}
