import { Show } from 'solid-js';

export function CodeBlock(props: {
  code: string;
  label?: string;
  class?: string;
}) {
  return (
    <div class={props.class}>
      <Show when={props.label}>
        <div class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
          {props.label}
        </div>
      </Show>
      <pre class="overflow-x-auto rounded-xl bg-gray-900 p-4 text-xs text-gray-100">
        {props.code || '—'}
      </pre>
    </div>
  );
}
