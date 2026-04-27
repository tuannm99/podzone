import { Show, type ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'
import { Spinner } from './Primitives'

function AlertBox(props: ParentProps<{ color: 'failure' | 'warning' | 'info' }>) {
  const colorClasses = {
    failure: 'border-red-200 bg-red-50 text-red-800',
    warning: 'border-amber-200 bg-amber-50 text-amber-800',
    info: 'border-blue-200 bg-blue-50 text-blue-800'
  }

  return (
    <div
      class={classes('rounded-2xl border px-4 py-3 text-sm shadow-sm', colorClasses[props.color])}
    >
      {props.children}
    </div>
  )
}

export function LoadingInline(props: { label: string }) {
  return (
    <div class="flex items-center gap-3 text-sm text-gray-500">
      <Spinner />
      <span>{props.label}</span>
    </div>
  )
}

export function LoadingBlock(props: { label: string }) {
  return (
    <div class="flex min-h-40 items-center justify-center">
      <LoadingInline label={props.label} />
    </div>
  )
}

export function EmptyBlock(props: { title: string; copy: string }) {
  return (
    <div class="rounded-2xl border border-dashed border-gray-200 bg-gray-50 px-6 py-10 text-center">
      <h3 class="text-lg font-semibold text-gray-900">{props.title}</h3>
      <p class="mt-2 text-sm text-gray-500">{props.copy}</p>
    </div>
  )
}

export function ErrorAlert(props: ParentProps) {
  return <AlertBox color="failure">{props.children}</AlertBox>
}

export function WarningAlert(props: ParentProps) {
  return <AlertBox color="warning">{props.children}</AlertBox>
}

export function InfoAlert(props: ParentProps) {
  return <AlertBox color="info">{props.children}</AlertBox>
}

export function InlineMessage(props: { when: boolean; label: string }) {
  return (
    <Show when={props.when}>
      <p class="text-sm text-gray-500">{props.label}</p>
    </Show>
  )
}
