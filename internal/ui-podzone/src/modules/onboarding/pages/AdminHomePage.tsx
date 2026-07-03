import {
  createEffect,
  createResource,
  createSignal,
  on,
  onCleanup,
  onMount,
} from 'solid-js'
import { ensureActiveTenant } from '@/services/auth'
import {
  checkPlatformPermission,
  createTenant,
  listUserTenants,
  type TenantMembership,
} from '@/services/iam'
import { createStoreRequest, retryStoreRequest } from '@/services/onboarding'
import { storeStorage } from '@/services/storeStorage'
import { tenantStorage } from '@/services/tenantStorage'
import { tokenStorage } from '@/services/tokenStorage'
import { AdminHomeContext } from './admin-home/context'
import { AdminHomeView } from './admin-home/AdminHomeView'
import {
  buildOrdersHref,
  membershipStatusColor,
  parseUserID,
  provisioningStatusLabel,
  slugify,
  type StoreAttention,
  type WorkspaceSummary,
} from './admin-home/presentation'
import { collectWorkspaceData } from './admin-home/loadWorkspaceData'
import { createStoreCollectionsViewModel } from './admin-home/createStoreCollectionsViewModel'
import type {
  CreateStoreFormValues,
  CreateWorkspaceFormValues,
} from './admin-home/forms'

export function createAdminHomeViewModel() {
  const user = tokenStorage.getUser()
  const userID = parseUserID(user?.id)

  const [tenantName, setTenantName] = createSignal('')
  const [tenantSlug, setTenantSlug] = createSignal('')
  const [tenantMutationError, setTenantError] = createSignal('')
  const [tenantMessage, setTenantMessage] = createSignal('')
  const [switchingTenant, setSwitchingTenant] = createSignal(false)
  const [creatingTenant, setCreatingTenant] = createSignal(false)
  const [creatingStoreTenantId, setCreatingStoreTenantId] = createSignal('')
  const [retryingStoreRequestId, setRetryingStoreRequestId] = createSignal('')
  const [storeNameByTenant, setStoreNameByTenant] = createSignal<
    Record<string, string>
  >({})
  const [selectedWorkspaceId, setSelectedWorkspaceId] = createSignal('')
  const [selectedStoreId, setSelectedStoreId] = createSignal('')
  const [workspaceDataResource, { refetch: refetchWorkspaceData }] =
    createResource(
      () => userID || undefined,
      async (
        currentUserID
      ): Promise<{
        memberships: TenantMembership[]
        summaries: WorkspaceSummary[]
        attention: StoreAttention[]
      }> => {
        const result = await listUserTenants(currentUserID)
        if (!result.success) throw new Error(result.message)
        const workspaceData = await collectWorkspaceData(result.data)
        return {
          memberships: result.data,
          ...workspaceData,
        }
      }
    )
  const memberships = () => workspaceDataResource()?.memberships || []
  const workspaceSummaries = () => workspaceDataResource()?.summaries || []
  const storeAttention = () => workspaceDataResource()?.attention || []
  const loadingTenants = () => workspaceDataResource.loading
  const loadingAttention = () => workspaceDataResource.loading
  const workspaceReadError = () =>
    workspaceDataResource.error instanceof Error
      ? workspaceDataResource.error.message
      : ''
  const [platformAccessResource, { refetch: refetchPlatformAccess }] =
    createResource(
      () =>
        userID && workspaceDataResource() !== undefined
          ? `${userID}:${memberships().length === 0}`
          : undefined,
      async () => {
        if (memberships().length === 0) {
          return { canCreateTenant: true, canManagePlatformIAM: false }
        }
        const [tenantResult, iamResult] = await Promise.all([
          checkPlatformPermission('tenant:create'),
          checkPlatformPermission('platform:manage_roles'),
        ])
        if (!tenantResult.success) throw new Error(tenantResult.message)
        return {
          canCreateTenant: tenantResult.data,
          canManagePlatformIAM: iamResult.success ? iamResult.data : false,
        }
      }
    )
  const canCreateTenant = () =>
    platformAccessResource()?.canCreateTenant || false
  const canManagePlatformIAM = () =>
    platformAccessResource()?.canManagePlatformIAM || false
  const checkingPlatformAccess = () => platformAccessResource.loading
  const tenantError = () =>
    tenantMutationError() ||
    workspaceReadError() ||
    (platformAccessResource.error instanceof Error
      ? platformAccessResource.error.message
      : '')
  const activeMemberships = () =>
    memberships().filter((membership) => membership.status === 'active')
  const canBootstrapFirstWorkspace = () =>
    !!userID &&
    workspaceDataResource() !== undefined &&
    memberships().length === 0
  const activeWorkspaceSummaries = () =>
    workspaceSummaries().filter((workspace) => workspace.status === 'active')
  const selectedWorkspace = () =>
    activeWorkspaceSummaries().find(
      (workspace) => workspace.tenantId === selectedWorkspaceId()
    )
  const selectedWorkspaceOptions = () =>
    activeWorkspaceSummaries().map((workspace) => ({
      name: `${workspace.tenantId} · ${workspace.roleName}`,
      value: workspace.tenantId,
    }))
  const storeCollections = createStoreCollectionsViewModel(
    selectedWorkspaceId,
    loadingTenants
  )
  const stores = storeCollections.stores
  const storeRequests = storeCollections.requests
  const currentSelectionLabel = () => {
    const workspace = selectedWorkspace()
    const store = stores.items().find((item) => item.id === selectedStoreId())
    if (!workspace) return 'No workspace selected'
    if (!store) return `${workspace.tenantId} selected, no store chosen`
    return `${workspace.tenantId} / ${store.name}`
  }

  createEffect(() => {
    const nextWorkspace =
      selectedWorkspaceId() ||
      activeWorkspaceSummaries()[0]?.tenantId ||
      activeMemberships()[0]?.tenantId ||
      ''
    if (!nextWorkspace) return
    if (nextWorkspace !== selectedWorkspaceId()) {
      setSelectedWorkspaceId(nextWorkspace)
    }
  })

  createEffect(() => {
    const items = stores.items()
    if (items.length === 0) {
      setSelectedStoreId('')
      return
    }
    const current = selectedStoreId()
    if (items.some((store) => store.id === current)) {
      return
    }
    const preferred = items.find((store) => store.isActive) || items[0]
    setSelectedStoreId(preferred?.id || '')
  })
  createEffect(
    on(
      selectedWorkspaceId,
      () => {
        stores.clear()
        storeRequests.clear()
        setSelectedStoreId('')
      },
      { defer: true }
    )
  )

  const loadMemberships = async () => {
    if (!userID) return
    await refetchWorkspaceData()
  }

  const loadPlatformAccess = async () => {
    if (!userID) return
    await refetchPlatformAccess()
  }

  const openStore = async (nextTenantID: string, nextStoreID: string) => {
    const normalizedTenantID = nextTenantID.trim()
    const normalizedStoreID = nextStoreID.trim()
    if (!normalizedTenantID || !normalizedStoreID) return

    setSwitchingTenant(true)
    setTenantError('')
    setTenantMessage('')
    try {
      const { success, data } = await ensureActiveTenant(normalizedTenantID)
      if (!success) {
        setTenantError(data.message || 'Failed to open store')
        return
      }

      tenantStorage.setTenantID(normalizedTenantID)
      storeStorage.setStoreID(normalizedTenantID, normalizedStoreID)
      window.location.href = `/t/${normalizedTenantID}?storeId=${encodeURIComponent(normalizedStoreID)}`
    } finally {
      setSwitchingTenant(false)
    }
  }

  const prepareTenant = async (nextTenantID: string) => {
    if (!nextTenantID) return

    setSwitchingTenant(true)
    setTenantError('')
    setTenantMessage('')
    try {
      const { success, data } = await ensureActiveTenant(nextTenantID)
      if (!success) {
        setTenantError(data.message || 'Failed to load workspace')
        return
      }
      tenantStorage.setTenantID(nextTenantID)
      setTenantMessage(
        `Loaded workspace ${nextTenantID}. Choose a store below.`
      )
      await loadMemberships()
    } finally {
      setSwitchingTenant(false)
    }
  }

  const setDraftStoreName = (nextTenantID: string, value: string) => {
    setStoreNameByTenant((current) => ({
      ...current,
      [nextTenantID]: value,
    }))
  }

  const createStoreFromForm = async (
    nextTenantID: string,
    values: CreateStoreFormValues
  ) => {
    const normalizedTenantID = nextTenantID.trim()
    const normalizedStoreName = values.name.trim()
    if (!normalizedTenantID || !normalizedStoreName) return false

    setCreatingStoreTenantId(normalizedTenantID)
    setTenantError('')
    setTenantMessage('')
    setDraftStoreName(normalizedTenantID, normalizedStoreName)
    try {
      const switched = await ensureActiveTenant(normalizedTenantID)
      if (!switched.success) {
        setTenantError(switched.data.message || 'Failed to load workspace')
        return false
      }
      const created = await createStoreRequest({
        tenantId: normalizedTenantID,
        name: normalizedStoreName,
        subdomain: slugify(normalizedStoreName),
      })
      if (!created.success) {
        setTenantError(created.message)
        return false
      }
      setDraftStoreName(normalizedTenantID, '')
      setTenantMessage(
        `Store request ${created.data.name} is ${created.data.status}. It will become selectable after provisioning completes.`
      )
      if (selectedWorkspaceId() === normalizedTenantID) {
        await storeRequests.reload()
      }
      await loadMemberships()
      return true
    } finally {
      setCreatingStoreTenantId('')
    }
  }

  const submitCreateStore = async (nextTenantID: string) => {
    await createStoreFromForm(nextTenantID, {
      name: storeNameByTenant()[nextTenantID.trim()] || '',
    })
  }

  const retryStore = async (tenantID: string, requestID: string) => {
    setRetryingStoreRequestId(requestID)
    setTenantError('')
    setTenantMessage('')
    try {
      const switched = await ensureActiveTenant(tenantID)
      if (!switched.success) {
        setTenantError(switched.data.message || 'Failed to load workspace')
        return
      }
      const result = await retryStoreRequest({
        tenantId: tenantID,
        requestId: requestID,
      })
      if (!result.success) {
        setTenantError(result.message)
        return
      }
      setTenantMessage('Store provisioning has been queued again.')
      if (selectedWorkspaceId() === tenantID) {
        await storeRequests.reload()
      }
      await loadMemberships()
    } finally {
      setRetryingStoreRequestId('')
    }
  }

  const createTenantFromForm = async (values: CreateWorkspaceFormValues) => {
    if (!userID) {
      setTenantError('No signed-in account found.')
      return false
    }
    if (!canCreateTenant() && !canBootstrapFirstWorkspace()) {
      setTenantError('Your account cannot create another workspace yet.')
      return false
    }

    const normalizedName = values.name.trim()
    const normalizedSlug = slugify(values.slug || normalizedName)
    if (!normalizedName || !normalizedSlug) {
      setTenantError('Workspace name and slug are required.')
      return false
    }

    setCreatingTenant(true)
    setTenantError('')
    setTenantMessage('')
    setTenantName(normalizedName)
    setTenantSlug(normalizedSlug)
    try {
      const result = await createTenant({
        name: normalizedName,
        slug: normalizedSlug,
      })
      if (!result.success) {
        setTenantError(result.message)
        return false
      }

      const createdTenantID = result.data.tenant?.id || ''
      const createdSlug = result.data.tenant?.slug || normalizedSlug
      setTenantName('')
      setTenantSlug('')
      setSelectedWorkspaceId(createdTenantID)
      setTenantMessage(
        createdTenantID
          ? `Created workspace ${createdSlug} (${createdTenantID}).`
          : `Created workspace ${createdSlug}.`
      )
      await loadMemberships()
      await loadPlatformAccess()
      return true
    } finally {
      setCreatingTenant(false)
    }
  }

  const submitCreateTenant = async (event: SubmitEvent) => {
    event.preventDefault()
    await createTenantFromForm({
      name: tenantName(),
      slug: tenantSlug(),
    })
  }

  onMount(() => {
    const refreshTimer = window.setInterval(() => {
      const hasActiveProvisioning = storeRequests
        .items()
        .some((request) =>
          [
            'requested',
            'planning',
            'planned',
            'pending_approval',
            'queued',
            'provisioning',
            'pending_platform_setup',
          ].includes(request.status)
        )
      if (hasActiveProvisioning && !storeRequests.loading()) {
        void storeRequests.reload()
        void stores.reload()
      }
    }, 10000)
    onCleanup(() => window.clearInterval(refreshTimer))
  })

  const viewModel = {
    user,
    userID,
    tenantName,
    setTenantName,
    tenantSlug,
    setTenantSlug,
    tenantError,
    tenantMessage,
    setTenantMessage,
    switchingTenant,
    creatingTenant,
    creatingStoreTenantId,
    retryingStoreRequestId,
    storeNameByTenant,
    loadingTenants,
    loadingAttention,
    checkingPlatformAccess,
    memberships,
    workspaceSummaries,
    storeAttention,
    canCreateTenant,
    canManagePlatformIAM,
    selectedWorkspaceId,
    setSelectedWorkspaceId,
    selectedStoreId,
    setSelectedStoreId,
    stores,
    storesError: storeCollections.storesError,
    storeRequests,
    storeRequestsError: storeCollections.requestsError,
    activeMemberships,
    canBootstrapFirstWorkspace,
    activeWorkspaceSummaries,
    selectedWorkspace,
    selectedWorkspaceOptions,
    currentSelectionLabel,
    slugify,
    membershipStatusColor,
    provisioningStatusLabel,
    buildOrdersHref,
    prepareTenant,
    openStore,
    setDraftStoreName,
    createStoreFromForm,
    submitCreateStore,
    retryStore,
    createTenantFromForm,
    submitCreateTenant,
  }

  return viewModel
}

export type AdminHomeViewModel = ReturnType<typeof createAdminHomeViewModel>

export default function AdminHomePage() {
  const viewModel = createAdminHomeViewModel()
  return (
    <AdminHomeContext.Provider value={viewModel}>
      <AdminHomeView />
    </AdminHomeContext.Provider>
  )
}
