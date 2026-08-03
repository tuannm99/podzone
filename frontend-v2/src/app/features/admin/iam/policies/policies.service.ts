import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, resource, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../../../environments/environment';
import { httpErrorMessage } from '../../../../core/http/http-error';
import {
  normalizePageInfo,
  toCollectionParams,
  type CollectionPage,
  type CollectionQuery,
  type CollectionSortDirection,
  type WirePageInfo,
} from '../../../../shared/services/collection-types';
import type { PolicyAttachmentInfo, PolicyInfo, PolicyVersionInfo } from '../iam.types';

const DETAIL_QUERY: CollectionQuery = { page: 1, pageSize: 50, sortBy: 'createdAt' };

// Route-scoped: owns the `policies` list + selected policy's
// versions/attachments (PZEP-0020) - scope fixed to 'platform' for this
// first cut (the platform admin's own managed policies), per ported
// frontend/packages/shared/services/iam/policies.ts. Read-only: no
// create/delete-policy UI yet — see PZEP-0020's Non-Goals.
@Injectable()
export class PoliciesService {
  private readonly http = inject(HttpClient);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private readonly _selectedPolicyName = signal('');
  readonly selectedPolicyName = this._selectedPolicyName.asReadonly();

  private readonly queryParams = toSignal(this.route.queryParamMap, { requireSync: true });

  readonly query = computed<CollectionQuery>(() => {
    const params = this.queryParams();
    return {
      page: Number(params.get('page') ?? '1') || 1,
      pageSize: Number(params.get('pageSize') ?? '10') || 10,
      search: params.get('search') ?? undefined,
      sortBy: 'name',
      sortDirection: 'SORT_DIRECTION_ASC' as CollectionSortDirection,
    };
  });

  readonly policies = resource({
    params: () => this.query(),
    loader: async ({ params }) => {
      try {
        const response = await firstValueFrom(
          this.http.get<{ policies?: PolicyInfo[]; pageInfo?: WirePageInfo }>(
            `${environment.apiBaseUrl}/auth/v1/iam/policies`,
            { params: { ...toCollectionParams(params), scope: 'platform' } },
          ),
        );
        const page: CollectionPage<PolicyInfo> = {
          items: response.policies ?? [],
          pageInfo: normalizePageInfo(response.pageInfo, params),
        };
        return page;
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load policies'));
      }
    },
  });

  readonly versions = resource({
    params: () => (this._selectedPolicyName() ? { name: this._selectedPolicyName() } : undefined),
    loader: async ({ params }) => {
      try {
        const response = await firstValueFrom(
          this.http.get<{ versions?: PolicyVersionInfo[] }>(
            `${environment.apiBaseUrl}/auth/v1/iam/policies/${encodeURIComponent(params.name)}/versions`,
            { params: { ...toCollectionParams(DETAIL_QUERY), scope: 'platform' } },
          ),
        );
        return response.versions ?? [];
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load policy versions'));
      }
    },
  });

  readonly attachments = resource({
    params: () => (this._selectedPolicyName() ? { name: this._selectedPolicyName() } : undefined),
    loader: async ({ params }) => {
      try {
        const response = await firstValueFrom(
          this.http.get<{ attachments?: PolicyAttachmentInfo[] }>(
            `${environment.apiBaseUrl}/auth/v1/iam/policies/${encodeURIComponent(params.name)}/attachments`,
            { params: { ...toCollectionParams(DETAIL_QUERY), scope: 'platform' } },
          ),
        );
        return response.attachments ?? [];
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load policy attachments'));
      }
    },
  });

  selectPolicy(name: string): void {
    this._selectedPolicyName.set(name);
  }

  updateSearch(search: string): void {
    void this.router.navigate([], {
      queryParams: { search: search || null, page: 1 },
      queryParamsHandling: 'merge',
    });
  }

  updatePage(page: number): void {
    void this.router.navigate([], { queryParams: { page }, queryParamsHandling: 'merge' });
  }
}
