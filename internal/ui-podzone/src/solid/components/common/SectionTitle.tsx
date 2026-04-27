import { Show } from 'solid-js'

export function SectionTitle(props: { title: string; subtitle?: string }) {
  return (
    <div class="space-y-1">
      <h2 class="text-xl font-semibold tracking-tight text-gray-900">{props.title}</h2>
      <Show when={props.subtitle}>
        <p class="text-sm text-gray-500">{props.subtitle}</p>
      </Show>
    </div>
  )
}
