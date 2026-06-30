import { createSignal, onMount } from 'solid-js'
import {
  checkPlatformPermission,
  listPlatformRoles,
  removePlatformRole,
  type PlatformRoleMembership,
  upsertPlatformRole,
} from '@/services/iam'
import type { PlatformRoleFormValues } from '../forms'
import { platformRoleOptions } from '../presentation'

export function createPlatformRolesViewModel(userID: number) {
  const [items, setItems] = createSignal<PlatformRoleMembership[]>([])
  const [loading, setLoading] = createSignal(false)
  const [saving, setSaving] = createSignal(false)
  const [canManage, setCanManage] = createSignal(false)
  const [userId, setUserId] = createSignal(userID ? String(userID) : '')
  const [roleName, setRoleName] = createSignal(platformRoleOptions[1].value)
  const [error, setError] = createSignal('')
  const [message, setMessage] = createSignal('')

  const refreshPermission = async () => {
    if (!userID) {
      setCanManage(false)
      return
    }
    const result = await checkPlatformPermission('platform:manage_roles')
    if (!result.success) {
      setError(result.message)
      setCanManage(false)
      return
    }
    setCanManage(result.data)
    if (result.data) {
      await reload()
    } else {
      setItems([])
    }
  }

  const reload = async () => {
    const targetUserID = Number.parseInt(userId().trim(), 10)
    if (
      !userID ||
      !canManage() ||
      !Number.isFinite(targetUserID) ||
      targetUserID <= 0
    ) {
      setItems([])
      return
    }
    setLoading(true)
    setError('')
    try {
      const result = await listPlatformRoles(targetUserID)
      if (!result.success) {
        setError(result.message)
        setItems([])
        return
      }
      setItems(result.data)
    } finally {
      setLoading(false)
    }
  }

  const save = async (values: PlatformRoleFormValues) => {
    const targetUserID = Number.parseInt(values.userId.trim(), 10)
    if (!canManage()) {
      setError(
        'You do not have permission to manage platform administration roles.'
      )
      return
    }
    if (!Number.isFinite(targetUserID) || targetUserID <= 0) {
      setError('Target user id is required for platform administration roles.')
      return
    }
    setSaving(true)
    setError('')
    setMessage('')
    setUserId(values.userId.trim())
    setRoleName(values.roleName)
    try {
      const result = await upsertPlatformRole({
        targetUserId: targetUserID,
        roleName: values.roleName,
      })
      if (!result.success) {
        setError(result.message)
        return
      }
      setMessage(
        `Saved platform admin role ${values.roleName} for user ${targetUserID}.`
      )
      await reload()
    } finally {
      setSaving(false)
    }
  }

  const remove = async (targetUserID: number, targetRoleName: string) => {
    if (!canManage()) {
      setError(
        'You do not have permission to manage platform administration roles.'
      )
      return
    }
    setError('')
    setMessage('')
    const result = await removePlatformRole(targetUserID, targetRoleName)
    if (!result.success) {
      setError(result.message)
      return
    }
    setMessage(
      `Removed platform admin role ${targetRoleName} from user ${targetUserID}.`
    )
    await reload()
  }

  onMount(() => void refreshPermission())

  return {
    items,
    loading,
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
