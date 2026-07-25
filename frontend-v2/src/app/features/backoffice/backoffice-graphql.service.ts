import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { httpErrorMessage } from '../../core/http/http-error';

// Ported from frontend/packages/shared/services/backofficeGraphql.ts. Lives
// under features/backoffice (not shared/) because it is scoped to one
// domain's GraphQL endpoint and header contract - agent/ANGULAR_STYLE_GUIDE.md
// requires shared/ to stay domain-neutral.
type GraphQLErrorItem = {
  message?: string;
  extensions?: {
    code?: string;
    permission?: string;
    resource?: string;
  };
};

type GraphQLPayload<T> = {
  data?: T;
  errors?: GraphQLErrorItem[];
};

export type GraphQLResult<T> = { success: true; data: T } | { success: false; message: string };

function graphQLErrorMessage(error?: GraphQLErrorItem): string {
  if (!error) return 'GraphQL request failed';
  const code = error.extensions?.code;
  const permission = error.extensions?.permission;
  const resource = error.extensions?.resource;

  if (code === 'FORBIDDEN' && permission) {
    return resource && resource !== '*'
      ? `Missing permission: ${permission} on ${resource}`
      : `Missing permission: ${permission}`;
  }
  if (code === 'UNAUTHENTICATED') {
    return error.message || 'Authentication is required.';
  }
  return error.message || 'GraphQL request failed';
}

@Injectable({ providedIn: 'root' })
export class BackofficeGraphqlService {
  private readonly http = inject(HttpClient);

  // storeId is optional: some backoffice queries (e.g. `stores`, the
  // tenant-level store list) are tenant-scoped only and must omit
  // X-Store-ID entirely - matches frontend/'s `includeStoreHeader: false`
  // option on the equivalent SolidJS call.
  async query<T>(
    query: string,
    variables: Record<string, unknown> | undefined,
    scope: { tenantId: string; storeId?: string },
  ): Promise<GraphQLResult<T>> {
    const tenantId = scope.tenantId.trim();
    const storeId = scope.storeId?.trim() ?? '';
    if (!tenantId) {
      return { success: false, message: 'No active tenant to query' };
    }

    const headers: Record<string, string> = { 'X-Tenant-ID': tenantId };
    if (storeId) headers['X-Store-ID'] = storeId;

    try {
      const payload = await firstValueFrom(
        this.http.post<GraphQLPayload<T>>(
          environment.graphqlApiUrl,
          { query, variables },
          { headers },
        ),
      );

      if (payload.errors && payload.errors.length > 0) {
        return { success: false, message: graphQLErrorMessage(payload.errors[0]) };
      }
      if (!payload.data) {
        return { success: false, message: 'GraphQL request returned no data' };
      }
      return { success: true, data: payload.data };
    } catch (error) {
      return { success: false, message: httpErrorMessage(error, 'GraphQL request failed') };
    }
  }
}
