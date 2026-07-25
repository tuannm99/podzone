import { Injectable, computed, inject, resource, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router } from '@angular/router';

import {
  normalizePageInfo,
  toGraphQLCollectionInput,
  type CollectionFilter,
  type CollectionPage,
  type CollectionQuery,
  type CollectionSortDirection,
  type WirePageInfo,
} from '../../../shared/services/collection-types';
import { BackofficeGraphqlService } from '../backoffice-graphql.service';
import type { StoreInfo } from './store-list.types';

const STORES_QUERY = `
  query Stores($collection: CollectionInput) {
    stores(collection: $collection) {
      items {
        id
        name
        ownerId: owner_id
        isActive: is_active
        description
        status
        createdAt: created_at
        updatedAt: updated_at
      }
      pageInfo {
        total
        page
        pageSize
        totalPages
        hasNext
        hasPrevious
      }
    }
  }
`;

function parseFilters(raw: string | null): CollectionFilter[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? (parsed as CollectionFilter[]) : [];
  } catch {
    return [];
  }
}

// Route-scoped: owns the `stores` remote read, per
// agent/ANGULAR_STYLE_GUIDE.md's "one feature service owns remote reads"
// convention. `stores` is tenant-scoped only (no store header) - see
// BackofficeGraphqlService.query's optional storeId. Query state (page/
// search/sort/filters) lives in the URL, not a plain signal, per
// ANGULAR_STYLE_GUIDE.md's "Collection state belongs in URL query params".
@Injectable()
export class StoreListService {
  private readonly graphql = inject(BackofficeGraphqlService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly tenantId = signal('');

  private readonly queryParams = toSignal(this.route.queryParamMap, { requireSync: true });

  readonly query = computed<CollectionQuery>(() => {
    const params = this.queryParams();
    return {
      page: Number(params.get('page') ?? '1') || 1,
      pageSize: Number(params.get('pageSize') ?? '20') || 20,
      search: params.get('search') ?? undefined,
      sortBy: params.get('sortBy') ?? 'createdAt',
      sortDirection:
        (params.get('sortDirection') as CollectionSortDirection | null) ?? 'SORT_DIRECTION_DESC',
      filters: parseFilters(params.get('filters')),
    };
  });

  readonly page = resource({
    params: () => {
      const tenantId = this.tenantId();
      return tenantId ? { tenantId, query: this.query() } : undefined;
    },
    loader: async ({ params }) => {
      const result = await this.graphql.query<{
        stores: { items?: StoreInfo[]; pageInfo?: WirePageInfo };
      }>(
        STORES_QUERY,
        { collection: toGraphQLCollectionInput(params.query) },
        { tenantId: params.tenantId },
      );
      if (!result.success) throw new Error(result.message);
      const page: CollectionPage<StoreInfo> = {
        items: result.data.stores.items ?? [],
        pageInfo: normalizePageInfo(result.data.stores.pageInfo, params.query),
      };
      return page;
    },
  });

  setTenantId(tenantId: string): void {
    this.tenantId.set(tenantId);
  }

  // Page/search/filter/sort changes reset page to 1; page changes (the
  // caller passes `page` explicitly) preserve the requested page - matches
  // ANGULAR_STYLE_GUIDE.md's Collections rule.
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
}
