import { Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';

import { Badge } from '../../../../shared/ui/badge/badge.component';
import { Button } from '../../../../shared/ui/button/button.component';
import { Card } from '../../../../shared/ui/card/card.component';
import { ErrorAlert } from '../../../../shared/ui/error-alert/error-alert.component';
import { LoadingBlock } from '../../../../shared/ui/loading-block/loading-block.component';
import { readinessTone, storeStatusLabel } from '../../../onboarding/store/store.types';
import { ReadinessDetailService } from './readiness-detail.service';

// Readiness debug detail route (PZEP-0020) - looks up one request's
// placement/readiness state regardless of which tenant it belongs to, via
// GetStoreReadinessAsAdmin (store:read_all), same as OnboardingReviewPage's
// list but for a single id reached from the "Readiness" row action.
@Component({
  selector: 'app-readiness-detail-page',
  imports: [Card, Button, Badge, LoadingBlock, ErrorAlert],
  providers: [ReadinessDetailService],
  templateUrl: './readiness-detail-page.component.html',
  styleUrl: './readiness-detail-page.component.scss',
})
export class ReadinessDetailPage {
  private readonly route = inject(ActivatedRoute);
  protected readonly service = inject(ReadinessDetailService);

  private readonly paramMap = toSignal(this.route.paramMap, { requireSync: true });
  protected readonly requestId = computed(() => this.paramMap().get('id') ?? '');

  protected readonly label = storeStatusLabel;

  protected readonly tone = computed(() => {
    const readiness = this.service.readiness.value();
    if (!readiness) return 'neutral' as const;
    return readinessTone(readiness.request_status, readiness.ui_state);
  });

  constructor() {
    this.service.setRequestId(this.requestId());
  }
}
