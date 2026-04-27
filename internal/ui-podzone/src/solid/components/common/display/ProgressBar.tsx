import { Show, createMemo } from 'solid-js'
import { classes } from '../../../shared/utils'
import { type AlertTone, toneFillClasses } from './shared'

export function ProgressBar(props: {
  value: number
  max?: number
  label?: string
  tone?: AlertTone
  showValue?: boolean
  class?: string
}) {
  const percent = createMemo(() => {
    const max = props.max ?? 100
    if (max <= 0) return 0
    return Math.max(0, Math.min(100, Math.round((props.value / max) * 100)))
  })

  return (
    <div class={classes('space-y-2', props.class)}>
      <Show when={props.label || props.showValue}>
        <div class="flex items-center justify-between gap-4 text-sm">
          <Show when={props.label}>
            <span class="font-medium text-gray-700">{props.label}</span>
          </Show>
          <Show when={props.showValue}>
            <span class="text-gray-500">{percent()}%</span>
          </Show>
        </div>
      </Show>
      <div class="h-2.5 overflow-hidden rounded-full bg-gray-200">
        <div
          class={classes(
            'h-full rounded-full transition-[width]',
            toneFillClasses[props.tone ?? 'blue']
          )}
          style={{ width: `${percent()}%` }}
        />
      </div>
    </div>
  )
}
