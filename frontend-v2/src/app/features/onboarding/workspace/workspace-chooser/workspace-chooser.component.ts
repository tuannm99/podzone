import { Component, computed, input, output } from '@angular/core';

import { EmptyBlock } from '../../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import {
  ListGroup,
  type ListGroupItem,
} from '../../../../shared/ui/list-group/list-group.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import type { TenantMembership } from '../workspace.types';

@Component({
  selector: 'app-workspace-chooser',
  imports: [LoadingBlock, ErrorAlert, EmptyBlock, ListGroup],
  templateUrl: './workspace-chooser.component.html',
})
export class WorkspaceChooser {
  memberships = input.required<TenantMembership[]>();
  loading = input(false);
  error = input('');
  selectedTenantId = input('');
  selectingTenantId = input('');

  select = output<string>();

  protected items = computed<ListGroupItem[]>(() =>
    this.memberships().map((membership) => ({
      label: membership.tenantId,
      description:
        membership.roleName +
        (this.selectingTenantId() === membership.tenantId ? ' · switching…' : ''),
      active: membership.tenantId === this.selectedTenantId(),
      onClick: () => this.select.emit(membership.tenantId),
    })),
  );
}
