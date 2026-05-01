import { useParams } from '@tanstack/solid-router';
import { createEffect, createSignal } from 'solid-js';
import { TENANT_GQL_URL } from '../../../services/baseurl';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import { PageShell } from '../../components/common/PageShell';
import { EmptyBlock } from '../../components/common/Feedback';
import { Badge, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

export default function TenantHomePage() {
  const params = useParams({ from: '/t/$tenantId' });
  const [tenantReady, setTenantReady] = createSignal(
    tokenStorage.getActiveTenantID() === params().tenantId
  );

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
    setTenantReady(tokenStorage.getActiveTenantID() === params().tenantId);
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Tenant Workspace"
          title={`Tenant ${params().tenantId}`}
          copy="The tenant route now syncs the active tenant token before tenant-aware requests move into the GraphQL layer."
        />
      </Card>

      <div class="grid gap-4 md:grid-cols-3">
        <StatCard label="Tenant id" value={params().tenantId} />
        <StatCard label="Transport" value="GraphQL" />
        <StatCard
          label="Endpoint"
          value={TENANT_GQL_URL.replace(/^https?:\/\//, '')}
        />
      </div>

      <Card class="space-y-4">
        <SectionTitle
          title="Session context"
          subtitle="Requests now rely on the active tenant claim in the JWT. The local tenant value is kept only for navigation state."
        />
        <div class="flex flex-wrap gap-2">
          <Badge
            content={`active_tenant_id: ${tokenStorage.getActiveTenantID() || 'missing'}`}
            color={tenantReady() ? 'green' : 'yellow'}
          />
          <Badge
            content={`local tenant route: ${tenantStorage.getTenantID() || params().tenantId}`}
            color="indigo"
          />
          <Badge content="Authorization: Bearer ..." color="green" />
        </div>
        {!tenantReady() ? (
          <EmptyBlock
            title="Tenant token not ready"
            copy="The client could not confirm this tenant as the current active workspace yet."
          />
        ) : null}
      </Card>
    </PageShell>
  );
}
