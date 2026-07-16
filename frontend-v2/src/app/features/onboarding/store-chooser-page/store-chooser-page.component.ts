import { Component, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import { Card } from '../../../shared/ui/card/card.component';
import { ToasterService } from '../../../shared/services/toaster.service';
import { OnboardingStoreService } from '../store/onboarding-store.service';
import { StoreChooser } from '../store/store-chooser/store-chooser.component';

@Component({
  selector: 'app-store-chooser-page',
  imports: [Card, StoreChooser],
  providers: [OnboardingStoreService],
  templateUrl: './store-chooser-page.component.html',
  styleUrl: './store-chooser-page.component.scss',
})
export class StoreChooserPage {
  private readonly route = inject(ActivatedRoute);
  private readonly toaster = inject(ToasterService);
  protected readonly store = inject(OnboardingStoreService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  protected readonly tenantId = computed(() => this.paramMap().get('tenantId') ?? '');

  protected newStoreName = signal('');

  constructor() {
    this.store.setTenantId(this.tenantId());
  }

  async onCreateStore() {
    const name = this.newStoreName();
    const result = await this.store.createStore(name);
    if (!result.success) {
      this.toaster.error(result.message);
      return;
    }
    this.newStoreName.set('');
    this.toaster.success(`Store "${name}" requested.`);
  }

  async onRetryStore(requestId: string) {
    const result = await this.store.retryStore(requestId);
    if (!result.success) {
      this.toaster.error(result.message);
      return;
    }
    this.toaster.success('Retry requested.');
  }
}
