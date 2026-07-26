import { Component, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import { Card } from '../../../shared/ui/card/card.component';
import { CollectionFilters } from '../../../shared/ui/collection-filters/collection-filters.component';
import { CollectionToolbar } from '../../../shared/ui/collection-toolbar/collection-toolbar.component';
import { ColumnPicker } from '../../../shared/ui/column-picker/column-picker.component';
import {
  DataTable,
  type DataTableColumn,
  type DataTableSort,
} from '../../../shared/ui/data-table/data-table.component';
import type { CollectionFilter } from '../../../shared/services/collection-types';
import { RoutedOrderListService } from './routed-order-list.service';
import { routedOrderFieldSpec, type RoutedOrder } from './routed-order-list.types';

const COLUMNS: DataTableColumn<RoutedOrder>[] = [
  { key: 'productTitle', header: 'Product', cell: (row) => row.productTitle, sortable: true },
  { key: 'partner', header: 'Partner', cell: (row) => row.partner },
  { key: 'customerName', header: 'Customer', cell: (row) => row.customerName },
  { key: 'status', header: 'Status', cell: (row) => row.status, sortable: true },
  { key: 'exceptionStatus', header: 'Exception', cell: (row) => row.exceptionStatus || '—' },
  { key: 'updatedAt', header: 'Updated', cell: (row) => row.updatedAt, sortable: true },
];

// Second declarative-collection screen — see PZEP-0014, builds on
// PZEP-0012's pattern. Store-scoped (routedOrders needs X-Store-ID). Sort
// is driven by app-data-table's clickable column headers.
@Component({
  selector: 'app-routed-order-list-page',
  imports: [Card, CollectionToolbar, CollectionFilters, ColumnPicker, DataTable],
  providers: [RoutedOrderListService],
  templateUrl: './routed-order-list-page.component.html',
  styleUrl: './routed-order-list-page.component.scss',
})
export class RoutedOrderListPage {
  private readonly route = inject(ActivatedRoute);
  protected readonly orders = inject(RoutedOrderListService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  protected readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected readonly fieldSpec = routedOrderFieldSpec;
  protected readonly columns = COLUMNS;

  protected readonly sortDirection = computed<'asc' | 'desc'>(() =>
    this.orders.query().sortDirection === 'SORT_DIRECTION_ASC' ? 'asc' : 'desc',
  );

  // Display preference, not collection state - deliberately not URL-synced.
  protected hiddenColumns = signal<ReadonlySet<string>>(new Set());

  constructor() {
    this.orders.setTenantId(this.tenantId());
  }

  protected onSearch(search: string) {
    this.orders.updateQuery({ search });
  }

  protected onSortChange(sort: DataTableSort) {
    this.orders.updateQuery({
      sortBy: sort.active,
      sortDirection: sort.direction === 'asc' ? 'SORT_DIRECTION_ASC' : 'SORT_DIRECTION_DESC',
    });
  }

  protected onFiltersChange(filters: CollectionFilter[]) {
    this.orders.updateQuery({ filters });
  }

  protected onToggleColumn(key: string) {
    this.hiddenColumns.update((current) => {
      const next = new Set(current);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }

  protected onPageChange(event: { pageIndex: number; pageSize: number }) {
    this.orders.updateQuery({ page: event.pageIndex + 1, pageSize: event.pageSize });
  }
}
