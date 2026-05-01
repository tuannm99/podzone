import { Show, createSignal, type ParentProps } from 'solid-js';
import { classes } from '../../shared/utils';

export function ClipboardButton(props: {
  text: string;
  label?: string;
  copiedLabel?: string;
  class?: string;
}) {
  const [copied, setCopied] = createSignal(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(props.text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  }

  return (
    <button
      type="button"
      class={classes(
        'inline-flex items-center gap-2 rounded-xl border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition hover:bg-gray-50',
        copied() && 'border-green-200 bg-green-50 text-green-700',
        props.class
      )}
      onClick={() => void handleCopy()}
    >
      <span>
        {copied() ? (props.copiedLabel ?? 'Copied') : (props.label ?? 'Copy')}
      </span>
    </button>
  );
}

export function DeviceMockup(
  props: ParentProps<{
    label?: string;
    class?: string;
    screenClass?: string;
  }>
) {
  return (
    <div
      class={classes('inline-flex flex-col items-center gap-3', props.class)}
    >
      <div class="rounded-[2.5rem] border-8 border-gray-900 bg-gray-900 p-3 shadow-2xl">
        <div class="mx-auto mb-3 h-1.5 w-20 rounded-full bg-gray-700" />
        <div
          class={classes(
            'min-h-96 w-[20rem] overflow-hidden rounded-[2rem] bg-white',
            props.screenClass
          )}
        >
          {props.children}
        </div>
      </div>
      <Show when={props.label}>
        <p class="text-sm font-medium text-gray-500">{props.label}</p>
      </Show>
    </div>
  );
}
