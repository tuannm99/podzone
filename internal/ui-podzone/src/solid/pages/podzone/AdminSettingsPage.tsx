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
  createTenantInvite,
  listTenantMembers,
  listTenantInvites,
  listPlatformRoles,
  listUserTenants,
  removePlatformRole,
  revokeTenantInvite,
  removeTenantMember,
  type PlatformRoleMembership,
  type TenantInvite,
  upsertPlatformRole,
  upsertTenantMember,
  upsertTenantMemberByIdentity,
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
  { name: 'Store owner', value: 'tenant_owner' },
  { name: 'Store admin', value: 'tenant_admin' },
  { name: 'Store operator', value: 'tenant_editor' },
  { name: 'Store viewer', value: 'tenant_viewer' },
];

const platformRoleOptions = [
  { name: 'Platform owner', value: 'platform_owner' },
  { name: 'Platform admin', value: 'platform_admin' },
];

function sessionStatusColor(status?: string) {
  return status === 'active' ? 'green' : 'dark';
}

function membershipStatusColor(status?: string) {
  return status === 'active' ? 'green' : 'dark';
}

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
  const [memberIdentity, setMemberIdentity] = createSignal('');
  const [roleName, setRoleName] = createSignal(roleOptions[1].value);
  const [inviteEmail, setInviteEmail] = createSignal('');
  const [inviteRoleName, setInviteRoleName] = createSignal(roleOptions[1].value);
  const [memberships, setMemberships] = createSignal<TenantMembership[]>([]);
  const [members, setMembers] = createSignal<TenantMembership[]>([]);
  const [invites, setInvites] = createSignal<TenantInvite[]>([]);
  const [platformRoles, setPlatformRoles] = createSignal<PlatformRoleMembership[]>([]);
  const [sessions, setSessions] = createSignal<SessionInfo[]>([]);
  const [auditLogs, setAuditLogs] = createSignal<AuditLogInfo[]>([]);
  const [loadingTenants, setLoadingTenants] = createSignal(false);
  const [loadingMembers, setLoadingMembers] = createSignal(false);
  const [loadingInvites, setLoadingInvites] = createSignal(false);
  const [loadingSessions, setLoadingSessions] = createSignal(false);
  const [loadingAuditLogs, setLoadingAuditLogs] = createSignal(false);
  const [savingMember, setSavingMember] = createSignal(false);
  const [savingInvite, setSavingInvite] = createSignal(false);
  const [checkingPermissions, setCheckingPermissions] = createSignal(false);
  const [loadingPlatformRoles, setLoadingPlatformRoles] = createSignal(false);
  const [savingPlatformRole, setSavingPlatformRole] = createSignal(false);
  const [pageError, setPageError] = createSignal('');
  const [memberActionMessage, setMemberActionMessage] = createSignal('');
  const [latestInviteAcceptURL, setLatestInviteAcceptURL] = createSignal('');
  const [canReadTenant, setCanReadTenant] = createSignal(false);
  const [canManageMembers, setCanManageMembers] = createSignal(false);
  const [canManagePlatformRoles, setCanManagePlatformRoles] = createSignal(false);
  const [platformUserId, setPlatformUserId] = createSignal(userID ? String(userID) : '');
  const [platformRoleName, setPlatformRoleName] = createSignal(platformRoleOptions[1].value);
  const tenantOptions = () =>
    memberships().map((membership) => ({
      name: `${membership.tenantId} · ${membership.roleName}`,
      value: membership.tenantId,
    }));
  const currentSessionCount = () => sessions().filter((session) => session.id === sessionID()).length;
  const otherSessionCount = () => sessions().filter((session) => session.id !== sessionID()).length;

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

  const loadTenantInvites = async (tenantId = memberTenantId().trim()) => {
    if (!tenantId || !canManageMembers()) {
      setInvites([]);
      return;
    }
    setLoadingInvites(true);
    setPageError('');
    try {
      const result = await listTenantInvites(tenantId);
      if (!result.success) {
        setPageError(result.message);
        setInvites([]);
        return;
      }
      setInvites(result.data);
    } finally {
      setLoadingInvites(false);
    }
  };

  const submitMember = async (event: SubmitEvent) => {
    event.preventDefault();
    const tenantId = memberTenantId().trim();
    const parsedUserId = Number.parseInt(memberUserId().trim(), 10);
    const identity = memberIdentity().trim();
    if (!tenantId || (identity === '' && (!Number.isFinite(parsedUserId) || parsedUserId <= 0))) {
      setPageError('Store id and either teammate identity or user id are required.');
      return;
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage team access for this store.');
      return;
    }

    setSavingMember(true);
    setPageError('');
    setMemberActionMessage('');
    try {
      if (identity) {
        const result = await upsertTenantMemberByIdentity({
          tenantId,
          identity,
          roleName: roleName(),
        });
        if (!result.success) {
          setPageError(result.message);
          return;
        }
        setMemberUserId(result.data.userId ? String(result.data.userId) : '');
        setMemberIdentity('');
        setMemberActionMessage(
          result.data.createdUser
            ? `Created a new account and granted store access for ${identity} in ${tenantId}.`
            : `Granted store access to existing user ${result.data.userId} for ${identity} in ${tenantId}.`
        );
      } else {
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
        setMemberActionMessage(`Saved store access for user ${parsedUserId} in ${tenantId}.`);
      }
      await loadTenantMembers(tenantId);
    } finally {
      setSavingMember(false);
    }
  };

  const handleRemoveMember = async (tenantId: string, userId: number) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to remove store access.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await removeTenantMember(tenantId, userId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Removed user ${userId} from store ${tenantId}.`);
    await loadTenantMembers(tenantId);
  };

  const submitInvite = async (event: SubmitEvent) => {
    event.preventDefault();
    const tenantId = memberTenantId().trim();
    const email = inviteEmail().trim();
    if (!tenantId || !email) {
      setPageError('Store id and invite email are required.');
      return;
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage store invites.');
      return;
    }

    setSavingInvite(true);
    setPageError('');
    setMemberActionMessage('');
    setLatestInviteAcceptURL('');
    try {
      const result = await createTenantInvite({
        tenantId,
        email,
        roleName: inviteRoleName(),
      });
      if (!result.success) {
        setPageError(result.message);
        return;
      }
      setInviteEmail('');
      setLatestInviteAcceptURL(result.data.acceptUrl);
      setMemberActionMessage(`Created a store invite for ${email} in ${tenantId}.`);
      await loadTenantInvites(tenantId);
    } finally {
      setSavingInvite(false);
    }
  };

  const handleRevokeInvite = async (inviteId: string, tenantId: string, email: string) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to revoke store invites.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await revokeTenantInvite(inviteId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Revoked store invite for ${email}.`);
    await loadTenantInvites(tenantId);
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
      setPageError('You do not have permission to manage platform administration roles.');
      return;
    }
    if (!Number.isFinite(targetUserID) || targetUserID <= 0) {
      setPageError('Target user id is required for platform administration roles.');
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
      setMemberActionMessage(`Saved platform admin role ${platformRoleName()} for user ${targetUserID}.`);
      await loadPlatformRoleAssignments();
    } finally {
      setSavingPlatformRole(false);
    }
  };

  const handleRemovePlatformRole = async (targetUserID: number, roleName: string) => {
    if (!canManagePlatformRoles()) {
      setPageError('You do not have permission to manage platform administration roles.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await removePlatformRole(targetUserID, roleName);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Removed platform admin role ${roleName} from user ${targetUserID}.`);
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
    if (!memberTenantId().trim()) {
      const nextTenantId = activeTenantId() || memberships()[0]?.tenantId || '';
      if (nextTenantId) {
        setMemberTenantId(nextTenantId);
      }
    }
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
      if (canManageMembers()) {
        await loadTenantInvites(tenantId);
      } else {
        setInvites([]);
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
          title="Manage store access, sessions, and platform controls."
          copy="This area brings together the operational controls behind the backoffice: current sessions, team access, store invites, and platform administration."
        />
      </Card>

      <Show when={pageError()}>
        <ErrorAlert>{pageError()}</ErrorAlert>
      </Show>

      <Show when={memberActionMessage()}>
        <InfoAlert>{memberActionMessage()}</InfoAlert>
      </Show>

      <Show when={checkingPermissions()}>
        <LoadingInline label="Checking store access permissions..." />
      </Show>

      <Show when={canManagePlatformRoles()}>
        <InfoAlert>Platform administration controls are enabled for this session.</InfoAlert>
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
              <p class="font-semibold text-gray-900">Store GraphQL API</p>
              <p class="mt-1 break-all">{TENANT_GQL_URL}</p>
            </div>
          </div>
        </Card>

      <Card class="space-y-4">
        <SectionTitle
          title="Local session state"
          subtitle="Storage-backed sign-in and navigation state."
          />
          <div class="flex flex-wrap gap-2">
            <Badge
              content={hasToken ? 'token present' : 'no token'}
              color={hasToken ? 'green' : 'red'}
            />
            <Badge
              content={
                activeTenantId()
                  ? `current store ${activeTenantId()}`
                  : 'current store not set'
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
                  ? `last opened store ${routeTenantId()}`
                  : 'last opened store not set'
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
            Clear last opened store
          </Button>
        </Card>
      </div>

      <Card class="space-y-4">
        <SectionTitle
          title="Sessions"
          subtitle="Review active sign-ins and revoke sessions that should no longer access your stores."
        />
        <div class="flex flex-wrap gap-3">
          <Badge content={`current ${currentSessionCount()}`} color="yellow" />
          <Badge content={`other ${otherSessionCount()}`} color="indigo" />
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
              copy="Signed-in sessions will appear here once this account starts using the backoffice."
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
                        store {session.activeTenantId || 'not selected'} · expires {session.expiresAt || 'unknown'}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={session.status || 'unknown'} color={sessionStatusColor(session.status)} />
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
          subtitle="Recent access, invite, session, and platform administration actions performed by this account."
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
              copy="Sensitive sign-in, access, and administration actions will appear here after they run."
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
                        {log.resourceType || 'resource'} {log.resourceId || 'n/a'} · store {log.tenantId || 'global'}
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
            title="My store access"
            subtitle="Stores this account can access right now."
          />
          <Show when={loadingTenants()}>
            <LoadingInline label="Loading store access..." />
          </Show>
          <Show
            when={!loadingTenants() && memberships().length > 0}
            fallback={
              <EmptyBlock
                title="No store access yet"
                copy="Create or join a store to see your working spaces here."
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
                      teammate {membership.userId}
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
                        Open team access
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
            title="Team access"
            subtitle="List, add, update, or remove store teammates. Start from one of your stores instead of typing technical IDs by hand."
          />

          <form class="space-y-4" onSubmit={submitMember}>
            <Show when={tenantOptions().length > 0}>
              <SelectField
                label="Store"
                value={memberTenantId()}
                options={tenantOptions()}
                onChange={(event) => setMemberTenantId(event.currentTarget.value)}
              />
            </Show>
            <InputField
              label="Store id override"
              value={memberTenantId()}
              placeholder="store id"
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
            <InputField
              label="Teammate email or username"
              value={memberIdentity()}
              placeholder="ops@store.com or store_operator"
              onInput={(event) => setMemberIdentity(event.currentTarget.value)}
            />
            <div class="flex flex-wrap gap-3">
              <Button
                type="button"
                color="light"
                disabled={!userID}
                onClick={() => setMemberUserId(String(userID))}
              >
                Use my user id
              </Button>
              <Button
                type="button"
                color="light"
                disabled={!tokenStorage.getUser()?.email}
                onClick={() => setMemberIdentity(tokenStorage.getUser()?.email || '')}
              >
                Use my email
              </Button>
              <Button
                type="submit"
                loading={savingMember()}
                disabled={!canManageMembers()}
              >
                Save access
              </Button>
              <Button
                type="button"
                color="alternative"
                disabled={!canReadTenant()}
                onClick={() => {
                  void loadTenantMembers();
                }}
              >
                Reload team
              </Button>
            </div>
          </form>

          <Show when={loadingMembers()}>
            <LoadingInline label="Loading store team..." />
          </Show>

          <Show
            when={!loadingMembers() && members().length > 0}
            fallback={
              <EmptyBlock
                title="No team members loaded"
                copy="Choose a store and reload team access to inspect who can operate in that store."
              />
            }
          >
            <Show when={!canReadTenant()}>
              <EmptyBlock
                title="No store access"
                copy="You do not currently have permission to inspect team access for this store."
              />
            </Show>
            <Show when={canReadTenant() && !canManageMembers()}>
              <InfoAlert>
                You can inspect this store, but only authorized store owners or admins can manage team access.
              </InfoAlert>
            </Show>
            <div class="space-y-3">
              <For each={members()}>
                {(membership) => (
                  <div class="rounded-2xl border border-gray-200 p-4">
                    <div class="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <p class="font-semibold text-gray-900">
                          teammate {membership.userId}
                        </p>
                        <p class="mt-1 text-sm text-gray-500">
                          {membership.tenantId}
                        </p>
                      </div>
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge content={membership.roleName} color="blue" />
                        <Badge content={membership.status} color={membershipStatusColor(membership.status)} />
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
          title="Store invites"
          subtitle="Create email invites, track pending team access, and revoke old invite links."
        />
        <Show when={!canManageMembers()}>
          <InfoAlert>
            Store invites require access to manage team permissions for this store.
          </InfoAlert>
        </Show>
        <form class="space-y-4" onSubmit={submitInvite}>
          <Show when={tenantOptions().length > 0}>
            <SelectField
              label="Store"
              value={memberTenantId()}
              options={tenantOptions()}
              onChange={(event) => setMemberTenantId(event.currentTarget.value)}
            />
          </Show>
          <div class="grid gap-4 md:grid-cols-2">
            <InputField
              label="Invite email"
              value={inviteEmail()}
              placeholder="owner@shop.com"
              onInput={(event) => setInviteEmail(event.currentTarget.value)}
            />
            <SelectField
              label="Role"
              value={inviteRoleName()}
              options={roleOptions}
              onChange={(event) => setInviteRoleName(event.currentTarget.value)}
            />
          </div>
          <div class="flex flex-wrap gap-3">
            <Button
              type="button"
              color="light"
              disabled={!tokenStorage.getUser()?.email}
              onClick={() => setInviteEmail(tokenStorage.getUser()?.email || '')}
            >
              Use my email
            </Button>
            <Button
              type="submit"
              loading={savingInvite()}
              disabled={!canManageMembers()}
            >
              Create store invite
            </Button>
            <Button
              type="button"
              color="alternative"
              disabled={!canManageMembers()}
              onClick={() => {
                void loadTenantInvites();
              }}
            >
              Reload invites
            </Button>
          </div>
        </form>

        <Show when={latestInviteAcceptURL()}>
          <div class="rounded-2xl bg-gray-50 p-4 text-sm text-gray-700">
            <p class="font-semibold text-gray-900">Latest join link</p>
            <p class="mt-2 break-all">{latestInviteAcceptURL()}</p>
          </div>
        </Show>

        <Show when={loadingInvites()}>
          <LoadingInline label="Loading store invites..." />
        </Show>

        <Show
          when={!loadingInvites() && invites().length > 0}
          fallback={
            <EmptyBlock
              title="No store invites loaded"
              copy="Create an invite or reload invites for the selected store."
            />
          }
        >
          <div class="space-y-3">
            <For each={invites()}>
              {(invite) => (
                <div class="rounded-2xl border border-gray-200 p-4">
                  <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p class="font-semibold text-gray-900">{invite.email}</p>
                      <p class="mt-1 text-sm text-gray-500">
                        store {invite.tenantId} · {invite.roleName} · expires {invite.expiresAt || 'unknown'}
                      </p>
                    </div>
                    <div class="flex flex-wrap items-center gap-2">
                      <Badge content={invite.status} color={membershipStatusColor(invite.status)} />
                      <Button
                        color="red"
                        size="xs"
                        disabled={!canManageMembers() || invite.status !== 'pending'}
                        onClick={() => {
                          void handleRevokeInvite(invite.id, invite.tenantId, invite.email);
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
          title="Platform administration"
          subtitle="Assign or revoke platform-wide roles such as store creation and admin governance."
        />
        <Show when={!canManagePlatformRoles()}>
          <InfoAlert>
            Platform administration requires dedicated platform access.
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
              label="Platform admin role"
              value={platformRoleName()}
              options={platformRoleOptions}
              onChange={(event) => setPlatformRoleName(event.currentTarget.value)}
            />
          </div>
          <div class="flex flex-wrap gap-3">
            <Button
              type="button"
              color="light"
              disabled={!userID}
              onClick={() => setPlatformUserId(String(userID))}
            >
              Use my user id
            </Button>
            <Button
              type="submit"
              loading={savingPlatformRole()}
              disabled={!canManagePlatformRoles()}
            >
              Save admin role
            </Button>
            <Button
              type="button"
              color="alternative"
              disabled={!canManagePlatformRoles()}
              onClick={() => {
                void loadPlatformRoleAssignments();
              }}
            >
              Reload admin roles
            </Button>
          </div>
        </form>

        <Show when={loadingPlatformRoles()}>
          <LoadingInline label="Loading admin roles..." />
        </Show>

        <Show
          when={!loadingPlatformRoles() && platformRoles().length > 0}
          fallback={
            <EmptyBlock
              title="No admin roles loaded"
              copy="Choose a target user to inspect platform-level administration access."
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
                      <Badge content={membership.status} color={membershipStatusColor(membership.status)} />
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
