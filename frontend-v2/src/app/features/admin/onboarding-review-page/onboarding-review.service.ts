import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, resource, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { httpErrorMessage } from '../../../core/http/http-error';
import {
  emptyPageInfo,
  toCollectionParams,
  type CollectionFilter,
  type CollectionPage,
  type CollectionQuery,
  type CollectionSortDirection,
  type PageInfo,
} from '../../../shared/services/collection-types';
import { environment } from '../../../../environments/environment';
import type { StoreReadiness, StoreRequest } from '../../onboarding/store/store.types';
import { toAdminStoreRequest, type AdminStoreRequest } from './onboarding-review.types';

type ActionResult = { success: boolean; message: string };

function parseFilters(raw: string | null): CollectionFilter[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? (parsed as CollectionFilter[]) : [];
  } catch {
    return [];
  }
}

// Route-scoped: owns the cross-tenant `/admin/requests` remote read
// (PZEP-0020) - calls internal/onboarding's own REST surface directly
// (not the backoffice GraphQL layer StoreListService uses), since these
// admin routes live on onboarding's HTTP server. Query state lives in the
// URL, per ANGULAR_STYLE_GUIDE.md's "Collection state belongs in URL query
// params" rule.
@Injectable()
export class OnboardingReviewService {
  private readonly http = inject(HttpClient);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private readonly _actioningId = signal('');
  readonly actioningId = this._actioningId.asReadonly();

  private readonly queryParams = toSignal(this.route.queryParamMap, { requireSync: true });

  readonly query = computed<CollectionQuery>(() => {
    const params = this.queryParams();
    return {
      page: Number(params.get('page') ?? '1') || 1,
      pageSize: Number(params.get('pageSize') ?? '20') || 20,
      search: params.get('search') ?? undefined,
      sortBy: params.get('sortBy') ?? 'updatedAt',
      sortDirection:
        (params.get('sortDirection') as CollectionSortDirection | null) ?? 'SORT_DIRECTION_DESC',
      filters: parseFilters(params.get('filters')),
    };
  });

  readonly page = resource({
    params: () => this.query(),
    loader: async ({ params }) => {
      let response: { items?: StoreRequest[]; pageInfo?: PageInfo };
      try {
        response = await firstValueFrom(
          this.http.get<{ items?: StoreRequest[]; pageInfo?: PageInfo }>(
            `${environment.onboardingApiUrl}/onboarding/v1/admin/requests`,
            { params: toCollectionParams(params) },
          ),
        );
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load store requests'));
      }
      const page: CollectionPage<AdminStoreRequest> = {
        items: (response.items ?? []).map(toAdminStoreRequest),
        pageInfo: response.pageInfo ?? emptyPageInfo(params),
      };
      return page;
    },
  });

  updateQuery(patch: Partial<CollectionQuery>): void {
    const queryParams: Record<string, string | number | null> = {
      ...(patch.page ? {} : { page: 1 }),
    };
    if (patch.page !== undefined) queryParams['page'] = patch.page;
    if (patch.pageSize !== undefined) queryParams['pageSize'] = patch.pageSize;
    if (patch.search !== undefined) queryParams['search'] = patch.search || null;
    if (patch.sortBy !== undefined) queryParams['sortBy'] = patch.sortBy;
    if (patch.sortDirection !== undefined) queryParams['sortDirection'] = patch.sortDirection;
    if (patch.filters !== undefined) {
      queryParams['filters'] = patch.filters.length > 0 ? JSON.stringify(patch.filters) : null;
    }
    void this.router.navigate([], { queryParams, queryParamsHandling: 'merge' });
  }

  async approve(id: string): Promise<ActionResult> {
    return this.transition(id, 'approve', 'Failed to approve store request');
  }

  async reject(id: string): Promise<ActionResult> {
    return this.transition(id, 'reject', 'Failed to reject store request');
  }

  // Approve/Reject reuse the existing, unchanged tenant-scoped endpoints
  // (POST /onboarding/v1/requests/:id/approve|reject) - they're already
  // platform-permission-gated (store:approve) and don't require a tenant
  // header at all (ApproveStoreRequest/RejectStoreRequest authorize via
  // authorizeApproval only), so no admin-specific route was needed here
  // per PZEP-0020's Goals.
  private async transition(id: string, action: 'approve' | 'reject', fallback: string) {
    this._actioningId.set(id);
    try {
      await firstValueFrom(
        this.http.post(
          `${environment.onboardingApiUrl}/onboarding/v1/requests/${encodeURIComponent(id)}/${action}`,
          undefined,
        ),
      );
      this.page.reload();
      return { success: true, message: '' };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, fallback) };
    } finally {
      this._actioningId.set('');
    }
  }

  async getReadiness(
    id: string,
  ): Promise<{ success: true; data: StoreReadiness } | { success: false; message: string }> {
    try {
      const response = await firstValueFrom(
        this.http.get<StoreReadiness>(
          `${environment.onboardingApiUrl}/onboarding/v1/admin/requests/${encodeURIComponent(id)}/readiness`,
        ),
      );
      return { success: true, data: response };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'Failed to load store readiness') };
    }
  }
}
