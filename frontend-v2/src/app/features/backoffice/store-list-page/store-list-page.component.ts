import { Component, TemplateRef, computed, inject, viewChild } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import type {
  CollectionFilter,
  CollectionSortDirection,
} from '../../../shared/services/collection-types';
import { CollectionFilters } from '../../../shared/ui/collection-filters/collection-filters.component';
import type { CollectionSortOption } from '../../../shared/ui/collection-toolbar/collection-toolbar.component';
import { CollectionToolbar } from '../../../shared/ui/collection-toolbar/collection-toolbar.component';
import { Card } from '../../../shared/ui/card/card.component';
import {
  DataTable,
  type DataTableColumn,
} from '../../../shared/ui/data-table/data-table.component';
import { Badge } from '../../../shared/ui/badge/badge.component';
import { StoreListService } from './store-list.service';
import { storeFieldSpec, type StoreInfo } from './store-list.types';

const SORT_OPTIONS: CollectionSortOption[] = storeFieldSpec
  .filter((field) => field.sortable)
  .map((field) => ({ label: field.label, value: field.field }));

// Route component: thin composition root, per
// agent/ANGULAR_STYLE_GUIDE.md. First real declarative-collection screen
// built on ADR-0008's CollectionFieldSpec pattern - future Backoffice
// collection screens (catalog, routing) follow this same shape.
@Component({
  selector: 'app-store-list-page',
  imports: [Card, CollectionToolbar, CollectionFilters, DataTable, Badge],
  providers: [StoreListService],
  templateUrl: './store-list-page.component.html',
  styleUrl: './store-list-page.component.scss',
})
export class StoreListPage {
  private readonly route = inject(ActivatedRoute);
  protected readonly stores = inject(StoreListService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  protected readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected readonly sortOptions = SORT_OPTIONS;
  protected readonly fieldSpec = storeFieldSpec;

  private readonly statusCell =
    viewChild.required<TemplateRef<{ $implicit: StoreInfo }>>('statusCell');

  protected columns = computed<DataTableColumn<StoreInfo>[]>(() => [
    { key: 'name', header: 'Name', cell: (row) => row.name },
    { key: 'description', header: 'Description', cell: (row) => row.description || '—' },
    { key: 'ownerId', header: 'Owner', cell: (row) => row.ownerId },
    { key: 'status', header: 'Status', cellTemplate: this.statusCell() },
    { key: 'createdAt', header: 'Created', cell: (row) => row.createdAt ?? '' },
  ]);

  constructor() {
    this.stores.setTenantId(this.tenantId());
  }

  protected onSearch(search: string) {
    this.stores.updateQuery({ search });
  }

  protected onSortByChange(sortBy: string) {
    this.stores.updateQuery({ sortBy });
  }

  protected onSortDirectionChange(sortDirection: CollectionSortDirection) {
    this.stores.updateQuery({ sortDirection });
  }

  protected onFiltersChange(filters: CollectionFilter[]) {
    this.stores.updateQuery({ filters });
  }

  protected onPageChange(event: { pageIndex: number; pageSize: number }) {
    this.stores.updateQuery({ page: event.pageIndex + 1, pageSize: event.pageSize });
  }
}
