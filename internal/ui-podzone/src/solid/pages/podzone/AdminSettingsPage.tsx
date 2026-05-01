import { For, Show, createEffect, createSignal, onMount } from 'solid-js';
import { GW_API_URL, TENANT_GQL_URL } from '../../../services/baseurl';
import {
  listAuditLogs,
  listSessions,
  revokeSession,
  type AuditLogInfo,
  type SessionInfo,
} from '../../../services/auth';
import {
  checkPermission,
  checkPlatformPermission,
  listTenantMembers,
  listPlatformRoles,
  listUserTenants,
  removePlatformRole,
  removeTenantMember,
  type PlatformRoleMembership,
  upsertPlatformRole,
  upsertTenantMember,
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
  SelectField,
} from '../../components/common/Primitives';
import { SectionLead } from '../../components/common/SectionLead';
import { SectionTitle } from '../../components/common/SectionTitle';

const roleOptions = [
  { name: 'Tenant owner', value: 'tenant_owner' },
  { name: 'Tenant admin', value: 'tenant_admin' },
  { name: 'Tenant editor', value: 'tenant_editor' },
  { name: 'Tenant viewer', value: 'tenant_viewer' },
];

const platformRoleOptions = [
  { name: 'Platform owner', value: 'platform_owner' },
  { name: 'Platform admin', value: 'platform_admin' },
];

function parseUserID(raw: unknown): number {
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw;
  if (typeof raw === 'string') {
    const parsed = Number.parseInt(raw, 10);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export default function AdminSettingsPage() {
  const hasToken = Boolean(tokenStorage.getToken());
  const userID = parseUserID(tokenStorage.getUser()?.id);
  const activeTenantId = () => tokenStorage.getActiveTenantID();
  const sessionID = () => tokenStorage.getSessionID();

  const [routeTenantId, setRouteTenantID] = createSignal(
    tenantStorage.getTenantID()
  );
  const [memberTenantId, setMemberTenantId] = createSignal(
    activeTenantId() || tenantStorage.getTenantID()
  );
  const [memberUserId, setMemberUserId] = createSignal('');
  const [roleName, setRoleName] = createSignal(roleOptions[1].value);
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);
  const [members, setMembers] = createSignal<TenantMembership[]>([]);
  const [platformRoles, setPlatformRoles] = createSignal<PlatformRoleMembership[]>([]);
  const [sessions, setSessions] = createSignal<SessionInfo[]>([]);
  const [auditLogs, setAuditLogs] = createSignal<AuditLogInfo[]>([]);
  const [loadingTenants, setLoadingTenants] = createSignal(false);
  const [loadingMembers, setLoadingMembers] = createSignal(false);
  const [loadingSessions, setLoadingSessions] = createSignal(false);
  const [loadingAuditLogs, setLoadingAuditLogs] = createSignal(false);
  const [savingMember, setSavingMember] = createSignal(false);
  const [checkingPermissions, setCheckingPermissions] = createSignal(false);
  const [loadingPlatformRoles, setLoadingPlatformRoles] = createSignal(false);
  const [savingPlatformRole, setSavingPlatformRole] = createSignal(false);
  const [pageError, setPageError] = createSignal('');
  const [memberActionMessage, setMemberActionMessage] = createSignal('');
  const [canReadTenant, setCanReadTenant] = createSignal(false);
  const [canManageMembers, setCanManageMembers] = createSignal(false);
  const [canManagePlatformRoles, setCanManagePlatformRoles] = createSignal(false);
  const [platformUserId, setPlatformUserId] = createSignal(userID ? String(userID) : '');
  const [platformRoleName, setPlatformRoleName] = createSignal(platformRoleOptions[1].value);

  const loadSessions = async () => {
    setLoadingSessions(true);
    try {
      const result = await listSessions();
      if (!result.success) {
        setPageError(result.data.message);
        return;
      }
      setSessions(result.data);
    } finally {
      setLoadingSessions(false);
    }
  };

  const loadAuditLogs = async () => {
    setLoadingAuditLogs(true);
    try {
      const result = await listAuditLogs(25);
      if (!result.success) {
        setPageError(result.data.message);
        return;
      }
      setAuditLogs(result.data);
    } finally {
      setLoadingAuditLogs(false);
    }
  };

  const loadUserTenants = async () => {
    if (!userID) return;
    setLoadingTenants(true);
    try {
      const result = await listUserTenants(userID);
      if (!result.success) {
        setPageError(result.message);
        return;
      }
      setMemberships(result.data);
    } finally {
      setLoadingTenants(false);
    }
  };

  const loadTenantMembers = async (tenantId = memberTenantId().trim()) => {
    if (!tenantId) {
      setMembers([]);
      return;
    }
    if (!canReadTenant()) {
      setMembers([]);
      return;
    }
    setLoadingMembers(true);
    setPageError('');
    try {
      const result = await listTenantMembers(tenantId);
      if (!result.success) {
        setPageError(result.message);
        setMembers([]);
        return;
      }
      setMembers(result.data);
    } finally {
      setLoadingMembers(false);
    }
  };

  const submitMember = async (event: SubmitEvent) => {
    event.preventDefault();
    const tenantId = memberTenantId().trim();
    const parsedUserId = Number.parseInt(memberUserId().trim(), 10);
    if (!tenantId || !Number.isFinite(parsedUserId) || parsedUserId <= 0) {
      setPageError('Tenant id and user id are required.');
      return;
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage tenant members.');
      return;
    }

    setSavingMember(true);
    setPageError('');
    setMemberActionMessage('');
    try {
      const result = await upsertTenantMember({
        tenantId,
        userId: parsedUserId,
        roleName: roleName(),
      });
      if (!result.success) {
        setPageError(result.message);
        return;
      }
      setMemberUserId('');
      setMemberActionMessage(`Saved member ${parsedUserId} in ${tenantId}.`);
      await loadTenantMembers(tenantId);
    } finally {
      setSavingMember(false);
    }
  };

  const handleRemoveMember = async (tenantId: string, userId: number) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to remove tenant members.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await removeTenantMember(tenantId, userId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Removed member ${userId} from ${tenantId}.`);
    await loadTenantMembers(tenantId);
  };

  const refreshTenantPermissions = async (tenantId = memberTenantId().trim()) => {
    if (!userID || !tenantId) {
      setCanReadTenant(false);
      setCanManageMembers(false);
      return;
    }
    setCheckingPermissions(true);
    try {
      const [readResult, manageResult] = await Promise.all([
        checkPermission({
          tenantId,
          userId: userID,
          permission: 'tenant:read',
        }),
        checkPermission({
          tenantId,
          userId: userID,
          permission: 'tenant:manage_members',
        }),
      ]);
      if (!readResult.success) {
        setPageError(readResult.message);
        setCanReadTenant(false);
      } else {
        setCanReadTenant(readResult.data);
      }
      if (!manageResult.success) {
        setPageError(manageResult.message);
        setCanManageMembers(false);
      } else {
        setCanManageMembers(manageResult.data);
      }
    } finally {
      setCheckingPermissions(false);
    }
  };

  const refreshPlatformPermissions = async () => {
    if (!userID) {
      setCanManagePlatformRoles(false);
      return;
    }
    const result = await checkPlatformPermission('platform:manage_roles');
    if (!result.success) {
      setPageError(result.message);
      setCanManagePlatformRoles(false);
      return;
    }
    setCanManagePlatformRoles(result.data);
  };

  const loadPlatformRoleAssignments = async () => {
    const targetUserID = Number.parseInt(platformUserId().trim(), 10);
    if (!userID || !canManagePlatformRoles() || !Number.isFinite(targetUserID) || targetUserID <= 0) {
      setPlatformRoles([]);
      return;
    }
    setLoadingPlatformRoles(true);
    try {
      const result = await listPlatformRoles(targetUserID);
      if (!result.success) {
        setPageError(result.message);
        setPlatformRoles([]);
        return;
      }
      setPlatformRoles(result.data);
    } finally {
      setLoadingPlatformRoles(false);
    }
  };

  const submitPlatformRole = async (event: SubmitEvent) => {
    event.preventDefault();
    const targetUserID = Number.parseInt(platformUserId().trim(), 10);
    if (!canManagePlatformRoles()) {
      setPageError('You do not have permission to manage platform roles.');
      return;
    }
    if (!Number.isFinite(targetUserID) || targetUserID <= 0) {
      setPageError('Target user id is required for platform role management.');
      return;
    }
    setSavingPlatformRole(true);
    setPageError('');
    setMemberActionMessage('');
    try {
      const result = await upsertPlatformRole({
        targetUserId: targetUserID,
        roleName: platformRoleName(),
      });
      if (!result.success) {
        setPageError(result.message);
        return;
      }
      setMemberActionMessage(`Saved platform role ${platformRoleName()} for user ${targetUserID}.`);
      await loadPlatformRoleAssignments();
    } finally {
      setSavingPlatformRole(false);
    }
  };

  const handleRemovePlatformRole = async (targetUserID: number, roleName: string) => {
    if (!canManagePlatformRoles()) {
      setPageError('You do not have permission to manage platform roles.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await removePlatformRole(targetUserID, roleName);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Removed platform role ${roleName} from user ${targetUserID}.`);
    await loadPlatformRoleAssignments();
  };

  const handleRevokeSession = async (sessionId: string) => {
    setPageError('');
    setMemberActionMessage('');
    const result = await revokeSession(sessionId);
    if (!result.success) {
      setPageError(result.data.message || 'Failed to revoke session');
      return;
    }
    setMemberActionMessage(`Revoked session ${sessionId}.`);
    await loadSessions();
  };

  onMount(() => {
    void loadUserTenants();
    void refreshPlatformPermissions();
    void loadSessions();
    void loadAuditLogs();
  });

  createEffect(() => {
    const tenantId = memberTenantId().trim();
    if (!tenantId || !userID) {
      setCanReadTenant(false);
      setCanManageMembers(false);
      setMembers([]);
      return;
    }

    void (async () => {
      await refreshTenantPermissions(tenantId);
      if (canReadTenant()) {
        await loadTenantMembers(tenantId);
      } else {
        setMembers([]);
      }
    })();
  });

  createEffect(() => {
    if (!canManagePlatformRoles()) {
      setPlatformRoles([]);
      return;
    }
    void loadPlatformRoleAssignments();
  });

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Settings"
          title="Inspect session state and manage tenant access."
          copy="The frontend now exposes the IAM MVP surface directly: active session metadata, user memberships, and tenant member upserts and removals."
        />
      </Card>

      <Show when={pageError()}>
        <ErrorAlert>{pageError()}</ErrorAlert>
      </Show>

      <Show when={memberActionMessage()}>
        <InfoAlert>{memberActionMessage()}</InfoAlert>
      </Show>

      <Show when={checkingPermissions()}>
        <LoadingInline label="Checking tenant permissions..." />
      </Show>

      <Show when={canManagePlatformRoles()}>
        <InfoAlert>Platform role management is enabled for this session.</InfoAlert>
      </Show>

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
            subtitle="Storage-backed auth and navigation state."
          />
          <div class="flex flex-wrap gap-2">
            <Badge
              content={hasToken ? 'token present' : 'no token'}
              color={hasToken ? 'green' : 'red'}
            />
            <Badge
              content={
                activeTenantId()
                  ? `token tenant ${activeTenantId()}`
                  : 'token tenant not set'
              }
              color={activeTenantId() ? 'green' : 'dark'}
            />
            <Badge
              content={sessionID() ? `session ${sessionID()}` : 'session missing'}
              color={sessionID() ? 'indigo' : 'red'}
            />
            <Badge
              content={
                routeTenantId()
                  ? `last route tenant ${routeTenantId()}`
                  : 'last route tenant not set'
              }
              color={routeTenantId() ? 'indigo' : 'dark'}
            />
          </div>
          <Button
            color="alternative"
            onClick={() => {
              tenantStorage.clearTenantID();
              setRouteTenantID('');
              window.location.reload();
            }}
          >
            Clear route tenant
          </Button>
        </Card>
      </div>

      <Card class="space-y-4">
        <SectionTitle
          title="Sessions"
          subtitle="Inspect and revoke active or stale sessions for the current user."
        />
        <div class="flex flex-wrap gap-3">
          <Button
            color="alternative"
            onClick={() => {
              void loadSessions();
            }}
          >
            Reload sessions
          </Button>
        </div>
        <Show when={loadingSessions()}>
          <LoadingInline label="Loading sessions..." />
        </Show>
        <Show
          when={!loadingSessions() && sessions().length > 0}
          fallback={
            <EmptyBlock
              title="No sessions loaded"
              copy="Authenticated sessions will appear here once the user signs in."
            />
          }
        >
          <div class="space-y-3">
            <For each={sessions()}>
              {(session) => (
                <div class="rounded-2xl border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">{session.id}</p>
                      <p class="mt-1 text-sm text-gray-500">
                        tenant {session.activeTenantId || 'unset'} · expires {session.expiresAt || 'unknown'}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={session.status || 'unknown'} color={session.status === 'active' ? 'green' : 'dark'} />
                      <Show when={session.id === sessionID()}>
                        <Badge content="current" color="yellow" />
                      </Show>
                      <Button
                        color="red"
                        size="xs"
                        onClick={() => {
                          void handleRevokeSession(session.id);
                        }}
                      >
                        Revoke
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Audit trail"
          subtitle="Recent IAM and session actions performed by the current authenticated user."
        />
        <div class="flex flex-wrap gap-3">
          <Button
            color="alternative"
            onClick={() => {
              void loadAuditLogs();
            }}
          >
            Reload audit logs
          </Button>
        </div>
        <Show when={loadingAuditLogs()}>
          <LoadingInline label="Loading audit logs..." />
        </Show>
        <Show
          when={!loadingAuditLogs() && auditLogs().length > 0}
          fallback={
            <EmptyBlock
              title="No audit logs yet"
              copy="Sensitive auth and IAM actions will appear here after they run."
            />
          }
        >
          <div class="space-y-3">
            <For each={auditLogs()}>
              {(log) => (
                <div class="rounded-2xl border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">{log.action || 'unknown action'}</p>
                      <p class="mt-1 text-sm text-gray-500">
                        {log.resourceType || 'resource'} {log.resourceId || 'n/a'} · tenant {log.tenantId || 'global'}
                      </p>
                      <Show when={log.payloadJson}>
                        <pre class="mt-3 overflow-x-auto rounded-2xl bg-gray-50 p-3 text-xs text-gray-700">{log.payloadJson}</pre>
                      </Show>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={log.status || 'unknown'} color={log.status === 'success' ? 'green' : 'dark'} />
                      <Badge content={log.createdAt || 'time unknown'} color="indigo" />
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Card>

      <div class="grid gap-6 lg:grid-cols-[0.95fr_1.05fr]">
        <Card class="space-y-4">
          <SectionTitle
            title="My tenant memberships"
            subtitle="Memberships returned by IAM for the current user."
          />
          <Show when={loadingTenants()}>
            <LoadingInline label="Loading memberships..." />
          </Show>
          <Show
            when={!loadingTenants() && memberships().length > 0}
            fallback={
              <EmptyBlock
                title="No memberships yet"
                copy="Create or join a tenant to see workspace memberships here."
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
                    </div>
                    <p class="mt-3 font-semibold text-gray-900">
                      {membership.tenantId}
                    </p>
                    <p class="mt-1 text-sm text-gray-500">
                      user {membership.userId}
                    </p>
                    <div class="mt-3">
                      <Button
                        size="sm"
                        color="alternative"
                        onClick={() => {
                          setMemberTenantId(membership.tenantId);
                          void loadTenantMembers(membership.tenantId);
                        }}
                      >
                        Inspect members
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
            title="Tenant member management"
            subtitle="List, add, update, or remove members for a tenant."
          />

          <form class="space-y-4" onSubmit={submitMember}>
            <InputField
              label="Tenant id"
              value={memberTenantId()}
              placeholder="tenant id"
              onInput={(event) => setMemberTenantId(event.currentTarget.value)}
            />
            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="User id"
                value={memberUserId()}
                placeholder="42"
                onInput={(event) => setMemberUserId(event.currentTarget.value)}
              />
              <SelectField
                label="Role"
                value={roleName()}
                options={roleOptions}
                onChange={(event) => setRoleName(event.currentTarget.value)}
              />
            </div>
            <div class="flex flex-wrap gap-3">
              <Button
                type="submit"
                loading={savingMember()}
                disabled={!canManageMembers()}
              >
                Save member
              </Button>
              <Button
                color="alternative"
                disabled={!canReadTenant()}
                onClick={() => {
                  void loadTenantMembers();
                }}
              >
                Reload members
              </Button>
            </div>
          </form>

          <Show when={loadingMembers()}>
            <LoadingInline label="Loading tenant members..." />
          </Show>

          <Show
            when={!loadingMembers() && members().length > 0}
            fallback={
              <EmptyBlock
                title="No tenant members loaded"
                copy="Pick a tenant and reload members to inspect the current IAM bindings."
              />
            }
          >
            <Show when={!canReadTenant()}>
              <EmptyBlock
                title="No tenant read access"
                copy="You do not currently have permission to inspect members for this tenant."
              />
            </Show>
            <Show when={canReadTenant() && !canManageMembers()}>
              <InfoAlert>
                You can inspect this tenant, but only tenant owners can manage members
                in the current IAM policy set.
              </InfoAlert>
            </Show>
            <div class="space-y-3">
              <For each={members()}>
                {(membership) => (
                  <div class="rounded-2xl border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">
                          user {membership.userId}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          {membership.tenantId}
                        </p>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge content={membership.roleName} color="blue" />
                        <Badge content={membership.status} color="green" />
                        <Button
                          color="red"
                          size="xs"
                          disabled={!canManageMembers()}
                          onClick={() => {
                            void handleRemoveMember(
                              membership.tenantId,
                              membership.userId
                            );
                          }}
                        >
                          Remove
                        </Button>
                      </div>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>
        </Card>
      </div>

      <Card class="space-y-4">
        <SectionTitle
          title="Platform roles"
          subtitle="Assign or revoke platform-wide roles such as tenant creation and platform role management."
        />
        <Show when={!canManagePlatformRoles()}>
          <InfoAlert>
            Platform role management requires the permission <code>platform:manage_roles</code>.
          </InfoAlert>
        </Show>
        <form class="space-y-4" onSubmit={submitPlatformRole}>
          <div class="grid gap-4 md:grid-cols-2">
            <InputField
              label="Target user id"
              value={platformUserId()}
              placeholder="1"
              onInput={(event) => setPlatformUserId(event.currentTarget.value)}
            />
            <SelectField
              label="Platform role"
              value={platformRoleName()}
              options={platformRoleOptions}
              onChange={(event) => setPlatformRoleName(event.currentTarget.value)}
            />
          </div>
          <div class="flex flex-wrap gap-3">
            <Button
              type="submit"
              loading={savingPlatformRole()}
              disabled={!canManagePlatformRoles()}
            >
              Save platform role
            </Button>
            <Button
              color="alternative"
              disabled={!canManagePlatformRoles()}
              onClick={() => {
                void loadPlatformRoleAssignments();
              }}
            >
              Reload platform roles
            </Button>
          </div>
        </form>

        <Show when={loadingPlatformRoles()}>
          <LoadingInline label="Loading platform roles..." />
        </Show>

        <Show
          when={!loadingPlatformRoles() && platformRoles().length > 0}
          fallback={
            <EmptyBlock
              title="No platform roles loaded"
              copy="Choose a target user to inspect platform-level role assignments."
            />
          }
        >
          <div class="space-y-3">
            <For each={platformRoles()}>
              {(membership) => (
                <div class="rounded-2xl border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">
                        user {membership.userId}
                      </p>
                      <p class="mt-1 text-sm text-gray-500">
                        {membership.roleName}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={membership.status} color="green" />
                      <Button
                        color="red"
                        size="xs"
                        disabled={!canManagePlatformRoles()}
                        onClick={() => {
                          void handleRemovePlatformRole(
                            membership.userId,
                            membership.roleName
                          );
                        }}
                      >
                        Remove
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Card>
    </PageShell>
  );
}
