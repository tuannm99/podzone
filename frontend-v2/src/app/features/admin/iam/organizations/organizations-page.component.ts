import { Component, TemplateRef, computed, inject, signal, viewChild } from '@angular/core';

import { Button } from '../../../../shared/ui/button/button.component';
import { Card } from '../../../../shared/ui/card/card.component';
import { CollectionToolbar } from '../../../../shared/ui/collection-toolbar/collection-toolbar.component';
import {
  DataTable,
  type DataTableColumn,
} from '../../../../shared/ui/data-table/data-table.component';
import { EmptyBlock } from '../../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import { InputField } from '../../../../shared/ui/input-field/input-field.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import type { OrganizationInfo, OrganizationMembership } from '../iam.types';
import { OrganizationsService } from './organizations.service';

// Organizations section of /admin/iam (PZEP-0020) — list + create
// organizations, select one to manage its members and attach/detach
// tenants by id. No tenant-directory listing here — see PZEP-0020's Open
// Questions: the legacy IAMService has no RPC to list tenants at all.
@Component({
  selector: 'app-organizations-page',
  imports: [
    Card,
    Button,
    CollectionToolbar,
    DataTable,
    EmptyBlock,
    ErrorAlert,
    LoadingBlock,
    InputField,
  ],
  providers: [OrganizationsService],
  templateUrl: './organizations-page.component.html',
  styleUrl: './organizations-page.component.scss',
})
export class OrganizationsPage {
  protected readonly service = inject(OrganizationsService);

  private readonly selectCell =
    viewChild.required<TemplateRef<{ $implicit: OrganizationInfo }>>('selectCell');

  protected readonly columns = computed<DataTableColumn<OrganizationInfo>[]>(() => [
    { key: 'name', header: 'Name', cell: (row) => row.name },
    { key: 'slug', header: 'Slug', cell: (row) => row.slug },
    { key: 'createdAt', header: 'Created', cell: (row) => row.createdAt ?? '' },
    { key: 'select', header: '', cellTemplate: this.selectCell() },
  ]);

  protected readonly newOrgName = signal('');
  protected readonly newOrgSlug = signal('');
  protected readonly createError = signal('');

  protected readonly newMemberUserId = signal('');
  protected readonly newMemberRoleName = signal('');
  protected readonly memberError = signal('');

  protected readonly tenantIdInput = signal('');
  protected readonly tenantError = signal('');

  protected onSearch(search: string) {
    this.service.updateSearch(search);
  }

  protected onPageChange(event: { pageIndex: number }) {
    this.service.updatePage(event.pageIndex + 1);
  }

  protected selectOrg(org: OrganizationInfo) {
    this.service.selectOrg(org.id);
  }

  protected async onCreateOrg(event: SubmitEvent) {
    event.preventDefault();
    const name = this.newOrgName().trim();
    const slug = this.newOrgSlug().trim();
    if (!name || !slug) return;

    const result = await this.service.createOrganization({ name, slug });
    if (result.success) {
      this.newOrgName.set('');
      this.newOrgSlug.set('');
      this.createError.set('');
    } else {
      this.createError.set(result.message);
    }
  }

  protected async onAddMember(event: SubmitEvent) {
    event.preventDefault();
    const orgId = this.service.selectedOrgId();
    const userId = Number(this.newMemberUserId());
    const roleName = this.newMemberRoleName().trim();
    if (!orgId || !userId || !roleName) return;

    const result = await this.service.addMember(orgId, userId, roleName);
    if (result.success) {
      this.newMemberUserId.set('');
      this.newMemberRoleName.set('');
      this.memberError.set('');
    } else {
      this.memberError.set(result.message);
    }
  }

  protected async onRemoveMember(member: OrganizationMembership) {
    const orgId = this.service.selectedOrgId();
    if (!orgId) return;
    const result = await this.service.removeMember(orgId, member.userId);
    if (!result.success) this.memberError.set(result.message);
  }

  protected async onAttachTenant(event: SubmitEvent) {
    event.preventDefault();
    const orgId = this.service.selectedOrgId();
    const tenantId = this.tenantIdInput().trim();
    if (!orgId || !tenantId) return;

    const result = await this.service.attachTenant(orgId, tenantId);
    if (result.success) {
      this.tenantIdInput.set('');
      this.tenantError.set('');
    } else {
      this.tenantError.set(result.message);
    }
  }

  protected async onDetachTenant() {
    const orgId = this.service.selectedOrgId();
    const tenantId = this.tenantIdInput().trim();
    if (!orgId || !tenantId) return;

    const result = await this.service.detachTenant(orgId, tenantId);
    if (result.success) {
      this.tenantIdInput.set('');
      this.tenantError.set('');
    } else {
      this.tenantError.set(result.message);
    }
  }
}
