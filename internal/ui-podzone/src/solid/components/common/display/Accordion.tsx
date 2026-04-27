import { For, Show, type JSX } from 'solid-js'
import { classes } from '../../../shared/utils'

export type AccordionItem = {
  title: string
  content: JSX.Element
  description?: string
  defaultOpen?: boolean
  badge?: string
}

export function Accordion(props: { items: AccordionItem[]; class?: string }) {
  return (
    <div
      class={classes(
        'divide-y divide-gray-200 overflow-hidden rounded-2xl border border-gray-200 bg-white',
        props.class
      )}
    >
      <For each={props.items}>
        {(item) => (
          <details open={item.defaultOpen} class="group px-5 py-4">
            <summary class="flex cursor-pointer list-none items-start justify-between gap-4">
              <div class="space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-sm font-semibold text-gray-900">{item.title}</h3>
                  <Show when={item.badge}>
                    <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
                      {item.badge}
                    </span>
                  </Show>
                </div>
                <Show when={item.description}>
                  <p class="text-sm text-gray-500">{item.description}</p>
                </Show>
              </div>
              <span class="pt-0.5 text-gray-400 transition group-open:rotate-180">⌄</span>
            </summary>
            <div class="mt-4 text-sm leading-6 text-gray-600">{item.content}</div>
          </details>
        )}
      </For>
    </div>
  )
}
