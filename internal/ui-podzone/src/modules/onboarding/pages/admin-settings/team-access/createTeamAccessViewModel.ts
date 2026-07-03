import { createSignal } from 'solid-js'
import {
  removeTenantMember,
  upsertTenantMember,
  upsertTenantMemberByIdentity,
} from '@/services/iam'
import type { TeamMemberFormValues } from '../forms'
import { roleOptions } from '../presentation'
import type { WorkspaceAccessViewModel } from './createWorkspaceAccessViewModel'

export function createTeamAccessViewModel(
  userID: number,
  access: WorkspaceAccessViewModel
) {
  const [userId, setUserId] = createSignal('')
  const [identity, setIdentity] = createSignal('')
  const [roleName, setRoleName] = createSignal(roleOptions[1].value)
  const [saving, setSaving] = createSignal(false)
  const [mutationError, setMutationError] = createSignal('')
  const [message, setMessage] = createSignal('')
  const error = () => mutationError() || access.error()

  const reload = async (tenantID = access.selectedTenantID().trim()) => {
    if (!tenantID) return
    access.setSelectedTenantID(tenantID)
    await access.members.reload()
  }

  const save = async (values: TeamMemberFormValues) => {
    const tenantID = values.tenantId.trim()
    const parsedUserID = Number.parseInt(values.userId.trim(), 10)
    const memberIdentity = values.identity.trim()
    if (
      !tenantID ||
      (!memberIdentity && (!Number.isFinite(parsedUserID) || parsedUserID <= 0))
    ) {
      setMutationError(
        'Workspace id and either teammate identity or user id are required.'
      )
      return
    }
    if (!access.canManage()) {
      setMutationError(
        'You do not have permission to manage team access for this workspace.'
      )
      return
    }

    setSaving(true)
    setMutationError('')
    setMessage('')
    access.setSelectedTenantID(tenantID)
    setUserId(values.userId.trim())
    setIdentity(memberIdentity)
    setRoleName(values.roleName)
    try {
      if (memberIdentity) {
        const result = await upsertTenantMemberByIdentity({
          tenantId: tenantID,
          identity: memberIdentity,
          roleName: values.roleName,
        })
        if (!result.success) {
          setMutationError(result.message)
          return
        }
        setUserId(result.data.userId ? String(result.data.userId) : '')
        setIdentity('')
        setMessage(
          result.data.createdUser
            ? `Created a new account and granted workspace access for ${memberIdentity} in ${tenantID}.`
            : `Granted workspace access to existing user ${result.data.userId} for ${memberIdentity} in ${tenantID}.`
        )
      } else {
        const result = await upsertTenantMember({
          tenantId: tenantID,
          userId: parsedUserID,
          roleName: values.roleName,
        })
        if (!result.success) {
          setMutationError(result.message)
          return
        }
        setUserId('')
        setMessage(
          `Saved workspace access for user ${parsedUserID} in ${tenantID}.`
        )
      }
      await reload(tenantID)
    } finally {
      setSaving(false)
    }
  }

  const remove = async (tenantID: string, memberUserID: number) => {
    if (!access.canManage()) {
      setMutationError('You do not have permission to remove workspace access.')
      return
    }
    setMutationError('')
    setMessage('')
    const result = await removeTenantMember(tenantID, memberUserID)
    if (!result.success) {
      setMutationError(result.message)
      return
    }
    setMessage(`Removed user ${memberUserID} from workspace ${tenantID}.`)
    await reload(tenantID)
  }

  return {
    userID,
    userId,
    setUserId,
    identity,
    setIdentity,
    roleName,
    setRoleName,
    saving,
    error,
    message,
    reload,
    save,
    remove,
    access,
  }
}

export type TeamAccessViewModel = ReturnType<typeof createTeamAccessViewModel>
