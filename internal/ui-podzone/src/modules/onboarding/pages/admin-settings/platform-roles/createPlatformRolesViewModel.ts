import { createSignal } from 'solid-js'
import {
  listPlatformRoles,
  removePlatformRole,
  upsertPlatformRole,
} from '@/services/iam'
import { createPaginatedResource } from '@/solid/pagination'
import type { PlatformRoleFormValues } from '../forms'
import { platformRoleOptions } from '../presentation'

export function createPlatformRolesViewModel(userID: number) {
  const [userId, setUserId] = createSignal(userID ? String(userID) : '')
  const [selectedUserID, setSelectedUserID] = createSignal(userID)
  const [roleName, setRoleName] = createSignal(platformRoleOptions[1].value)
  const [saving, setSaving] = createSignal(false)
  const [mutationError, setMutationError] = createSignal('')
  const [message, setMessage] = createSignal('')
  const roles = createPaginatedResource(
    {
      page: 1,
      pageSize: 10,
      sortBy: 'createdAt',
      sortDirection: 'SORT_DIRECTION_DESC',
    },
    async (query) => {
      const result = await listPlatformRoles(selectedUserID(), query)
      if (!result.success) throw new Error(result.message)
      return result.data
    },
    {
      enabled: () => Number.isFinite(selectedUserID()) && selectedUserID() > 0,
      dependency: selectedUserID,
    }
  )
  const canManage = () => roles.resolved()
  const error = () => mutationError()

  const reload = async () => {
    const nextUserID = Number.parseInt(userId().trim(), 10)
    if (!Number.isFinite(nextUserID) || nextUserID <= 0) return
    if (selectedUserID() === nextUserID) {
      await roles.reload()
      return
    }
    setSelectedUserID(nextUserID)
  }

  const save = async (values: PlatformRoleFormValues) => {
    const nextUserID = Number.parseInt(values.userId.trim(), 10)
    if (!Number.isFinite(nextUserID) || nextUserID <= 0) {
      setMutationError(
        'Target user id is required for platform administration roles.'
      )
      return
    }
    setSaving(true)
    setMutationError('')
    setMessage('')
    setUserId(values.userId.trim())
    setRoleName(values.roleName)
    try {
      const result = await upsertPlatformRole({
        targetUserId: nextUserID,
        roleName: values.roleName,
      })
      if (!result.success) {
        setMutationError(result.message)
        return
      }
      setMessage(
        `Saved platform admin role ${values.roleName} for user ${nextUserID}.`
      )
      if (selectedUserID() === nextUserID) {
        await roles.reload()
      } else {
        setSelectedUserID(nextUserID)
      }
    } finally {
      setSaving(false)
    }
  }

  const remove = async (memberUserID: number, targetRoleName: string) => {
    setMutationError('')
    setMessage('')
    const result = await removePlatformRole(memberUserID, targetRoleName)
    if (!result.success) {
      setMutationError(result.message)
      return
    }
    setMessage(
      `Removed platform admin role ${targetRoleName} from user ${memberUserID}.`
    )
    await roles.reload()
  }

  return {
    items: roles.items,
    query: roles.query,
    pageInfo: roles.pageInfo,
    loading: roles.loading,
    updateQuery: roles.updateQuery,
    collectionError: roles.error,
    saving,
    canManage,
    userID,
    userId,
    setUserId,
    roleName,
    setRoleName,
    error,
    message,
    reload,
    save,
    remove,
  }
}

export type PlatformRolesViewModel = ReturnType<
  typeof createPlatformRolesViewModel
>
