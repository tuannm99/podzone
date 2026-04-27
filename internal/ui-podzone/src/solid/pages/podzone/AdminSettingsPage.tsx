import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Button, Card } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

export default function AdminSettingsPage() {
  const hasToken = Boolean(tokenStorage.getToken());
  const tenantId = tenantStorage.getTenantID();

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Settings"
          title="Inspect local frontend state."
          copy="This page keeps the imported card and typography system while exposing the same Podzone client configuration already used by the old app."
        />
      </Card>

      <div class="grid gap-6 lg:grid-cols-2">
        <Card class="space-y-4">
          <SectionTitle
            title="Runtime endpoints"
            subtitle="Current frontend targets."
          />
          <div class="space-y-3 text-sm text-gray-600">
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Gateway API</p>
              <p class="mt-1 break-all">{GW_API_URL}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Tenant GraphQL API</p>
              <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
            </div>
          </div>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Local session state"
            subtitle="Storage-backed auth and tenant data."
          />
          <div class="flex flex-wrap gap-2">
            <Badge
              content={hasToken ? 'token present' : 'no token'}
              color={hasToken ? 'green' : 'red'}
            />
            <Badge
              content={tenantId || 'tenant not set'}
              color={tenantId ? 'indigo' : 'dark'}
            />
          </div>
          <Button
            color="alternative"
            onClick={() => {
              tenantStorage.clearTenantID();
              window.location.reload();
            }}
          >
            Clear tenant id
          </Button>
        </Card>
      </div>
    </PageShell>
  );
}
