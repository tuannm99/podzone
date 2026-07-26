import { Injectable, computed, inject, resource, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router } from '@angular/router';

import { StoreStorageService } from '../../../core/storage/store-storage.service';
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
import type { RoutedOrder } from './routed-order-list.types';

const ROUTED_ORDERS_QUERY = `
  query RoutedOrders($collection: CollectionInput) {
    routedOrders(collection: $collection) {
      items {
        id
        productTitle
        partner
        customerName
        status
        exceptionType
        exceptionStatus
        shipmentStatus
        operatorAssignee
        settlementStatus
        realizedMargin
        createdAt
        updatedAt
        shipmentSlaDueAt
        issueSlaDueAt
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

// Route-scoped: owns the `routedOrders` remote read. Store-scoped (unlike
// StoreListService's tenant-only `stores`) - storeId resolves from the
// query param first, falling back to StoreStorageService, same as
// TenantHomePage. Query state lives in the URL, not a plain signal, per
// ANGULAR_STYLE_GUIDE.md's "Collection state belongs in URL query params".
@Injectable()
export class RoutedOrderListService {
  private readonly graphql = inject(BackofficeGraphqlService);
  private readonly storeStorage = inject(StoreStorageService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private readonly tenantId = signal('');

  private readonly queryParams = toSignal(this.route.queryParamMap, { requireSync: true });

  readonly storeId = computed(() => {
    const tenantId = this.tenantId();
    const requested = this.queryParams().get('storeId') ?? '';
    return requested || this.storeStorage.getStoreId(tenantId);
  });

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
      const storeId = this.storeId();
      return tenantId && storeId ? { tenantId, storeId, query: this.query() } : undefined;
    },
    loader: async ({ params }) => {
      const result = await this.graphql.query<{
        routedOrders: { items?: RoutedOrder[]; pageInfo?: WirePageInfo };
      }>(
        ROUTED_ORDERS_QUERY,
        { collection: toGraphQLCollectionInput(params.query) },
        { tenantId: params.tenantId, storeId: params.storeId },
      );
      if (!result.success) throw new Error(result.message);
      const page: CollectionPage<RoutedOrder> = {
        items: result.data.routedOrders.items ?? [],
        pageInfo: normalizePageInfo(result.data.routedOrders.pageInfo, params.query),
      };
      return page;
    },
  });

  setTenantId(tenantId: string): void {
    this.tenantId.set(tenantId);
  }

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
