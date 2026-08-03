import { Component, TemplateRef, computed, inject, signal, viewChild } from '@angular/core';

import { Badge } from '../../../shared/ui/badge/badge.component';
import { Button } from '../../../shared/ui/button/button.component';
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
import { storeStatusLabel, storeStatusTone } from '../../onboarding/store/store.types';
import { OnboardingReviewService } from './onboarding-review.service';
import {
  adminStoreRequestFieldSpec,
  canApprove,
  canReject,
  type AdminStoreRequest,
} from './onboarding-review.types';

// Cross-tenant onboarding review screen (PZEP-0020, TASK-0004) - first
// declarative-collection screen under /admin, same ADR-0008 pattern as
// StoreListPage/RoutedOrderListPage. Approve/Reject call the existing,
// unchanged onboarding endpoints (see OnboardingReviewService.transition);
// readiness debug is a separate detail route, not an inline expand.
@Component({
  selector: 'app-onboarding-review-page',
  imports: [Card, CollectionToolbar, CollectionFilters, ColumnPicker, DataTable, Badge, Button],
  providers: [OnboardingReviewService],
  templateUrl: './onboarding-review-page.component.html',
  styleUrl: './onboarding-review-page.component.scss',
})
export class OnboardingReviewPage {
  protected readonly requests = inject(OnboardingReviewService);

  protected readonly fieldSpec = adminStoreRequestFieldSpec;
  protected readonly label = storeStatusLabel;
  protected readonly canApprove = canApprove;
  protected readonly canReject = canReject;

  protected readonly sortDirection = computed<'asc' | 'desc'>(() =>
    this.requests.query().sortDirection === 'SORT_DIRECTION_ASC' ? 'asc' : 'desc',
  );

  private readonly statusCell =
    viewChild.required<TemplateRef<{ $implicit: AdminStoreRequest }>>('statusCell');
  private readonly actionsCell =
    viewChild.required<TemplateRef<{ $implicit: AdminStoreRequest }>>('actionsCell');

  protected columns = computed<DataTableColumn<AdminStoreRequest>[]>(() => [
    { key: 'name', header: 'Name', cell: (row) => row.name, sortable: true },
    { key: 'workspaceId', header: 'Tenant', cell: (row) => row.workspaceId, sortable: true },
    { key: 'subdomain', header: 'Subdomain', cell: (row) => row.subdomain, sortable: true },
    { key: 'requestedBy', header: 'Requested By', cell: (row) => row.requestedBy, sortable: true },
    { key: 'status', header: 'Status', cellTemplate: this.statusCell(), sortable: true },
    { key: 'updatedAt', header: 'Updated', cell: (row) => row.updatedAt, sortable: true },
    { key: 'actions', header: '', cellTemplate: this.actionsCell() },
  ]);

  protected readonly tone = (row: AdminStoreRequest) => storeStatusTone(row.status);

  // Display preference, not collection state - deliberately not URL-synced.
  protected hiddenColumns = signal<ReadonlySet<string>>(new Set());

  protected onSearch(search: string) {
    this.requests.updateQuery({ search });
  }

  protected onSortChange(sort: DataTableSort) {
    this.requests.updateQuery({
      sortBy: sort.active,
      sortDirection: sort.direction === 'asc' ? 'SORT_DIRECTION_ASC' : 'SORT_DIRECTION_DESC',
    });
  }

  protected onFiltersChange(filters: CollectionFilter[]) {
    this.requests.updateQuery({ filters });
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
    this.requests.updateQuery({ page: event.pageIndex + 1, pageSize: event.pageSize });
  }

  protected async onApprove(row: AdminStoreRequest) {
    await this.requests.approve(row.id);
  }

  protected async onReject(row: AdminStoreRequest) {
    await this.requests.reject(row.id);
  }
}
