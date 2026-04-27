import type { ParentProps } from 'solid-js'

export function PageShell(props: ParentProps) {
  return <div class="space-y-6">{props.children}</div>
}
