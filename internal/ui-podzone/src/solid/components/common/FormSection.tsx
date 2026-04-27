import type { JSX, ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'

export function FormSection(
  props: ParentProps<{
    title: string
    description?: string
    aside?: JSX.Element
    class?: string
    contentClass?: string
  }>
) {
  return (
    <section
      class={classes('rounded-2xl border border-gray-200 bg-white p-6 shadow-sm', props.class)}
    >
      <div class="flex flex-col gap-4 border-b border-gray-100 pb-4 md:flex-row md:items-start md:justify-between">
        <div class="space-y-1">
          <h2 class="text-lg font-semibold text-gray-900">{props.title}</h2>
          {props.description ? (
            <p class="max-w-2xl text-sm text-gray-500">{props.description}</p>
          ) : null}
        </div>
        {props.aside ? <div class="shrink-0">{props.aside}</div> : null}
      </div>

      <div class={classes('mt-6 space-y-5', props.contentClass)}>{props.children}</div>
    </section>
  )
}
