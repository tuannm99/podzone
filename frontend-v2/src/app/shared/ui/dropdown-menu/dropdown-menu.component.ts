import { Component, ElementRef, computed, effect, input, signal, viewChild } from '@angular/core';
import { classes } from '../../utils';

export type DropdownItem = {
  label: string;
  href?: string;
  onSelect?: () => void;
  tone?: 'default' | 'danger';
};

@Component({
  selector: 'app-dropdown-menu',
  templateUrl: './dropdown-menu.component.html',
})
export class DropdownMenu {
  label = input<string>();
  items = input.required<DropdownItem[]>();
  class = input<string>();
  menuClass = input<string>();

  protected open = signal(false);
  private container = viewChild<ElementRef<HTMLDivElement>>('container');

  protected containerClass = computed(() =>
    classes('relative inline-block text-left', this.class()),
  );
  protected menuClassName = computed(() =>
    classes(
      'absolute right-0 z-30 mt-2 min-w-48 rounded-lg border border-gray-200 bg-white p-2 shadow-xl',
      this.menuClass(),
    ),
  );

  constructor() {
    effect((onCleanup) => {
      if (!this.open()) return;

      const handlePointerDown = (event: MouseEvent) => {
        const element = this.container()?.nativeElement;
        if (element && !element.contains(event.target as Node)) {
          this.open.set(false);
        }
      };

      document.addEventListener('mousedown', handlePointerDown);
      onCleanup(() => document.removeEventListener('mousedown', handlePointerDown));
    });
  }

  protected toggle() {
    this.open.update((value) => !value);
  }

  protected linkItemClass(item: DropdownItem) {
    return classes(
      'block rounded-md px-3 py-2 text-sm transition hover:bg-gray-50',
      item.tone === 'danger' ? 'text-red-600' : 'text-gray-700',
    );
  }

  protected buttonItemClass(item: DropdownItem) {
    return classes(
      'block w-full rounded-md px-3 py-2 text-left text-sm transition hover:bg-gray-50',
      item.tone === 'danger' ? 'text-red-600' : 'text-gray-700',
    );
  }

  protected select(item: DropdownItem) {
    item.onSelect?.();
    this.open.set(false);
  }
}
