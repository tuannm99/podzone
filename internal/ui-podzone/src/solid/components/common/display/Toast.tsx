import { Show, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../../shared/utils'
import { type AlertTone, toneClasses } from './shared'

export function Toast(
  props: ParentProps<{
    show?: boolean
    tone?: AlertTone
    title?: string
    fixed?: boolean
    action?: JSX.Element
    onClose?: () => void
    class?: string
  }>
) {
  return (
    <Show when={props.show ?? true}>
      <div
        class={classes(
          'z-50 flex max-w-sm items-start justify-between gap-4 rounded-2xl border bg-white px-4 py-3 shadow-xl',
          props.fixed !== false && 'fixed bottom-4 right-4',
          toneClasses[props.tone ?? 'dark'],
          props.class
        )}
        role="status"
      >
        <div class="space-y-1">
          <Show when={props.title}>
            <p class="text-sm font-semibold">{props.title}</p>
          </Show>
          <div class="text-sm">{props.children}</div>
        </div>
        <div class="flex items-start gap-2">
          <Show when={props.action}>
            <div class="shrink-0">{props.action}</div>
          </Show>
          <Show when={props.onClose}>
            <button
              type="button"
              class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 hover:bg-white/70 hover:text-gray-900"
              onClick={() => props.onClose?.()}
              aria-label="Close toast"
            >
              ✕
            </button>
          </Show>
        </div>
      </div>
    </Show>
  )
}
