import { Show, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../../shared/utils'
import { type AlertTone, toneClasses } from './shared'

export { type AlertTone } from './shared'

export function Alert(
  props: ParentProps<{
    tone?: AlertTone
    title?: string
    action?: JSX.Element
    class?: string
  }>
) {
  return (
    <div
      class={classes(
        'flex flex-col gap-3 rounded-2xl border px-4 py-3 shadow-sm md:flex-row md:items-start md:justify-between',
        toneClasses[props.tone ?? 'blue'],
        props.class
      )}
      role="alert"
    >
      <div class="space-y-1">
        <Show when={props.title}>
          <p class="text-sm font-semibold">{props.title}</p>
        </Show>
        <div class="text-sm">{props.children}</div>
      </div>
      <Show when={props.action}>
        <div class="shrink-0">{props.action}</div>
      </Show>
    </div>
  )
}
