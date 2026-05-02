import { For, Show, createSignal, onMount } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import { ensureActiveTenant } from '../../../services/auth';
import {
  checkPlatformPermission,
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
import {
  Badge,
  Button,
  Card,
  InputField,
} from '../../components/common/Primitives';
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

function membershipStatusColor(status: string) {
  return status === 'active' ? 'green' : 'dark';
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
  const [checkingPlatformAccess, setCheckingPlatformAccess] =
    createSignal(false);
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);
  const [canCreateTenant, setCanCreateTenant] = createSignal(false);
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active');

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

  const loadPlatformAccess = async () => {
    if (!userID) {
      setCanCreateTenant(false);
      return;
    }
    setCheckingPlatformAccess(true);
    try {
      const result = await checkPlatformPermission('tenant:create');
      if (!result.success) {
        setTenantError(result.message);
        setCanCreateTenant(false);
        return;
      }
      setCanCreateTenant(result.data);
    } finally {
      setCheckingPlatformAccess(false);
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
      setTenantError(data.message || 'Failed to open store');
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
      setTenantError('No signed-in account found.');
      return;
    }
    if (!canCreateTenant()) {
      setTenantError('Your account cannot create stores yet.');
      return;
    }

    const normalizedName = tenantName().trim();
    const normalizedSlug = slugify(tenantSlug() || normalizedName);
    if (!normalizedName || !normalizedSlug) {
      setTenantError('Store name and store slug are required.');
      return;
    }

    setCreatingTenant(true);
    setTenantError('');
    setTenantMessage('');
    try {
      const result = await createTenant({
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
          ? `Created store ${createdSlug} (${createdTenantID}).`
          : `Created store ${createdSlug}.`
      );
      await loadMemberships();
    } finally {
      setCreatingTenant(false);
    }
  };

  onMount(() => {
    void loadPlatformAccess();
    void loadMemberships();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Seller Backoffice"
          title="Manage your stores from one control room."
          copy="Create a new store, review where your team has access, and open the right workspace without relying on technical IDs."
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
          label="My stores"
          value={`${activeMemberships().length}/${memberships().length}`}
        />
        <StatCard
          label="Current store"
          value={tokenStorage.getActiveTenantID() || 'Not selected'}
        />
      </div>

      <Show when={checkingPlatformAccess()}>
        <LoadingInline label="Checking store creation access..." />
      </Show>

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="Create store"
            subtitle="Open a new store workspace and make yourself the store owner."
          />
          <Show when={!canCreateTenant() && !checkingPlatformAccess()}>
            <InfoAlert>
              Creating stores requires platform approval. The first platform
              owner can grant this access to additional operators.
            </InfoAlert>
          </Show>
          <form class="space-y-4" onSubmit={submitCreateTenant}>
            <InputField
              label="Store name"
              value={tenantName()}
              placeholder="Urban Finds"
              onInput={(event) => {
                const value = event.currentTarget.value;
                setTenantName(value);
                if (!tenantSlug().trim()) {
                  setTenantSlug(slugify(value));
                }
              }}
            />
            <InputField
              label="Store slug"
              value={tenantSlug()}
              placeholder="urban-finds"
              onInput={(event) =>
                setTenantSlug(slugify(event.currentTarget.value))
              }
            />
            <div class="flex flex-wrap gap-3">
              <Button
                type="submit"
                loading={creatingTenant()}
                disabled={!tenantName().trim() || !canCreateTenant()}
              >
                Create store
              </Button>
              <Badge
                content={
                  tenantSlug().trim()
                    ? `slug ${tenantSlug().trim()}`
                    : 'slug pending'
                }
                color={tenantSlug().trim() ? 'indigo' : 'dark'}
              />
            </div>
          </form>
        </Card>

        <Card class="space-y-4">
          <SectionTitle
            title="Quick store jump"
            subtitle="Keep direct store opening for debugging, but prefer opening from your assigned stores below."
          />
          <div class="flex flex-col gap-3 sm:flex-row">
            <input
              class="block w-full rounded-xl border border-gray-300 bg-white px-3 py-2.5 text-sm text-gray-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
              value={tenantId()}
              placeholder="store id"
              onInput={(event) => setTenantId(event.currentTarget.value)}
            />
            <Button
              disabled={!tenantId().trim() || switchingTenant()}
              loading={switchingTenant()}
              onClick={() => {
                void openTenant();
              }}
            >
              Open store
            </Button>
          </div>
          {!tenantId().trim() ? (
            <EmptyBlock
              title="No store selected"
              copy="Create a store or pick one from your assigned stores to open the right workspace."
            />
          ) : null}
        </Card>
      </div>

      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="My stores"
            subtitle="Stores your account can access right now."
          />
          <Show when={loadingTenants()}>
            <LoadingInline label="Loading stores..." />
          </Show>
          <Show
            when={!loadingTenants() && memberships().length > 0}
            fallback={
              <EmptyBlock
                title="No stores yet"
                copy="Create your first store to start managing catalog, team access, and future POD operations."
              />
            }
          >
            <div class="space-y-3">
              <For each={memberships()}>
                {(membership) => (
                  <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4">
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={membership.roleName} color="blue" />
                      <Badge
                        content={membership.status}
                        color={membershipStatusColor(membership.status)}
                      />
                      <Show
                        when={
                          membership.tenantId ===
                          tokenStorage.getActiveTenantID()
                        }
                      >
                        <Badge content="current" color="yellow" />
                      </Show>
                    </div>
                    <p class="mt-3 font-semibold text-gray-900">
                      {membership.tenantId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      access role for user {membership.userId}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-3">
                      <Button
                        size="sm"
                        onClick={() => {
                          void openTenant(membership.tenantId);
                        }}
                      >
                        Open store
                      </Button>
                      <Button
                        size="sm"
                        color="alternative"
                        onClick={() => {
                          setTenantId(membership.tenantId);
                          setTenantMessage(
                            `Prepared store ${membership.tenantId} as the next quick-jump target.`
                          );
                        }}
                      >
                        Use for quick jump
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
            title="Runtime endpoints"
            subtitle="Current application entrypoints used by the backoffice."
          />
          <div class="space-y-3 text-sm text-gray-600">
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Gateway</p>
              <p class="mt-1 break-all">{GW_API_URL}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Store GraphQL</p>
              <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4">
              <p class="font-semibold text-gray-900">Next step</p>
              <p class="mt-1">
                Use the settings page to manage team access, store invites,
                sessions, and platform administration.
              </p>
            </div>
          </div>
        </Card>
      </div>
    </PageShell>
  );
}
