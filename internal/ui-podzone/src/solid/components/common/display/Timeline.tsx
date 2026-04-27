import { For, Show } from 'solid-js'
import { classes } from '../../../shared/utils'

export type TimelineItem = {
  title: string
  meta?: string
  copy?: string
}

export function Timeline(props: { items: TimelineItem[]; class?: string }) {
  return (
    <ol class={classes('relative space-y-6 border-s border-gray-200 ps-6', props.class)}>
      <For each={props.items}>
        {(item) => (
          <li class="relative">
            <span class="absolute -start-[2.05rem] mt-1.5 size-3 rounded-full bg-blue-600 ring-4 ring-white" />
            <div class="space-y-1">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-900">{item.title}</h3>
                <Show when={item.meta}>
                  <span class="text-xs uppercase tracking-wide text-gray-400">{item.meta}</span>
                </Show>
              </div>
              <Show when={item.copy}>
                <p class="text-sm leading-6 text-gray-500">{item.copy}</p>
              </Show>
            </div>
          </li>
        )}
      </For>
    </ol>
  )
}
