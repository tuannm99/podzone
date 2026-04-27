import { For } from 'solid-js'
import { classes } from '../../../shared/utils'

export function Rating(props: { value: number; max?: number; class?: string }) {
  const max = () => props.max ?? 5

  return (
    <div
      class={classes('inline-flex items-center gap-1', props.class)}
      aria-label={`Rated ${props.value} out of ${max()}`}
    >
      <For each={Array.from({ length: max() }, (_, index) => index + 1)}>
        {(item) => (
          <span
            class={classes('text-lg', item <= props.value ? 'text-amber-400' : 'text-gray-300')}
          >
            ★
          </span>
        )}
      </For>
    </div>
  )
}
