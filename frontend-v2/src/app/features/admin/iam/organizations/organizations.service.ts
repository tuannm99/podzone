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
import type {
  CreateOrganizationPayload,
  OrganizationInfo,
  OrganizationMembership,
} from '../iam.types';

type ActionResult = { success: boolean; message: string };

// Route-scoped: owns the `organizations` + selected org's `members` remote
// reads (PZEP-0020) - calls the legacy IAMService REST surface via
// grpcgateway directly, the only IAM surface reachable from a browser (see
// PZEP-0020's Problem). Ported from frontend/packages/shared/services/
// iam/organizations.ts.
@Injectable()
export class OrganizationsService {
  private readonly http = inject(HttpClient);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  private readonly _selectedOrgId = signal('');
  private readonly _mutating = signal(false);

  readonly selectedOrgId = this._selectedOrgId.asReadonly();
  readonly mutating = this._mutating.asReadonly();

  private readonly queryParams = toSignal(this.route.queryParamMap, { requireSync: true });

  readonly query = computed<CollectionQuery>(() => {
    const params = this.queryParams();
    return {
      page: Number(params.get('page') ?? '1') || 1,
      pageSize: Number(params.get('pageSize') ?? '10') || 10,
      search: params.get('search') ?? undefined,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC' as CollectionSortDirection,
    };
  });

  readonly organizations = resource({
    params: () => this.query(),
    loader: async ({ params }) => {
      try {
        const response = await firstValueFrom(
          this.http.get<{
            organizations?: OrganizationInfo[];
            pageInfo?: WirePageInfo;
          }>(`${environment.apiBaseUrl}/auth/v1/iam/organizations`, {
            params: toCollectionParams(params),
          }),
        );
        const page: CollectionPage<OrganizationInfo> = {
          items: response.organizations ?? [],
          pageInfo: normalizePageInfo(response.pageInfo, params),
        };
        return page;
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load organizations'));
      }
    },
  });

  readonly members = resource({
    params: () => (this._selectedOrgId() ? { orgId: this._selectedOrgId() } : undefined),
    loader: async ({ params }) => {
      try {
        const query: CollectionQuery = { page: 1, pageSize: 50, sortBy: 'createdAt' };
        const response = await firstValueFrom(
          this.http.get<{
            memberships?: OrganizationMembership[];
            pageInfo?: WirePageInfo;
          }>(`${environment.apiBaseUrl}/auth/v1/iam/organizations/${params.orgId}/members`, {
            params: toCollectionParams(query),
          }),
        );
        return response.memberships ?? [];
      } catch (error) {
        throw new Error(httpErrorMessage(error, 'Failed to load organization members'));
      }
    },
  });

  selectOrg(orgId: string): void {
    this._selectedOrgId.set(orgId);
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

  async createOrganization(payload: CreateOrganizationPayload): Promise<ActionResult> {
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.post(`${environment.apiBaseUrl}/auth/v1/iam/organizations`, payload),
      );
      this.organizations.reload();
    }, 'Failed to create organization');
  }

  async addMember(orgId: string, userId: number, roleName: string): Promise<ActionResult> {
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.post(`${environment.apiBaseUrl}/auth/v1/iam/organizations/${orgId}/members`, {
          orgId,
          userId,
          roleName,
        }),
      );
      this.members.reload();
    }, 'Failed to add organization member');
  }

  async removeMember(orgId: string, userId: string): Promise<ActionResult> {
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.delete(
          `${environment.apiBaseUrl}/auth/v1/iam/organizations/${orgId}/members/${userId}`,
        ),
      );
      this.members.reload();
    }, 'Failed to remove organization member');
  }

  async attachTenant(orgId: string, tenantId: string): Promise<ActionResult> {
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.post(
          `${environment.apiBaseUrl}/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`,
          {
            orgId,
            tenantId,
          },
        ),
      );
    }, 'Failed to attach tenant to organization');
  }

  async detachTenant(orgId: string, tenantId: string): Promise<ActionResult> {
    return this.mutate(async () => {
      await firstValueFrom(
        this.http.delete(
          `${environment.apiBaseUrl}/auth/v1/iam/organizations/${orgId}/tenants/${tenantId}`,
        ),
      );
    }, 'Failed to detach tenant from organization');
  }

  private async mutate(action: () => Promise<void>, fallback: string): Promise<ActionResult> {
    this._mutating.set(true);
    try {
      await action();
      return { success: true, message: '' };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, fallback) };
    } finally {
      this._mutating.set(false);
    }
  }
}
