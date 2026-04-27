import { useParams } from '@tanstack/solid-router';
import { createEffect } from 'solid-js';
import { TENANT_GQL_URL } from '../../../services/baseurl';
import { tenantStorage } from '../../../services/tenantStorage';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

export default function TenantHomePage() {
  const params = useParams({ from: '/t/$tenantId' });

  createEffect(() => {
    tenantStorage.setTenantID(params().tenantId);
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Tenant Workspace"
          title={`Tenant ${params().tenantId}`}
          copy="The tenant route preserves the existing route contract and local tenant header behavior, now wrapped in the imported Solid UI shell."
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
          title="Session headers"
          subtitle="Requests from tenant-aware clients include the current tenant id."
        />
        <div class="flex flex-wrap gap-2">
          <Badge content={`X-Tenant-ID: ${params().tenantId}`} color="indigo" />
          <Badge content="Authorization: Bearer ..." color="green" />
        </div>
      </Card>
    </PageShell>
  );
}
