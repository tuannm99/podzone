import { Show, createSignal, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../../shared/utils'
import { type AlertTone, toneClasses } from './shared'

export function Banner(
  props: ParentProps<{
    tone?: AlertTone
    title?: string
    action?: JSX.Element
    dismissible?: boolean
    class?: string
  }>
) {
  const [dismissed, setDismissed] = createSignal(false)

  return (
    <Show when={!dismissed()}>
      <div
        class={classes(
          'flex flex-col gap-3 rounded-2xl border px-4 py-3 shadow-sm md:flex-row md:items-center md:justify-between',
          toneClasses[props.tone ?? 'dark'],
          props.class
        )}
      >
        <div class="space-y-1">
          <Show when={props.title}>
            <p class="text-sm font-semibold">{props.title}</p>
          </Show>
          <div class="text-sm">{props.children}</div>
        </div>
        <div class="flex items-center gap-2">
          <Show when={props.action}>
            <div>{props.action}</div>
          </Show>
          <Show when={props.dismissible}>
            <button
              type="button"
              class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 transition hover:bg-white/70 hover:text-gray-900"
              onClick={() => setDismissed(true)}
              aria-label="Dismiss banner"
            >
              ✕
            </button>
          </Show>
        </div>
      </div>
    </Show>
  )
}
