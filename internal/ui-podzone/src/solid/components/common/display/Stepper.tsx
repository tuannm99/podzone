import { For, Show } from 'solid-js'
import { classes } from '../../../shared/utils'
import { type StepStatus } from './shared'

export type StepperItem = {
  title: string
  description?: string
  status?: StepStatus
}

export function Stepper(props: { items: StepperItem[]; class?: string }) {
  return (
    <ol class={classes('space-y-4', props.class)}>
      <For each={props.items}>
        {(item, index) => {
          const status = () => item.status ?? 'upcoming'

          return (
            <li class="flex gap-4">
              <div class="flex flex-col items-center">
                <div
                  class={classes(
                    'flex size-8 items-center justify-center rounded-full text-xs font-semibold',
                    status() === 'complete'
                      ? 'bg-green-600 text-white'
                      : status() === 'current'
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 text-gray-600'
                  )}
                >
                  {status() === 'complete' ? '✓' : index() + 1}
                </div>
                <Show when={index() < props.items.length - 1}>
                  <div class="mt-2 h-full min-h-6 w-px bg-gray-200" />
                </Show>
              </div>
              <div class="space-y-1 pb-2">
                <p class="text-sm font-semibold text-gray-900">{item.title}</p>
                <Show when={item.description}>
                  <p class="text-sm text-gray-500">{item.description}</p>
                </Show>
              </div>
            </li>
          )
        }}
      </For>
    </ol>
  )
}
