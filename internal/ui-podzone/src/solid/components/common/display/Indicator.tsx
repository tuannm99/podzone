import { Show } from 'solid-js'
import { classes } from '../../../shared/utils'
import { indicatorClasses } from './shared'

export function Indicator(props: {
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'gray'
  ping?: boolean
  class?: string
}) {
  return (
    <span class={classes('relative inline-flex size-3', props.class)}>
      <Show when={props.ping}>
        <span
          class={classes(
            'absolute inline-flex h-full w-full animate-ping rounded-full opacity-75',
            indicatorClasses[props.color ?? 'green']
          )}
        />
      </Show>
      <span
        class={classes(
          'relative inline-flex size-3 rounded-full',
          indicatorClasses[props.color ?? 'green']
        )}
      />
    </span>
  )
}
