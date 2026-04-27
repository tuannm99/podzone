import { For, Show, type JSX } from 'solid-js'
import { classes } from '../../../shared/utils'

export type ListGroupItem = {
  label: string
  description?: string
  href?: string
  prefix?: JSX.Element
  suffix?: JSX.Element
  active?: boolean
  onClick?: () => void
}

export function ListGroup(props: { items: ListGroupItem[]; class?: string }) {
  return (
    <div
      class={classes(
        'overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm',
        props.class
      )}
    >
      <For each={props.items}>
        {(item) => {
          const itemClass = classes(
            'flex w-full items-center justify-between gap-4 border-b border-gray-100 px-4 py-3 text-left transition last:border-b-0',
            item.active ? 'bg-blue-50 text-blue-900' : 'text-gray-700 hover:bg-gray-50'
          )

          const content = (
            <>
              <div class="flex min-w-0 items-start gap-3">
                <Show when={item.prefix}>
                  <div class="pt-0.5">{item.prefix}</div>
                </Show>
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium">{item.label}</p>
                  <Show when={item.description}>
                    <p class="mt-1 text-sm text-gray-500">{item.description}</p>
                  </Show>
                </div>
              </div>
              <Show when={item.suffix}>
                <div class="shrink-0">{item.suffix}</div>
              </Show>
            </>
          )

          return item.href ? (
            <a href={item.href} class={itemClass}>
              {content}
            </a>
          ) : item.onClick ? (
            <button type="button" class={itemClass} onClick={item.onClick}>
              {content}
            </button>
          ) : (
            <div class={itemClass}>{content}</div>
          )
        }}
      </For>
    </div>
  )
}
