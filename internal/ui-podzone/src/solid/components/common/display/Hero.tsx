import { Show, type JSX } from 'solid-js'
import { classes } from '../../../shared/utils'

export function Hero(props: {
    eyebrow?: string
    title: string
    copy?: string
    action?: JSX.Element
    secondaryAction?: JSX.Element
    class?: string
}) {
    return (
        <section class={classes('rounded-lg border border-gray-200 bg-white p-8 shadow-sm sm:p-10', props.class)}>
            <div class="max-w-3xl space-y-4">
                <Show when={props.eyebrow}>
                    <p class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-500">{props.eyebrow}</p>
                </Show>
                <h1 class="text-4xl font-semibold tracking-tight text-gray-900 sm:text-5xl">{props.title}</h1>
                <Show when={props.copy}>
                    <p class="text-base leading-7 text-gray-600 sm:text-lg">{props.copy}</p>
                </Show>
                <Show when={props.action || props.secondaryAction}>
                    <div class="flex flex-wrap gap-3 pt-2">
                        <Show when={props.action}>
                            <div>{props.action}</div>
                        </Show>
                        <Show when={props.secondaryAction}>
                            <div>{props.secondaryAction}</div>
                        </Show>
                    </div>
                </Show>
            </div>
        </section>
    )
}

export const Jumbotron = Hero
