import { For } from 'solid-js';
import { classes } from '../../shared/utils';

export type TabItem = {
  value: string;
  label: string;
};

export function Tabs(props: {
  value: string;
  items: TabItem[];
  onChange: (value: string) => void;
  class?: string;
}) {
  return (
    <div class={classes('flex flex-wrap items-center gap-2', props.class)}>
      <For each={props.items}>
        {(item) => (
          <button
            type="button"
            class={classes(
              'rounded-full px-4 py-2 text-sm font-medium transition',
              props.value === item.value
                ? 'bg-blue-700 text-white shadow-sm'
                : 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50'
            )}
            onClick={() => props.onChange(item.value)}
          >
            {item.label}
          </button>
        )}
      </For>
    </div>
  );
}
