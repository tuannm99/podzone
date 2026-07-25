import { Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import { StoreStorageService } from '../../../core/storage/store-storage.service';
import { Badge } from '../../../shared/ui/badge/badge.component';
import { Card } from '../../../shared/ui/card/card.component';
import { EmptyBlock } from '../../../shared/ui/empty-block/empty-block.component';
import { ErrorAlert } from '../../../shared/ui/error-alert/error-alert.component';
import { LoadingBlock } from '../../../shared/ui/loading-block/loading-block.component';
import { TenantHomeService } from './tenant-home.service';

@Component({
  selector: 'app-tenant-home-page',
  imports: [Card, Badge, LoadingBlock, ErrorAlert, EmptyBlock],
  providers: [TenantHomeService],
  templateUrl: './tenant-home-page.component.html',
  styleUrl: './tenant-home-page.component.scss',
})
export class TenantHomePage {
  private readonly route = inject(ActivatedRoute);
  private readonly storeStorage = inject(StoreStorageService);
  protected readonly home = inject(TenantHomeService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  private readonly queryParamMap = toSignal(this.route.queryParamMap, { requireSync: true });

  protected readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');
  // Query param wins (fresh handoff from M3); falls back to the persisted
  // active store for direct/reload navigation - mirrors frontend/'s
  // TenantWorkspaceProvider `initialStoreID` pattern.
  protected readonly storeId = computed(() => {
    const tenantId = this.tenantId();
    const requested = this.queryParamMap().get('storeId') ?? '';
    return requested || this.storeStorage.getStoreId(tenantId);
  });

  constructor() {
    const tenantId = this.tenantId();
    const storeId = this.storeId();
    if (tenantId && storeId) this.storeStorage.setStoreId(tenantId, storeId);
    this.home.setScope(tenantId, storeId);
  }
}
