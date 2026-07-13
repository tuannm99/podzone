import { Component, computed, input, output } from '@angular/core';
import { classes } from '../../utils';

export type TabItem<T extends string = string> = {
  value: T;
  label: string;
};

@Component({
  selector: 'app-tabs',
  templateUrl: './tabs.component.html',
})
export class Tabs<T extends string = string> {
  value = input.required<T>();
  items = input.required<Array<TabItem<T>>>();
  variant = input<'pills' | 'underline'>('pills');
  ariaLabel = input<string>();
  class = input<string>();

  valueChange = output<T>();

  protected isUnderline = computed(() => this.variant() === 'underline');

  protected rootClass = computed(() =>
    classes(
      this.isUnderline()
        ? 'overflow-x-auto border-b border-gray-200'
        : 'flex flex-wrap items-center gap-2',
      this.class(),
    ),
  );

  protected groupClass = computed(() =>
    classes(this.isUnderline() ? 'flex min-w-max gap-6' : 'contents'),
  );

  protected tabClass(item: TabItem<T>) {
    const selected = this.value() === item.value;
    const underline = this.isUnderline();
    return classes(
      'text-sm font-medium transition',
      underline ? 'border-b-2 px-1 py-3' : 'rounded-md px-3 py-2',
      selected && underline
        ? 'border-gray-950 text-gray-950'
        : selected
          ? 'bg-gray-950 text-white shadow-sm'
          : underline
            ? 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-900'
            : 'border border-gray-200 bg-white text-gray-700 hover:bg-gray-50',
    );
  }

  protected select(item: TabItem<T>) {
    this.valueChange.emit(item.value);
  }

  protected onKeyDown(event: KeyboardEvent, index: number) {
    const items = this.items();
    const lastIndex = items.length - 1;
    let nextIndex: number | undefined;

    if (event.key === 'ArrowRight') nextIndex = index === lastIndex ? 0 : index + 1;
    if (event.key === 'ArrowLeft') nextIndex = index === 0 ? lastIndex : index - 1;
    if (event.key === 'Home') nextIndex = 0;
    if (event.key === 'End') nextIndex = lastIndex;
    if (nextIndex === undefined) return;

    event.preventDefault();
    const nextItem = items[nextIndex];
    if (!nextItem) return;
    this.valueChange.emit(nextItem.value);

    const tablist = (event.currentTarget as HTMLElement).closest('[role="tablist"]');
    const tabs = tablist?.querySelectorAll<HTMLButtonElement>('[role="tab"]');
    tabs?.[nextIndex]?.focus();
  }
}
