import { Injectable, inject, resource, signal } from '@angular/core';

import { BackofficeGraphqlService } from '../backoffice-graphql.service';

type ProductSetupSnapshot = {
  drafts: { id: string }[];
  candidates: { id: string; status: string }[];
};

const SNAPSHOT_QUERY = `
  query ProductSetupSnapshot {
    productSetupSnapshot {
      drafts { id }
      candidates { id status }
    }
  }
`;

// Route-scoped: owns the one protected read this page makes, per
// agent/ANGULAR_STYLE_GUIDE.md's "one feature service owns remote reads"
// convention. Deliberately queries a subset of the fields frontend/'s
// SolidJS home page did - this page shows two counts, not the full ported
// dashboard (see PZEP-0011 Non-Goals).
@Injectable()
export class TenantHomeService {
  private readonly graphql = inject(BackofficeGraphqlService);

  private readonly _tenantId = signal('');
  private readonly _storeId = signal('');

  readonly snapshot = resource({
    params: () => {
      const tenantId = this._tenantId();
      const storeId = this._storeId();
      return tenantId && storeId ? { tenantId, storeId } : undefined;
    },
    loader: async ({ params }) => {
      const result = await this.graphql.query<{ productSetupSnapshot: ProductSetupSnapshot }>(
        SNAPSHOT_QUERY,
        undefined,
        params,
      );
      if (!result.success) throw new Error(result.message);
      return result.data.productSetupSnapshot;
    },
  });

  setScope(tenantId: string, storeId: string): void {
    this._tenantId.set(tenantId);
    this._storeId.set(storeId);
  }
}
