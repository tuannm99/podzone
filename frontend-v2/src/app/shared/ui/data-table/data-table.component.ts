import { NgTemplateOutlet } from '@angular/common';
import { Component, computed, input, output, TemplateRef } from '@angular/core';
import { MatPaginatorModule, type PageEvent } from '@angular/material/paginator';
import { MatTableModule } from '@angular/material/table';
import { EmptyBlock } from '../empty-block/empty-block.component';
import { ErrorAlert } from '../error-alert/error-alert.component';
import { LoadingBlock } from '../loading-block/loading-block.component';

// `cell` covers the common case (render a row field as text) so most
// columns need zero extra markup. `cellTemplate` is the escape hatch for
// anything richer (a status app-badge, an action app-button) — pass a
// `<ng-template let-row>` and the row is available as the template's
// implicit context.
export type DataTableColumn<T> = {
  key: string;
  header: string;
  cell?: (row: T) => string;
  cellTemplate?: TemplateRef<{ $implicit: T }>;
};

@Component({
  selector: 'app-data-table',
  imports: [
    MatTableModule,
    MatPaginatorModule,
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

  // Pagination is fully controlled — this component renders whatever page
  // of `rows` it's given and only reports the user's requested page/size
  // back via `pageChange`; the consumer's feature service owns fetching
  // (matches ANGULAR_STYLE_GUIDE.md's "Collection state belongs in URL
  // query params" rule — this component never owns that state itself).
  totalCount = input(0);
  pageSize = input(20);
  pageIndex = input(0);

  pageChange = output<{ pageIndex: number; pageSize: number }>();

  protected displayedColumns = computed(() => this.columns().map((column) => column.key));

  protected onPage(event: PageEvent) {
    this.pageChange.emit({ pageIndex: event.pageIndex, pageSize: event.pageSize });
  }
}
