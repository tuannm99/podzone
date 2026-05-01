import type { ParentProps } from 'solid-js';

export function PageShell(props: ParentProps) {
  return <div class="space-y-4 lg:space-y-5">{props.children}</div>;
}
