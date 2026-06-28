import { createEffect, createResource, createSignal, onMount } from 'solid-js'
import {
  checkPermission,
  createTenantInvite,
  listTenantMembers,
  listTenantInvites,
  listUserTenants,
  revokeTenantInvite,
  removeTenantMember,
  upsertTenantMember,
  upsertTenantMemberByIdentity,
  type TenantMembership,
} from '@/services/iam'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'
import { parseUserID, roleOptions } from './admin-settings/presentation'
import { AdminSettingsContext } from './admin-settings/context'
import { AdminSettingsView } from './admin-settings/AdminSettingsView'
import type {
  TeamMemberFormValues,
  TenantInviteFormValues,
} from './admin-settings/forms'
import { createSessionAdmin } from './admin-settings/createSessionAdmin'
import { createPlatformRoleAdmin } from './admin-settings/createPlatformRoleAdmin'

export function createAdminSettingsViewModel() {
  const hasToken = Boolean(tokenStorage.getToken())
  const userID = parseUserID(tokenStorage.getUser()?.id)
  const activeTenantId = () => tokenStorage.getActiveTenantID()
  const sessionID = () => tokenStorage.getSessionID()

  const [routeTenantId, setRouteTenantID] = createSignal(
    tenantStorage.getTenantID()
  )
  const [memberTenantId, setMemberTenantId] = createSignal(
    activeTenantId() || tenantStorage.getTenantID()
  )
  const [memberUserId, setMemberUserId] = createSignal('')
  const [memberIdentity, setMemberIdentity] = createSignal('')
  const [roleName, setRoleName] = createSignal(roleOptions[1].value)
  const [inviteEmail, setInviteEmail] = createSignal('')
  const [inviteRoleName, setInviteRoleName] = createSignal(roleOptions[1].value)
  const [savingMember, setSavingMember] = createSignal(false)
  const [savingInvite, setSavingInvite] = createSignal(false)
  const [mutationError, setPageError] = createSignal('')
  const [memberActionMessage, setMemberActionMessage] = createSignal('')
  const [latestInviteAcceptURL, setLatestInviteAcceptURL] = createSignal('')
  const [membershipsResource, { refetch: refetchMemberships }] = createResource(
    () => userID || undefined,
    async (currentUserID): Promise<TenantMembership[]> => {
      const result = await listUserTenants(currentUserID)
      if (!result.success) throw new Error(result.message)
      return result.data
    }
  )
  const memberships = () => membershipsResource() || []
  const loadingTenants = () => membershipsResource.loading
  const [tenantAccessResource, { refetch: refetchTenantAccess }] =
    createResource(
      () => {
        const tenantId = memberTenantId().trim()
        return userID && tenantId ? { userID, tenantId } : undefined
      },
      async ({ userID: currentUserID, tenantId }) => {
        const [readResult, manageResult] = await Promise.all([
          checkPermission({
            tenantId,
            userId: currentUserID,
            permission: 'tenant:read',
          }),
          checkPermission({
            tenantId,
            userId: currentUserID,
            permission: 'tenant:manage_members',
          }),
        ])
        if (!readResult.success) throw new Error(readResult.message)
        if (!manageResult.success) throw new Error(manageResult.message)

        const [membersResult, invitesResult] = await Promise.all([
          readResult.data ? listTenantMembers(tenantId) : undefined,
          manageResult.data ? listTenantInvites(tenantId) : undefined,
        ])
        if (membersResult && !membersResult.success) {
          throw new Error(membersResult.message)
        }
        if (invitesResult && !invitesResult.success) {
          throw new Error(invitesResult.message)
        }
        return {
          canReadTenant: readResult.data,
          canManageMembers: manageResult.data,
          members: membersResult?.success ? membersResult.data : [],
          invites: invitesResult?.success ? invitesResult.data : [],
        }
      }
    )
  const members = () => tenantAccessResource()?.members || []
  const invites = () => tenantAccessResource()?.invites || []
  const canReadTenant = () => tenantAccessResource()?.canReadTenant || false
  const canManageMembers = () =>
    tenantAccessResource()?.canManageMembers || false
  const loadingMembers = () => tenantAccessResource.loading
  const loadingInvites = () => tenantAccessResource.loading
  const checkingPermissions = () => tenantAccessResource.loading
  const pageError = () =>
    mutationError() ||
    (membershipsResource.error instanceof Error
      ? membershipsResource.error.message
      : '') ||
    (tenantAccessResource.error instanceof Error
      ? tenantAccessResource.error.message
      : '')
  const sessionAdmin = createSessionAdmin(
    sessionID,
    setPageError,
    setMemberActionMessage
  )
  const platformRoleAdmin = createPlatformRoleAdmin(
    userID,
    setPageError,
    setMemberActionMessage
  )
  const tenantOptions = () =>
    memberships().map((membership) => ({
      name: `${membership.tenantId} · ${membership.roleName}`,
      value: membership.tenantId,
    }))

  const loadTenantMembers = async (tenantId = memberTenantId().trim()) => {
    if (!tenantId) return
    setPageError('')
    setMemberTenantId(tenantId)
    await refetchTenantAccess()
  }

  const loadTenantInvites = async (tenantId = memberTenantId().trim()) => {
    if (!tenantId) return
    setPageError('')
    setMemberTenantId(tenantId)
    await refetchTenantAccess()
  }

  const saveMemberFromForm = async (values: TeamMemberFormValues) => {
    const tenantId = values.tenantId.trim()
    const parsedUserId = Number.parseInt(values.userId.trim(), 10)
    const identity = values.identity.trim()
    if (
      !tenantId ||
      (identity === '' && (!Number.isFinite(parsedUserId) || parsedUserId <= 0))
    ) {
      setPageError(
        'Workspace id and either teammate identity or user id are required.'
      )
      return
    }
    if (!canManageMembers()) {
      setPageError(
        'You do not have permission to manage team access for this workspace.'
      )
      return
    }

    setSavingMember(true)
    setPageError('')
    setMemberActionMessage('')
    setMemberTenantId(tenantId)
    setMemberUserId(values.userId.trim())
    setMemberIdentity(identity)
    setRoleName(values.roleName)
    try {
      if (identity) {
        const result = await upsertTenantMemberByIdentity({
          tenantId,
          identity,
          roleName: values.roleName,
        })
        if (!result.success) {
          setPageError(result.message)
          return
        }
        setMemberUserId(result.data.userId ? String(result.data.userId) : '')
        setMemberIdentity('')
        setMemberActionMessage(
          result.data.createdUser
            ? `Created a new account and granted workspace access for ${identity} in ${tenantId}.`
            : `Granted workspace access to existing user ${result.data.userId} for ${identity} in ${tenantId}.`
        )
      } else {
        const result = await upsertTenantMember({
          tenantId,
          userId: parsedUserId,
          roleName: values.roleName,
        })
        if (!result.success) {
          setPageError(result.message)
          return
        }
        setMemberUserId('')
        setMemberActionMessage(
          `Saved workspace access for user ${parsedUserId} in ${tenantId}.`
        )
      }
      await loadTenantMembers(tenantId)
    } finally {
      setSavingMember(false)
    }
  }

  const handleRemoveMember = async (tenantId: string, userId: number) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to remove workspace access.')
      return
    }
    setPageError('')
    setMemberActionMessage('')
    const result = await removeTenantMember(tenantId, userId)
    if (!result.success) {
      setPageError(result.message)
      return
    }
    setMemberActionMessage(`Removed user ${userId} from workspace ${tenantId}.`)
    await loadTenantMembers(tenantId)
  }

  const createInviteFromForm = async (values: TenantInviteFormValues) => {
    const tenantId = values.tenantId.trim()
    const email = values.email.trim()
    if (!tenantId || !email) {
      setPageError('Workspace id and invite email are required.')
      return
    }
    if (!canManageMembers()) {
      setPageError('You do not have permission to manage workspace invites.')
      return
    }

    setSavingInvite(true)
    setPageError('')
    setMemberActionMessage('')
    setLatestInviteAcceptURL('')
    setMemberTenantId(tenantId)
    setInviteEmail(email)
    setInviteRoleName(values.roleName)
    try {
      const result = await createTenantInvite({
        tenantId,
        email,
        roleName: values.roleName,
      })
      if (!result.success) {
        setPageError(result.message)
        return
      }
      setInviteEmail('')
      setLatestInviteAcceptURL(result.data.acceptUrl)
      setMemberActionMessage(
        `Created a workspace invite for ${email} in ${tenantId}.`
      )
      await loadTenantInvites(tenantId)
    } finally {
      setSavingInvite(false)
    }
  }

  const handleRevokeInvite = async (
    inviteId: string,
    tenantId: string,
    email: string
  ) => {
    if (!canManageMembers()) {
      setPageError('You do not have permission to revoke workspace invites.')
      return
    }
    setPageError('')
    setMemberActionMessage('')
    const result = await revokeTenantInvite(inviteId)
    if (!result.success) {
      setPageError(result.message)
      return
    }
    setMemberActionMessage(`Revoked workspace invite for ${email}.`)
    await loadTenantInvites(tenantId)
  }

  onMount(() => {
    void platformRoleAdmin.refreshPlatformPermissions()
  })

  createEffect(() => {
    if (!memberTenantId().trim()) {
      const nextTenantId = activeTenantId() || memberships()[0]?.tenantId || ''
      if (nextTenantId) {
        setMemberTenantId(nextTenantId)
      }
    }
  })

  createEffect(() => {
    if (!platformRoleAdmin.canManagePlatformRoles()) {
      platformRoleAdmin.clearPlatformRoles()
      return
    }
    void platformRoleAdmin.loadPlatformRoleAssignments()
  })

  const viewModel = {
    hasToken,
    userID,
    activeTenantId,
    sessionID,
    routeTenantId,
    setRouteTenantID,
    memberTenantId,
    setMemberTenantId,
    memberUserId,
    setMemberUserId,
    memberIdentity,
    setMemberIdentity,
    roleName,
    setRoleName,
    inviteEmail,
    setInviteEmail,
    inviteRoleName,
    setInviteRoleName,
    memberships,
    members,
    invites,
    loadingTenants,
    loadingMembers,
    loadingInvites,
    savingMember,
    savingInvite,
    checkingPermissions,
    pageError,
    memberActionMessage,
    latestInviteAcceptURL,
    canReadTenant,
    canManageMembers,
    tenantOptions,
    refetchMemberships,
    ...sessionAdmin,
    ...platformRoleAdmin,
    loadTenantMembers,
    loadTenantInvites,
    handleRemoveMember,
    handleRevokeInvite,
    saveMemberFromForm,
    createInviteFromForm,
  }

  return viewModel
}

export type AdminSettingsViewModel = ReturnType<
  typeof createAdminSettingsViewModel
>

export default function AdminSettingsPage() {
  const viewModel = createAdminSettingsViewModel()
  return (
    <AdminSettingsContext.Provider value={viewModel}>
      <AdminSettingsView />
    </AdminSettingsContext.Provider>
  )
}
