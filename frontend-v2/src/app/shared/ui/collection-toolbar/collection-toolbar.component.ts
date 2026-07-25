import { Component, computed, input, output, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import type { CollectionSortDirection } from '../../services/collection-types';

export type CollectionSortOption = { label: string; value: string };

// Port of frontend/packages/shared/ui/components/common/CollectionToolbar.tsx.
// Page size is deliberately not handled here - app-data-table's mat-paginator
// already owns page-size selection, so this component only owns search +
// sort, per ADR-0008's collection-screen pattern.
@Component({
  selector: 'app-collection-toolbar',
  imports: [MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule, MatIconModule],
  templateUrl: './collection-toolbar.component.html',
  styleUrl: './collection-toolbar.component.scss',
})
export class CollectionToolbar {
  search = input('');
  searchPlaceholder = input('Search');
  sortBy = input('');
  sortDirection = input<CollectionSortDirection>('SORT_DIRECTION_DESC');
  sortOptions = input.required<CollectionSortOption[]>();

  searchChange = output<string>();
  sortByChange = output<string>();
  sortDirectionChange = output<CollectionSortDirection>();

  protected searchDraft = signal(this.search());
  protected isAscending = computed(() => this.sortDirection() === 'SORT_DIRECTION_ASC');

  protected onSubmit(event: SubmitEvent) {
    event.preventDefault();
    this.searchChange.emit(this.searchDraft().trim());
  }

  protected onSortByChange(value: string) {
    this.sortByChange.emit(value);
  }

  protected onSortDirectionChange(value: string) {
    this.sortDirectionChange.emit(value as CollectionSortDirection);
  }
}
