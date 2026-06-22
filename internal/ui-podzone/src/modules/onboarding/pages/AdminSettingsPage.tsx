import { createEffect, createSignal, onMount } from 'solid-js';
import {
  listAuditLogs,
  listSessions,
  revokeSession,
  type AuditLogInfo,
  type SessionInfo,
} from '@/services/auth';
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
} from '@/services/iam';
import { tenantStorage } from '@/services/tenantStorage';
import { tokenStorage } from '@/services/tokenStorage';
import {
  parseUserID,
  platformRoleOptions,
  roleOptions,
} from './admin-settings/presentation';
import { AdminSettingsContext } from './admin-settings/context';
import { AdminSettingsView } from './admin-settings/AdminSettingsView';

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
      setPageError('Workspace id and either teammate identity or user id are required.');
      return;
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage team access for this workspace.');
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
            ? `Created a new account and granted workspace access for ${identity} in ${tenantId}.`
            : `Granted workspace access to existing user ${result.data.userId} for ${identity} in ${tenantId}.`
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
        setMemberActionMessage(`Saved workspace access for user ${parsedUserId} in ${tenantId}.`);
      }
      await loadTenantMembers(tenantId);
    } finally {
      setSavingMember(false);
    }
  };

  const handleRemoveMember = async (tenantId: string, userId: number) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to remove workspace access.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await removeTenantMember(tenantId, userId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Removed user ${userId} from workspace ${tenantId}.`);
    await loadTenantMembers(tenantId);
  };

  const submitInvite = async (event: SubmitEvent) => {
    event.preventDefault();
    const tenantId = memberTenantId().trim();
    const email = inviteEmail().trim();
    if (!tenantId || !email) {
      setPageError('Workspace id and invite email are required.');
      return;
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage workspace invites.');
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
      setMemberActionMessage(`Created a workspace invite for ${email} in ${tenantId}.`);
      await loadTenantInvites(tenantId);
    } finally {
      setSavingInvite(false);
    }
  };

  const handleRevokeInvite = async (inviteId: string, tenantId: string, email: string) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to revoke workspace invites.');
      return;
    }
    setPageError('');
    setMemberActionMessage('');
    const result = await revokeTenantInvite(inviteId);
    if (!result.success) {
      setPageError(result.message);
      return;
    }
    setMemberActionMessage(`Revoked workspace invite for ${email}.`);
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

  const viewModel = {
    hasToken, userID, activeTenantId, sessionID, routeTenantId, setRouteTenantID, memberTenantId, setMemberTenantId, memberUserId, setMemberUserId, memberIdentity, setMemberIdentity, roleName, setRoleName, inviteEmail, setInviteEmail, inviteRoleName, setInviteRoleName, memberships, members, invites, platformRoles, sessions, auditLogs, loadingTenants, loadingMembers, loadingInvites, loadingSessions, loadingAuditLogs, savingMember, savingInvite, checkingPermissions, loadingPlatformRoles, savingPlatformRole, pageError, memberActionMessage, latestInviteAcceptURL, canReadTenant, canManageMembers, canManagePlatformRoles, platformUserId, setPlatformUserId, platformRoleName, setPlatformRoleName, tenantOptions, currentSessionCount, otherSessionCount, loadSessions, loadAuditLogs, loadTenantMembers, loadTenantInvites, loadPlatformRoleAssignments, handleRevokeSession, handleRemoveMember, handleRevokeInvite, handleRemovePlatformRole, submitMember, submitInvite, submitPlatformRole,
  };

  return (
    <AdminSettingsContext.Provider value={viewModel}>
      <AdminSettingsView />
    </AdminSettingsContext.Provider>
  );
}
