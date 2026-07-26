import { NgTemplateOutlet } from '@angular/common';
import { Component, TemplateRef, computed, input, output } from '@angular/core';
import { MatPaginatorModule, type PageEvent } from '@angular/material/paginator';
import { MatSortModule, type Sort } from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table';
import { EmptyBlock } from '../empty-block/empty-block.component';
import { ErrorAlert } from '../error-alert/error-alert.component';
import { LoadingBlock } from '../loading-block/loading-block.component';

// `cell` covers the common case (render a row field as text) so most
// columns need zero extra markup. `cellTemplate` is the escape hatch for
// anything richer (a status app-badge, an action app-button) — pass a
// `<ng-template let-row>` and the row is available as the template's
// implicit context. `sortable` puts a clickable mat-sort-header on this
// column instead of a separate sort dropdown elsewhere — sorting is
// server-driven (see `sortChange`), this only renders the arrow/click
// affordance.
export type DataTableColumn<T> = {
  key: string;
  header: string;
  cell?: (row: T) => string;
  cellTemplate?: TemplateRef<{ $implicit: T }>;
  sortable?: boolean;
};

export type DataTableSort = { active: string; direction: 'asc' | 'desc' };

@Component({
  selector: 'app-data-table',
  imports: [
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    NgTemplateOutlet,
    LoadingBlock,
    ErrorAlert,
    EmptyBlock,
  ],
  templateUrl: './data-table.component.html',
  styleUrl: './data-table.component.scss',
})
export class DataTable<T> {
  columns = input.required<DataTableColumn<T>[]>();
  rows = input.required<T[]>();
  loading = input(false);
  error = input('');
  emptyTitle = input('No results');
  emptyCopy = input('');

  // Column visibility is owned by the page (via app-column-picker, placed
  // in the same controls row as search/filters) - not this component.
  // Previously DataTable rendered its own toggle button/menu directly
  // above the table, isolated in its own row with no relationship to the
  // rest of the controls; extracting it to a sibling component lets the
  // page put the trigger wherever makes sense.
  hiddenColumns = input<ReadonlySet<string>>(new Set());

  // Pagination is fully controlled — this component renders whatever page
  // of `rows` it's given and only reports the user's requested page/size
  // back via `pageChange`; the consumer's feature service owns fetching
  // (matches ANGULAR_STYLE_GUIDE.md's "Collection state belongs in URL
  // query params" rule — this component never owns that state itself).
  totalCount = input(0);
  pageSize = input(20);
  pageIndex = input(0);

  // Sort is likewise fully controlled/server-driven: this only reflects
  // the current sort in the header arrow and reports clicks back.
  sortActive = input('');
  sortDirection = input<'asc' | 'desc'>('desc');

  pageChange = output<{ pageIndex: number; pageSize: number }>();
  sortChange = output<DataTableSort>();

  protected visibleColumns = computed(() =>
    this.columns().filter((c) => !this.hiddenColumns().has(c.key)),
  );
  protected displayedColumns = computed(() => this.visibleColumns().map((column) => column.key));

  protected onPage(event: PageEvent) {
    this.pageChange.emit({ pageIndex: event.pageIndex, pageSize: event.pageSize });
  }

  protected onSortChange(sort: Sort) {
    if (!sort.direction) return; // matSortDisableClear keeps this from firing, but guard anyway
    this.sortChange.emit({ active: sort.active, direction: sort.direction });
  }
}
