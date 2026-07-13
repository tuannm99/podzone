import { Component, input } from '@angular/core';
import {
  CollectionToolbar,
  type CollectionSortOption,
} from '../collection-toolbar/collection-toolbar.component';
import {
  CollectionFilters,
  type CollectionFilterField,
} from '../collection-filters/collection-filters.component';
import { ErrorAlert } from '../error-alert/error-alert.component';
import { LoadingInline } from '../loading-inline/loading-inline.component';
import type { CollectionQuery } from '../../services/collection-types';

@Component({
  selector: 'app-collection-controls',
  imports: [CollectionToolbar, CollectionFilters, ErrorAlert, LoadingInline],
  templateUrl: './collection-controls.component.html',
})
export class CollectionControls {
  query = input.required<CollectionQuery>();
  loading = input.required<boolean>();
  error = input.required<string>();
  searchPlaceholder = input.required<string>();
  sortOptions = input.required<CollectionSortOption[]>();
  filterFields = input.required<CollectionFilterField[]>();
  updateQuery = input.required<(patch: Partial<CollectionQuery>) => void>();
}
