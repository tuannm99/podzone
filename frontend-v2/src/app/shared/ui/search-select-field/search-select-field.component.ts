import {
  Component,
  ElementRef,
  computed,
  effect,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import { classes } from '../../utils';
import { Spinner } from '../spinner/spinner.component';

export type SearchSelectOption = {
  value: string;
  label: string;
  description?: string;
};

let uidCounter = 0;

@Component({
  selector: 'app-search-select-field',
  imports: [Spinner],
  templateUrl: './search-select-field.component.html',
})
export class SearchSelectField {
  private host = inject(ElementRef<HTMLElement>);
  private searchTimer?: ReturnType<typeof setTimeout>;

  label = input.required<string>();
  value = input.required<string>();
  options = input.required<SearchSelectOption[]>();
  placeholder = input<string>();
  loading = input(false);
  error = input<string>();
  emptyText = input<string>();

  valueChange = output<string>();
  searchQuery = output<string>();

  protected listboxId = `search-select-${++uidCounter}`;
  protected open = signal(false);
  protected query = signal('');
  protected activeIndex = signal(-1);

  protected selected = computed(() =>
    this.options().find((option) => option.value === this.value()),
  );

  constructor() {
    effect(() => {
      const option = this.selected();
      if (option && !this.open()) this.query.set(option.label);
      if (!this.value() && !this.open()) this.query.set('');
    });

    effect(() => {
      const optionCount = this.options().length;
      this.activeIndex.update((current) => (current < optionCount ? current : -1));
    });

    effect((onCleanup) => {
      const closeOnOutsideClick = (event: PointerEvent) => {
        if (!this.host.nativeElement.contains(event.target as Node)) this.open.set(false);
      };
      document.addEventListener('pointerdown', closeOnOutsideClick);
      onCleanup(() => document.removeEventListener('pointerdown', closeOnOutsideClick));
    });
  }

  protected onFocus() {
    this.open.set(true);
    this.searchQuery.emit('');
  }

  protected onInput(value: string) {
    this.query.set(value);
    this.open.set(true);
    this.activeIndex.set(-1);
    if (this.value()) this.valueChange.emit('');
    if (this.searchTimer) clearTimeout(this.searchTimer);
    this.searchTimer = setTimeout(() => this.searchQuery.emit(value), 250);
  }

  protected choose(option: SearchSelectOption) {
    this.valueChange.emit(option.value);
    this.query.set(option.label);
    this.open.set(false);
  }

  protected moveActiveOption(offset: number) {
    if (this.options().length === 0) return;
    this.open.set(true);
    this.activeIndex.update((current) => {
      const next = current + offset;
      if (next < 0) return this.options().length - 1;
      if (next >= this.options().length) return 0;
      return next;
    });
  }

  protected onKeyDown(event: KeyboardEvent) {
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      this.moveActiveOption(1);
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      this.moveActiveOption(-1);
    } else if (event.key === 'Enter' && this.open()) {
      const option = this.options()[this.activeIndex()];
      if (option) {
        event.preventDefault();
        this.choose(option);
      }
    } else if (event.key === 'Escape') {
      this.open.set(false);
    }
  }

  protected inputClass = computed(() =>
    classes(
      'block h-10 w-full rounded-md border bg-white px-3 pr-10 text-sm text-gray-900 outline-none transition',
      this.error()
        ? 'border-red-300 focus:border-red-500 focus:ring-2 focus:ring-red-100'
        : 'border-gray-300 focus:border-gray-950 focus:ring-2 focus:ring-gray-100',
    ),
  );

  protected optionClass(option: SearchSelectOption, index: number) {
    return classes(
      'block w-full rounded px-3 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none',
      (option.value === this.value() || index === this.activeIndex()) && 'bg-gray-100',
    );
  }
}
