import { For } from 'solid-js'
import { classes } from '../../../shared/utils'

export function Skeleton(props: { class?: string; circle?: boolean }) {
  return (
    <div
      class={classes(
        'animate-pulse bg-gray-200',
        props.circle ? 'rounded-full' : 'rounded-xl',
        props.class ?? 'h-4 w-full'
      )}
      aria-hidden="true"
    />
  )
}

export function SkeletonText(props: { lines?: number; class?: string }) {
  return (
    <div class={classes('space-y-3', props.class)} aria-hidden="true">
      <For each={Array.from({ length: props.lines ?? 3 }, (_, index) => index)}>
        {(index) => (
          <Skeleton
            class={classes('h-4 rounded-lg', index === (props.lines ?? 3) - 1 ? 'w-3/4' : 'w-full')}
          />
        )}
      </For>
    </div>
  )
}
