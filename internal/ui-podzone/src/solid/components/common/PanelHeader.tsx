import type { JSX } from 'solid-js'
import { classes } from '../../shared/utils'

export function PanelHeader(props: {
  title: string
  copy?: string
  action?: JSX.Element
  class?: string
}) {
  return (
    <div
      class={classes(
        'flex flex-col gap-4 md:flex-row md:items-start md:justify-between',
        props.class
      )}
    >
      <div class="space-y-1">
        <h2 class="text-xl font-semibold tracking-tight text-gray-900">{props.title}</h2>
        {props.copy ? <p class="text-sm text-gray-500">{props.copy}</p> : null}
      </div>
      {props.action ? <div class="shrink-0">{props.action}</div> : null}
    </div>
  )
}
