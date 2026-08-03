import { Component, TemplateRef, computed, inject, viewChild } from '@angular/core';

import { Button } from '../../../../shared/ui/button/button.component';
import { Card } from '../../../../shared/ui/card/card.component';
import { CollectionToolbar } from '../../../../shared/ui/collection-toolbar/collection-toolbar.component';
import {
  DataTable,
  type DataTableColumn,
} from '../../../../shared/ui/data-table/data-table.component';
import { EmptyBlock } from '../../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import type { PolicyInfo } from '../iam.types';
import { PoliciesService } from './policies.service';

// Policies section of /admin/iam (PZEP-0020) — read-only list of platform
// policies plus a selected policy's versions/attachments. No create/delete
// UI in this first cut (see PZEP-0020's Non-Goals).
@Component({
  selector: 'app-policies-page',
  imports: [Card, Button, CollectionToolbar, DataTable, EmptyBlock, ErrorAlert, LoadingBlock],
  providers: [PoliciesService],
  templateUrl: './policies-page.component.html',
  styleUrl: './policies-page.component.scss',
})
export class PoliciesPage {
  protected readonly service = inject(PoliciesService);

  private readonly selectCell =
    viewChild.required<TemplateRef<{ $implicit: PolicyInfo }>>('selectCell');

  protected readonly columns = computed<DataTableColumn<PolicyInfo>[]>(() => [
    { key: 'name', header: 'Name', cell: (row) => row.name },
    { key: 'description', header: 'Description', cell: (row) => row.description || '—' },
    { key: 'defaultVersion', header: 'Default Version', cell: (row) => row.defaultVersion || '—' },
    { key: 'select', header: '', cellTemplate: this.selectCell() },
  ]);

  protected onSearch(search: string) {
    this.service.updateSearch(search);
  }

  protected onPageChange(event: { pageIndex: number }) {
    this.service.updatePage(event.pageIndex + 1);
  }

  protected selectPolicy(policy: PolicyInfo) {
    this.service.selectPolicy(policy.name);
  }
}
