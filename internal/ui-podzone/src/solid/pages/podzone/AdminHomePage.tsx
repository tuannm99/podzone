import { createSignal } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import { EmptyBlock } from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import { Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

export default function AdminHomePage() {
  const [tenantId, setTenantId] = createSignal(tenantStorage.getTenantID());
  const user = tokenStorage.getUser();

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Control Room"
          title="Podzone admin is now running on the imported UI system."
          copy="The visual primitives, card surfaces, spacing, and navigation shell were cloned from the Judge Loop UI and adapted for Podzone routes."
        />
      </Card>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Session"
          value={tokenStorage.getToken() ? 'Active' : 'Missing'}
        />
        <StatCard
          label="User"
          value={user?.username || user?.email || 'Unknown'}
        />
        <StatCard
          label="Gateway"
          value={GW_API_URL.replace(/^https?:\/\//, '')}
        />
        <StatCard
          label="Tenant GQL"
          value={TENANT_GQL_URL.replace(/^https?:\/\//, '')}
        />
      </div>

      <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Tenant jump"
            subtitle="Store a tenant id locally, then move straight into the tenant workspace."
          />
          <div class="flex flex-col gap-3 sm:flex-row">
            <input
              class="block w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
              value={tenantId()}
              placeholder="tenant id"
              onInput={(event) => setTenantId(event.currentTarget.value)}
            />
            <Button
              href={tenantId().trim() ? `/t/${tenantId().trim()}` : undefined}
              disabled={!tenantId().trim()}
              onClick={() => tenantStorage.setTenantID(tenantId().trim())}
            >
              Open tenant
            </Button>
          </div>
          {!tenantId().trim() ? (
            <EmptyBlock
              title="No tenant selected"
              copy="Enter a tenant id above to open the tenant dashboard and orders view."
            />
          ) : null}
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Available routes"
            subtitle="The current shell is wired for these Podzone flows."
          />
          <div class="space-y-3 text-sm text-gray-600">
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Admin</p>
              <p class="mt-1">`/admin` and `/admin/settings`</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Tenant</p>
              <p class="mt-1">`/t/:tenantId` and `/t/:tenantId/orders`</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Auth</p>
              <p class="mt-1">`/auth/login` and `/auth/register`</p>
            </div>
          </div>
        </Card>
      </div>
    </PageShell>
  );
}
