import { Show } from 'solid-js'

export function SectionTitle(props: { title: string; subtitle?: string }) {
  return (
    <div class="space-y-1">
      <h2 class="text-base font-semibold text-gray-950">{props.title}</h2>
      <Show when={props.subtitle}>
        <p class="text-sm text-gray-600">{props.subtitle}</p>
      </Show>
    </div>
  )
}
