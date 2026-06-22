import { Show } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '@/services/baseurl';
import { tenantStorage } from '@/services/tenantStorage';
import { ErrorAlert, InfoAlert, LoadingInline } from '@/solid/components/common/Feedback';
import { Badge, Button, Card } from '@/solid/components/common/Primitives';
import { SectionLead } from '@/solid/components/common/SectionLead';
import { SectionTitle } from '@/solid/components/common/SectionTitle';
import { useAdminSettings } from './context';

export function HeaderRuntime() {
  const vm = useAdminSettings();

  return (
    <>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Settings"
          title="Manage workspace access, sessions, and platform controls."
          copy="This area brings together the operational controls behind the backoffice: current sessions, workspace access, workspace invites, and platform administration."
        />
        <div class="flex flex-wrap gap-3">
          <Button
            href="/admin/iam"
            color={vm.canManagePlatformRoles() ? 'dark' : 'light'}
            size="sm"
            disabled={!vm.canManagePlatformRoles()}
          >
            Open IAM console
          </Button>
        </div>
        <Show when={!vm.canManagePlatformRoles()}>
          <InfoAlert>Platform IAM is not available in this session.</InfoAlert>
        </Show>
      </Card>

      <Show when={vm.pageError()}>
        <ErrorAlert>{vm.pageError()}</ErrorAlert>
      </Show>
      <Show when={vm.memberActionMessage()}>
        <InfoAlert>{vm.memberActionMessage()}</InfoAlert>
      </Show>
      <Show when={vm.checkingPermissions()}>
        <LoadingInline label="Checking workspace access permissions..." />
      </Show>
      <Show when={vm.canManagePlatformRoles()}>
        <InfoAlert>
          Platform administration controls are enabled for this session.
        </InfoAlert>
      </Show>

      <div class="grid gap-6 lg:grid-cols-2">
        <RuntimeEndpoints />
        <LocalSessionState />
      </div>
    </>
  );
}

function RuntimeEndpoints() {
  return (
    <Card class="space-y-4">
      <SectionTitle title="Runtime endpoints" subtitle="Current frontend targets." />
      <div class="space-y-3 text-sm text-gray-600">
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-900">Gateway API</p>
          <p class="mt-1 break-all">{GW_API_URL}</p>
        </div>
        <div class="rounded-lg bg-gray-50 p-4">
          <p class="font-semibold text-gray-900">Store GraphQL API</p>
          <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
        </div>
      </div>
    </Card>
  );
}

function LocalSessionState() {
  const vm = useAdminSettings();

  return (
    <Card class="space-y-4">
      <SectionTitle
        title="Local session state"
        subtitle="Storage-backed sign-in and navigation state."
      />
      <div class="flex flex-wrap gap-2">
        <Badge
          content={vm.hasToken ? 'token present' : 'no token'}
          color={vm.hasToken ? 'green' : 'red'}
        />
        <Badge
          content={
            vm.activeTenantId()
              ? `current workspace ${vm.activeTenantId()}`
              : 'current workspace not set'
          }
          color={vm.activeTenantId() ? 'green' : 'dark'}
        />
        <Badge
          content={
            vm.sessionID() ? `session ${vm.sessionID()}` : 'session missing'
          }
          color={vm.sessionID() ? 'indigo' : 'red'}
        />
        <Badge
          content={
            vm.routeTenantId()
              ? `last opened workspace ${vm.routeTenantId()}`
              : 'last opened workspace not set'
          }
          color={vm.routeTenantId() ? 'indigo' : 'dark'}
        />
      </div>
      <Button
        color="alternative"
        onClick={() => {
          tenantStorage.clearTenantID();
          vm.setRouteTenantID('');
          window.location.reload();
        }}
      >
        Clear last opened workspace
      </Button>
    </Card>
  );
}
