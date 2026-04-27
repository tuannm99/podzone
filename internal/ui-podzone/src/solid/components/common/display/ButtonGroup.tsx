import type { ParentProps } from 'solid-js'
import { classes } from '../../../shared/utils'

export function ButtonGroup(props: ParentProps<{ class?: string }>) {
  return (
    <div role="group" class={classes('inline-flex flex-wrap items-center gap-2', props.class)}>
      {props.children}
    </div>
  )
}
