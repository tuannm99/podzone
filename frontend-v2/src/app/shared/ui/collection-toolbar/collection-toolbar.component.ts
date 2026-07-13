import { Component, input, linkedSignal, output } from '@angular/core';
import { Button } from '../button/button.component';
import type { CollectionSortDirection } from '../../services/collection-types';

export type CollectionSortOption = {
  label: string;
  value: string;
};

@Component({
  selector: 'app-collection-toolbar',
  imports: [Button],
  templateUrl: './collection-toolbar.component.html',
})
export class CollectionToolbar {
  search = input.required<string>();
  searchPlaceholder = input.required<string>();
  sortBy = input.required<string>();
  sortDirection = input.required<CollectionSortDirection>();
  pageSize = input.required<number>();
  sortOptions = input.required<CollectionSortOption[]>();

  searchChange = output<string>();
  sortByChange = output<string>();
  sortDirectionChange = output<CollectionSortDirection>();
  pageSizeChange = output<number>();

  protected readonly pageSizeOptions = [5, 10, 20, 50];
  // linkedSignal, not signal(this.search()) — required inputs aren't
  // readable at construction time (NG8118), and linkedSignal is the
  // supported pattern for a writable local draft seeded from (and
  // re-synced when) a reactive source changes. This differs slightly from
  // Solid's createSignal(props.search), which only reads the initial
  // value and never re-syncs — accepted as the more correct behavior
  // (the toolbar's draft should reflect an externally-reset search).
  protected searchDraft = linkedSignal(() => this.search());

  protected submitSearch(event: SubmitEvent) {
    event.preventDefault();
    this.searchChange.emit(this.searchDraft().trim());
  }

  protected onSearchInput(event: Event) {
    this.searchDraft.set((event.target as HTMLInputElement).value);
  }

  protected onSortByChange(event: Event) {
    this.sortByChange.emit((event.target as HTMLSelectElement).value);
  }

  protected onSortDirectionChange(event: Event) {
    this.sortDirectionChange.emit(
      (event.target as HTMLSelectElement).value as CollectionSortDirection,
    );
  }

  protected onPageSizeChange(event: Event) {
    this.pageSizeChange.emit(Number.parseInt((event.target as HTMLSelectElement).value, 10));
  }
}
