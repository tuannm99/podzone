import { createSignal } from 'solid-js'
import {
  checkPlatformPermission,
  listPlatformRoles,
  removePlatformRole,
  type PlatformRoleMembership,
  upsertPlatformRole,
} from '@/services/iam'
import type { PlatformRoleFormValues } from './forms'
import { platformRoleOptions } from './presentation'

type MessageSetter = (message: string) => void

export function createPlatformRoleAdmin(
  userID: number | undefined,
  setPageError: MessageSetter,
  setActionMessage: MessageSetter
) {
  const [platformRoles, setPlatformRoles] = createSignal<
    PlatformRoleMembership[]
  >([])
  const [loadingPlatformRoles, setLoadingPlatformRoles] = createSignal(false)
  const [savingPlatformRole, setSavingPlatformRole] = createSignal(false)
  const [canManagePlatformRoles, setCanManagePlatformRoles] =
    createSignal(false)
  const [platformUserId, setPlatformUserId] = createSignal(
    userID ? String(userID) : ''
  )
  const [platformRoleName, setPlatformRoleName] = createSignal(
    platformRoleOptions[1].value
  )

  const refreshPlatformPermissions = async () => {
    if (!userID) {
      setCanManagePlatformRoles(false)
      return
    }
    const result = await checkPlatformPermission('platform:manage_roles')
    if (!result.success) {
      setPageError(result.message)
      setCanManagePlatformRoles(false)
      return
    }
    setCanManagePlatformRoles(result.data)
  }

  const loadPlatformRoleAssignments = async () => {
    const targetUserID = Number.parseInt(platformUserId().trim(), 10)
    if (
      !userID ||
      !canManagePlatformRoles() ||
      !Number.isFinite(targetUserID) ||
      targetUserID <= 0
    ) {
      setPlatformRoles([])
      return
    }
    setLoadingPlatformRoles(true)
    try {
      const result = await listPlatformRoles(targetUserID)
      if (!result.success) {
        setPageError(result.message)
        setPlatformRoles([])
        return
      }
      setPlatformRoles(result.data)
    } finally {
      setLoadingPlatformRoles(false)
    }
  }

  const savePlatformRoleFromForm = async (values: PlatformRoleFormValues) => {
    const targetUserID = Number.parseInt(values.userId.trim(), 10)
    if (!canManagePlatformRoles()) {
      setPageError(
        'You do not have permission to manage platform administration roles.'
      )
      return
    }
    if (!Number.isFinite(targetUserID) || targetUserID <= 0) {
      setPageError(
        'Target user id is required for platform administration roles.'
      )
      return
    }
    setSavingPlatformRole(true)
    setPageError('')
    setActionMessage('')
    setPlatformUserId(values.userId.trim())
    setPlatformRoleName(values.roleName)
    try {
      const result = await upsertPlatformRole({
        targetUserId: targetUserID,
        roleName: values.roleName,
      })
      if (!result.success) {
        setPageError(result.message)
        return
      }
      setActionMessage(
        `Saved platform admin role ${values.roleName} for user ${targetUserID}.`
      )
      await loadPlatformRoleAssignments()
    } finally {
      setSavingPlatformRole(false)
    }
  }

  const handleRemovePlatformRole = async (
    targetUserID: number,
    roleName: string
  ) => {
    if (!canManagePlatformRoles()) {
      setPageError(
        'You do not have permission to manage platform administration roles.'
      )
      return
    }
    setPageError('')
    setActionMessage('')
    const result = await removePlatformRole(targetUserID, roleName)
    if (!result.success) {
      setPageError(result.message)
      return
    }
    setActionMessage(
      `Removed platform admin role ${roleName} from user ${targetUserID}.`
    )
    await loadPlatformRoleAssignments()
  }

  return {
    platformRoles,
    clearPlatformRoles: () => setPlatformRoles([]),
    loadingPlatformRoles,
    savingPlatformRole,
    canManagePlatformRoles,
    platformUserId,
    setPlatformUserId,
    platformRoleName,
    setPlatformRoleName,
    refreshPlatformPermissions,
    loadPlatformRoleAssignments,
    savePlatformRoleFromForm,
    handleRemovePlatformRole,
  }
}
