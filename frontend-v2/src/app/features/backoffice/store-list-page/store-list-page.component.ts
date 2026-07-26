import { Component, TemplateRef, computed, inject, signal, viewChild } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import type { CollectionFilter } from '../../../shared/services/collection-types';
import { CollectionFilters } from '../../../shared/ui/collection-filters/collection-filters.component';
import { CollectionToolbar } from '../../../shared/ui/collection-toolbar/collection-toolbar.component';
import { ColumnPicker } from '../../../shared/ui/column-picker/column-picker.component';
import { Card } from '../../../shared/ui/card/card.component';
import {
  DataTable,
  type DataTableColumn,
  type DataTableSort,
} from '../../../shared/ui/data-table/data-table.component';
import { Badge } from '../../../shared/ui/badge/badge.component';
import { StoreListService } from './store-list.service';
import { storeFieldSpec, type StoreInfo } from './store-list.types';

// Route component: thin composition root, per
// agent/ANGULAR_STYLE_GUIDE.md. First real declarative-collection screen
// built on ADR-0008's CollectionFieldSpec pattern - future Backoffice
// collection screens (catalog, routing) follow this same shape. Sort is
// driven by app-data-table's clickable column headers, not a separate
// dropdown - see the DataTable/CollectionToolbar rework this page and
// PZEP-0014's orders screen share.
@Component({
  selector: 'app-store-list-page',
  imports: [Card, CollectionToolbar, CollectionFilters, ColumnPicker, DataTable, Badge],
  providers: [StoreListService],
  templateUrl: './store-list-page.component.html',
  styleUrl: './store-list-page.component.scss',
})
export class StoreListPage {
  private readonly route = inject(ActivatedRoute);
  protected readonly stores = inject(StoreListService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  protected readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected readonly fieldSpec = storeFieldSpec;

  protected readonly sortDirection = computed<'asc' | 'desc'>(() =>
    this.stores.query().sortDirection === 'SORT_DIRECTION_ASC' ? 'asc' : 'desc',
  );

  private readonly statusCell =
    viewChild.required<TemplateRef<{ $implicit: StoreInfo }>>('statusCell');

  protected columns = computed<DataTableColumn<StoreInfo>[]>(() => [
    { key: 'name', header: 'Name', cell: (row) => row.name, sortable: true },
    { key: 'description', header: 'Description', cell: (row) => row.description || '—' },
    { key: 'ownerId', header: 'Owner', cell: (row) => row.ownerId },
    { key: 'status', header: 'Status', cellTemplate: this.statusCell(), sortable: true },
    { key: 'createdAt', header: 'Created', cell: (row) => row.createdAt ?? '', sortable: true },
  ]);

  // Display preference, not collection state - deliberately not URL-synced
  // (matches app-data-table's previous hiddenColumns note, now owned here
  // since app-column-picker lives in this page's controls row).
  protected hiddenColumns = signal<ReadonlySet<string>>(new Set());

  constructor() {
    this.stores.setTenantId(this.tenantId());
  }

  protected onSearch(search: string) {
    this.stores.updateQuery({ search });
  }

  protected onSortChange(sort: DataTableSort) {
    this.stores.updateQuery({
      sortBy: sort.active,
      sortDirection: sort.direction === 'asc' ? 'SORT_DIRECTION_ASC' : 'SORT_DIRECTION_DESC',
    });
  }

  protected onFiltersChange(filters: CollectionFilter[]) {
    this.stores.updateQuery({ filters });
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
    this.stores.updateQuery({ page: event.pageIndex + 1, pageSize: event.pageSize });
  }
}
