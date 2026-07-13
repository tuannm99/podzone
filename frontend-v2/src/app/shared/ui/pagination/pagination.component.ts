import { Component, computed, input, output } from '@angular/core';
import { classes } from '../../utils';
import { Button } from '../button/button.component';
import { Spinner } from '../spinner/spinner.component';

type PageItem = number | 'ellipsis';

function buildPageItems(page: number, totalPages: number): PageItem[] {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const pages = new Set<number>([1, totalPages, page - 1, page, page + 1]);
  const visible = [...pages]
    .filter((value) => value >= 1 && value <= totalPages)
    .sort((a, b) => a - b);
  const items: PageItem[] = [];

  for (const value of visible) {
    const previous = items[items.length - 1];
    if (typeof previous === 'number' && value - previous > 1) {
      items.push('ellipsis');
    }
    items.push(value);
  }

  return items;
}

// Page-number buttons are raw <button> elements, not <app-button>, so
// [attr.aria-current] can be set directly on the real interactive element —
// setting it on <app-button> from outside would land on that component's
// host tag, not the <button>/<a> it renders internally, which is not what
// a screen reader needs (see ANGULAR_STYLE_GUIDE.md's Pagination buttons
// rule). Classes below are copied from Button's pill/xs/dark/alternative
// variants to keep the exact same visual output.
const pageButtonBase =
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full font-medium focus:outline-none focus:ring-2 disabled:pointer-events-none disabled:opacity-60 h-8 px-3 text-xs';
const pageButtonActive = 'bg-gray-950 text-white hover:bg-gray-800 focus:ring-gray-300';
const pageButtonInactive =
  'border border-gray-300 bg-white text-gray-900 hover:bg-gray-50 focus:ring-gray-200';

@Component({
  selector: 'app-pagination',
  imports: [Button, Spinner],
  templateUrl: './pagination.component.html',
})
export class Pagination {
  page = input.required<number>();
  pageSize = input.required<number>();
  total = input.required<number>();
  loading = input(false);
  class = input<string>();

  pageChange = output<number>();

  // Angular templates can't call global functions directly — expose as a
  // class member so [prevPage]="Math.max(...)"-style expressions resolve.
  protected readonly Math = Math;

  protected totalPages = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize())));
  protected start = computed(() =>
    this.total() === 0 ? 0 : (this.page() - 1) * this.pageSize() + 1,
  );
  protected finish = computed(() => Math.min(this.page() * this.pageSize(), this.total()));
  protected items = computed(() => buildPageItems(this.page(), this.totalPages()));

  protected containerClass = computed(() =>
    classes(
      'flex flex-col gap-4 border-t border-gray-100 pt-4 md:flex-row md:items-center md:justify-between',
      this.class(),
    ),
  );

  protected pageButtonClass(item: number) {
    return classes(pageButtonBase, item === this.page() ? pageButtonActive : pageButtonInactive);
  }

  protected changePage(event: MouseEvent, page: number) {
    event.preventDefault();
    event.stopPropagation();
    if (this.loading() || page === this.page()) return;
    this.pageChange.emit(page);
  }
}
