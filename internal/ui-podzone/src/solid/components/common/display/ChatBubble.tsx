import { Show } from 'solid-js'
import { classes } from '../../../shared/utils'

export type ChatBubbleProps = {
  author?: string
  copy: string
  meta?: string
  align?: 'start' | 'end'
  class?: string
}

export function ChatBubble(props: ChatBubbleProps) {
  const isEnd = () => props.align === 'end'

  return (
    <div class={classes('flex w-full', isEnd() ? 'justify-end' : 'justify-start', props.class)}>
      <div
        class={classes(
          'max-w-xl rounded-2xl px-4 py-3 shadow-sm',
          isEnd() ? 'bg-blue-700 text-white' : 'bg-white text-gray-900 ring-1 ring-gray-200'
        )}
      >
        <Show when={props.author || props.meta}>
          <div
            class={classes(
              'mb-1 flex flex-wrap items-center gap-2 text-xs',
              isEnd() ? 'text-blue-100' : 'text-gray-500'
            )}
          >
            <Show when={props.author}>
              <span class="font-semibold">{props.author}</span>
            </Show>
            <Show when={props.meta}>
              <span>{props.meta}</span>
            </Show>
          </div>
        </Show>
        <p class="text-sm leading-6">{props.copy}</p>
      </div>
    </div>
  )
}
