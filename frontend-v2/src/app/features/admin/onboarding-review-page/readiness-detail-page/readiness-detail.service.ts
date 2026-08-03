import { HttpClient } from '@angular/common/http';
import { Injectable, inject, resource, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../../../environments/environment';
import { httpErrorMessage } from '../../../../core/http/http-error';
import type { StoreReadiness } from '../../../onboarding/store/store.types';

// Route-scoped: owns the single-request admin readiness read (PZEP-0020).
@Injectable()
export class ReadinessDetailService {
  private readonly http = inject(HttpClient);

  private readonly _requestId = signal('');

  readonly readiness = resource({
    params: () => (this._requestId() ? { requestId: this._requestId() } : undefined),
    loader: async ({ params }) => {
      try {
        return await firstValueFrom(
          this.http.get<StoreReadiness>(
            `${environment.onboardingApiUrl}/onboarding/v1/admin/requests/${encodeURIComponent(params.requestId)}/readiness`,
          ),
        );
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load store readiness'));
      }
    },
  });

  setRequestId(requestId: string): void {
    this._requestId.set(requestId);
  }
}
