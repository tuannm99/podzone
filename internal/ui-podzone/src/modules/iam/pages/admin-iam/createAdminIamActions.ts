import type { AdminIamLoaders } from './createAdminIamLoaders'
import { createAdminIamOrgPolicyGroupActions } from './createAdminIamOrgPolicyGroupActions'
import { createAdminIamPrincipalTrustActions } from './createAdminIamPrincipalTrustActions'
import type { AdminIamState } from './createAdminIamState'

export function createAdminIamActions(
  state: AdminIamState,
  loaders: AdminIamLoaders
) {
  const runAction = async (work: () => Promise<void>) => {
    state.setPageError('')
    state.setPageMessage('')
    try {
      await work()
    } catch (error) {
      state.setPageError(
        error instanceof Error ? error.message : 'Action failed'
      )
    }
  }

  return {
    ...createAdminIamOrgPolicyGroupActions(state, loaders, runAction),
    ...createAdminIamPrincipalTrustActions(state, loaders, runAction),
  }
}

export type AdminIamActions = ReturnType<typeof createAdminIamActions>
