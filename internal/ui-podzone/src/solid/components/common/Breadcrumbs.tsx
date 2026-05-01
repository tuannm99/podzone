import { For } from 'solid-js';
import { classes } from '../../shared/utils';

export type BreadcrumbItem = {
  label: string;
  href?: string;
  current?: boolean;
};

export function Breadcrumbs(props: {
  items: BreadcrumbItem[];
  class?: string;
}) {
  return (
    <nav aria-label="Breadcrumb" class={props.class}>
      <ol class="flex flex-wrap items-center gap-2 text-sm text-gray-500">
        <For each={props.items}>
          {(item, index) => {
            const current = () =>
              item.current || index() === props.items.length - 1;

            return (
              <>
                <li class="flex items-center gap-2">
                  {item.href && !current() ? (
                    <a
                      href={item.href}
                      class="font-medium text-gray-600 transition hover:text-blue-700"
                    >
                      {item.label}
                    </a>
                  ) : (
                    <span
                      class={classes(
                        'font-medium',
                        current() ? 'text-gray-900' : 'text-gray-600'
                      )}
                      aria-current={current() ? 'page' : undefined}
                    >
                      {item.label}
                    </span>
                  )}
                </li>
                <li
                  class="text-gray-300"
                  aria-hidden="true"
                  hidden={index() === props.items.length - 1}
                >
                  /
                </li>
              </>
            );
          }}
        </For>
      </ol>
    </nav>
  );
}
