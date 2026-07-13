import { Component, ElementRef, computed, effect, input, signal, viewChild } from '@angular/core';
import { classes } from '../../utils';
// NOTE: depends on Link.tsx's Angular port, assigned to a different
// parallel porting batch. Path assumed per this project's folder-per-
// component convention — reconcile at the consolidated build pass if the
// actual path differs.
import { Link } from '../link/link.component';

export type MegaMenuLink = {
  label: string;
  href: string;
  description?: string;
  // Solid's `icon?: JSX.Element` dropped — see nav-item.ts for the same
  // simplification and reasoning.
};

export type MegaMenuSection = {
  title: string;
  links: MegaMenuLink[];
};

export type MegaMenuItem = {
  label: string;
  href?: string;
  active?: boolean;
  sections?: MegaMenuSection[];
};

@Component({
  selector: 'app-mega-menu',
  imports: [Link],
  templateUrl: './mega-menu.component.html',
})
export class MegaMenu {
  items = input.required<MegaMenuItem[]>();
  class = input<string>();
  // Solid's `brand`/`actions` (JSX.Element) become content-projection
  // slots — use <div slot="brand"> / <div slot="actions"> inside
  // <app-mega-menu>. Same "actions rendered twice" simplification as
  // Navbar applies here — see navbar.component.html's comment.

  protected openIndex = signal<number | null>(null);
  protected container = viewChild<ElementRef<HTMLDivElement>>('container');

  protected className = computed(() =>
    classes(
      'relative rounded-lg border border-gray-200 bg-white px-5 py-4 shadow-sm',
      this.class(),
    ),
  );

  protected openSections = computed(() => {
    const index = this.openIndex();
    if (index === null) return [];
    return this.items()[index]?.sections ?? [];
  });

  constructor() {
    effect((onCleanup) => {
      if (this.openIndex() === null) return;

      const handlePointerDown = (event: MouseEvent) => {
        const element = this.container()?.nativeElement;
        if (element && !element.contains(event.target as Node)) {
          this.openIndex.set(null);
        }
      };

      document.addEventListener('mousedown', handlePointerDown);
      onCleanup(() => document.removeEventListener('mousedown', handlePointerDown));
    });
  }

  protected toggleIndex(index: number) {
    this.openIndex.update((value) => (value === index ? null : index));
  }

  protected itemButtonClass(item: MegaMenuItem, index: number) {
    return classes(
      'inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition',
      item.active || this.openIndex() === index
        ? 'bg-gray-100 text-gray-950'
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
    );
  }

  protected chevronClass(index: number) {
    return classes('text-gray-400 transition', this.openIndex() === index && 'rotate-180');
  }

  protected linkClass(item: MegaMenuItem) {
    return classes(
      'inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition',
      item.active
        ? 'bg-gray-950 text-white'
        : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
    );
  }

  protected closeMenu() {
    this.openIndex.set(null);
  }
}
