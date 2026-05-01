import { For, Show, createSignal, onMount } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import { ensureActiveTenant } from '../../../services/auth';
import {
  createTenant,
  listUserTenants,
  type TenantMembership,
} from '../../../services/iam';
import { tenantStorage } from '../../../services/tenantStorage';
import { tokenStorage } from '../../../services/tokenStorage';
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingInline,
} from '../../components/common/Feedback';
import { PageShell } from '../../components/common/PageShell';
import { Badge, Button, Card, InputField } from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';
import { StatCard } from '../../components/dashboard/StatCard';

function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export default function AdminHomePage() {
  const user = tokenStorage.getUser();
  const userID = parseUserID(user?.id);

  const [tenantId, setTenantId] = createSignal(
    tokenStorage.getActiveTenantID() || tenantStorage.getTenantID()
  );
  const [tenantName, setTenantName] = createSignal('');
  const [tenantSlug, setTenantSlug] = createSignal('');
  const [tenantError, setTenantError] = createSignal('');
  const [tenantMessage, setTenantMessage] = createSignal('');
  const [switchingTenant, setSwitchingTenant] = createSignal(false);
  const [creatingTenant, setCreatingTenant] = createSignal(false);
  const [loadingTenants, setLoadingTenants] = createSignal(false);
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);

  const loadMemberships = async () => {
    if (!userID) return;
    setLoadingTenants(true);
    try {
      const result = await listUserTenants(userID);
      if (!result.success) {
        setTenantError(result.message);
        return;
      }
      setMemberships(result.data);
    } finally {
      setLoadingTenants(false);
    }
  };

  const openTenant = async (nextTenantID = tenantId().trim()) => {
    if (!nextTenantID) return;

    setSwitchingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const { success, data } = await ensureActiveTenant(nextTenantID);
      if (!success) {
        setTenantError(data.message || 'Failed to switch tenant');
        return;
      }

      tenantStorage.setTenantID(nextTenantID);
      setTenantId(nextTenantID);
      window.location.href = `/t/${nextTenantID}`;
    } finally {
      setSwitchingTenant(false);
    }
  };

  const submitCreateTenant = async (event: SubmitEvent) => {
    event.preventDefault();

    if (!userID) {
      setTenantError('No authenticated user found.');
      return;
    }

    const normalizedName = tenantName().trim();
    const normalizedSlug = slugify(tenantSlug() || normalizedName);
    if (!normalizedName || !normalizedSlug) {
      setTenantError('Tenant name and slug are required.');
      return;
    }

    setCreatingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const result = await createTenant({
        ownerUserId: userID,
        name: normalizedName,
        slug: normalizedSlug,
      });
      if (!result.success) {
        setTenantError(result.message);
        return;
      }

      const createdTenantID = result.data.tenant?.id || '';
      const createdSlug = result.data.tenant?.slug || normalizedSlug;
      setTenantName('');
      setTenantSlug('');
      setTenantId(createdTenantID);
      setTenantMessage(
        createdTenantID
          ? `Created tenant ${createdSlug} (${createdTenantID}).`
          : `Created tenant ${createdSlug}.`
      );
      await loadMemberships();
    } finally {
      setCreatingTenant(false);
    }
  };

  onMount(() => {
    void loadMemberships();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Control Room"
          title="Admin now manages tenants instead of guessing them."
          copy="Create a workspace, inspect your memberships, and jump into the active tenant session without typing raw tenant ids every time."
        />
      </Card>

      <Show when={tenantError()}>
        <ErrorAlert>{tenantError()}</ErrorAlert>
      </Show>

      <Show when={tenantMessage()}>
        <InfoAlert>{tenantMessage()}</InfoAlert>
      </Show>

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
          label="My tenants"
          value={String(memberships().length)}
        />
        <StatCard
          label="Active tenant"
          value={tokenStorage.getActiveTenantID() || 'Unset'}
        />
      </div>

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create tenant"
            subtitle="Provision a new workspace and attach yourself as tenant owner."
          />
          <form class="space-y-4" onSubmit={submitCreateTenant}>
            <InputField
              label="Tenant name"
              value={tenantName()}
              placeholder="Seller Alpha"
              onInput={(event) => {
                const value = event.currentTarget.value;
                setTenantName(value);
                if (!tenantSlug().trim()) {
                  setTenantSlug(slugify(value));
                }
              }}
            />
            <InputField
              label="Tenant slug"
              value={tenantSlug()}
              placeholder="seller-alpha"
              onInput={(event) => setTenantSlug(slugify(event.currentTarget.value))}
            />
            <div class="flex flex-wrap gap-3">
              <Button
                type="submit"
                loading={creatingTenant()}
                disabled={!tenantName().trim()}
              >
                Create tenant
              </Button>
              <Badge
                content={tenantSlug().trim() ? `slug ${tenantSlug().trim()}` : 'slug pending'}
                color={tenantSlug().trim() ? 'indigo' : 'dark'}
              />
            </div>
          </form>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Direct jump"
            subtitle="Keep the raw tenant jump for debugging or quick access."
          />
          <div class="flex flex-col gap-3 sm:flex-row">
            <input
              class="block w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
              value={tenantId()}
              placeholder="tenant id"
              onInput={(event) => setTenantId(event.currentTarget.value)}
            />
            <Button
              disabled={!tenantId().trim() || switchingTenant()}
              loading={switchingTenant()}
              onClick={() => {
                void openTenant();
              }}
            >
              Open tenant
            </Button>
          </div>
          {!tenantId().trim() ? (
            <EmptyBlock
              title="No tenant selected"
              copy="Create a tenant below or pick one from your memberships to switch the active tenant."
            />
          ) : null}
        </Card>
      </div>

      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="My tenants"
            subtitle="IAM memberships for the current authenticated user."
          />
          <Show when={loadingTenants()}>
            <LoadingInline label="Loading tenants..." />
          </Show>
          <Show
            when={!loadingTenants() && memberships().length > 0}
            fallback={
              <EmptyBlock
                title="No tenants yet"
                copy="Create your first tenant to start using the multi-tenant workspace flow."
              />
            }
          >
            <div class="space-y-3">
              <For each={memberships()}>
                {(membership) => (
                  <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={membership.roleName} color="blue" />
                      <Badge content={membership.status} color="green" />
                      <Show when={membership.tenantId === tokenStorage.getActiveTenantID()}>
                        <Badge content="active" color="yellow" />
                      </Show>
                    </div>
                    <p class="mt-3 font-semibold text-gray-900">
                      {membership.tenantId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      user {membership.userId}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-3">
                      <Button
                        size="sm"
                        onClick={() => {
                          void openTenant(membership.tenantId);
                        }}
                      >
                        Open workspace
                      </Button>
                      <Button
                        size="sm"
                        color="alternative"
                        onClick={() => setTenantId(membership.tenantId)}
                      >
                        Use as jump target
                      </Button>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Platform endpoints"
            subtitle="Current admin and tenant entrypoints."
          />
          <div class="space-y-3 text-sm text-gray-600">
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Gateway</p>
              <p class="mt-1 break-all">{GW_API_URL}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Tenant GraphQL</p>
              <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Next step</p>
              <p class="mt-1">
                Use the settings page to manage tenant members and role bindings.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </PageShell>
  );
}
