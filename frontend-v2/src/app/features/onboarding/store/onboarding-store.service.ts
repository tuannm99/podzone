import { HttpClient } from '@angular/common/http';
import { DestroyRef, Injectable, inject, resource, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../../environments/environment';
import { httpErrorMessage } from '../../../core/http/http-error';
import { activeProvisioningStatuses, type StoreReadiness, type StoreRequest } from './store.types';

const READINESS_POLL_INTERVAL_MS = 10_000;

type ActionResult = { success: boolean; message: string };

function tenantHeaders(tenantId: string) {
  return { 'X-Tenant-ID': tenantId.trim() };
}

// Ported from frontend/apps/onboarding/src/pages/admin-home/presentation.ts.
function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

@Injectable()
export class OnboardingStoreService {
  private readonly http = inject(HttpClient);
  private readonly destroyRef = inject(DestroyRef);

  private readonly _tenantId = signal('');
  private readonly _creating = signal(false);
  private readonly _retryingId = signal('');

  readonly creating = this._creating.asReadonly();
  readonly retryingId = this._retryingId.asReadonly();

  readonly requests = resource({
    params: () => (this._tenantId() ? { tenantId: this._tenantId() } : undefined),
    loader: async ({ params }) => {
      const result = await this.listStoreRequests(params.tenantId);
      if (!result.success) throw new Error(result.message);
      return result.data;
    },
  });

  // Recomputed on every `requests` reload (including the 10s poll below),
  // matching Solid's createStoreReadinessViewModel.ts. `request.status`
  // alone is not authoritative for whether a store can be opened - see
  // `isReadyToOpen` in store.types.ts.
  readonly readiness = resource({
    params: () => {
      const tenantId = this._tenantId();
      const requests = this.requests.value();
      if (!tenantId || !requests || requests.length === 0) return undefined;
      return { tenantId, ids: requests.map((request) => request.id) };
    },
    loader: async ({ params }) => {
      const entries = await Promise.all(
        params.ids.map(async (id) => {
          const result = await this.getStoreReadiness(params.tenantId, id);
          return [id, result.success ? result.data : undefined] as const;
        }),
      );
      return Object.fromEntries(entries) as Record<string, StoreReadiness | undefined>;
    },
  });

  constructor() {
    const timer = setInterval(() => {
      const requests = this.requests.value();
      const hasActiveProvisioning = (requests ?? []).some((request) =>
        activeProvisioningStatuses.includes(request.status),
      );
      if (hasActiveProvisioning && !this.requests.isLoading()) {
        this.requests.reload();
      }
    }, READINESS_POLL_INTERVAL_MS);
    this.destroyRef.onDestroy(() => clearInterval(timer));
  }

  setTenantId(tenantId: string): void {
    this._tenantId.set(tenantId);
  }

  async createStore(name: string): Promise<ActionResult> {
    const tenantId = this._tenantId();
    const trimmedName = name.trim();
    if (!tenantId || !trimmedName) {
      return { success: false, message: 'Store name is required' };
    }

    this._creating.set(true);
    try {
      await firstValueFrom(
        this.http.post<StoreRequest>(
          `${environment.onboardingApiUrl}/onboarding/v1/requests`,
          { name: trimmedName, subdomain: slugify(trimmedName) },
          { headers: tenantHeaders(tenantId) },
        ),
      );
      this.requests.reload();
      return { success: true, message: '' };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'Failed to create store') };
    } finally {
      this._creating.set(false);
    }
  }

  async retryStore(requestId: string): Promise<ActionResult> {
    const tenantId = this._tenantId();
    if (!tenantId) {
      return { success: false, message: 'No active workspace' };
    }

    this._retryingId.set(requestId);
    try {
      await firstValueFrom(
        this.http.post(
          `${environment.onboardingApiUrl}/onboarding/v1/requests/${encodeURIComponent(requestId)}/retry`,
          undefined,
          { headers: tenantHeaders(tenantId) },
        ),
      );
      this.requests.reload();
      return { success: true, message: '' };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'Failed to retry store') };
    } finally {
      this._retryingId.set('');
    }
  }

  private async listStoreRequests(
    tenantId: string,
  ): Promise<{ success: true; data: StoreRequest[] } | { success: false; message: string }> {
    try {
      const response = await firstValueFrom(
        this.http.get<{ items?: StoreRequest[] }>(
          `${environment.onboardingApiUrl}/onboarding/v1/requests`,
          { headers: tenantHeaders(tenantId) },
        ),
      );
      return { success: true, data: response.items ?? [] };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'Failed to load stores') };
    }
  }

  private async getStoreReadiness(
    tenantId: string,
    requestId: string,
  ): Promise<{ success: true; data: StoreReadiness } | { success: false; message: string }> {
    try {
      const response = await firstValueFrom(
        this.http.get<StoreReadiness>(
          `${environment.onboardingApiUrl}/onboarding/v1/requests/${encodeURIComponent(requestId)}/readiness`,
          { headers: tenantHeaders(tenantId) },
        ),
      );
      return { success: true, data: response };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'Failed to load store readiness') };
    }
  }
}
