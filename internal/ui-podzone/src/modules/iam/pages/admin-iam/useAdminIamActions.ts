import type { AdminIamLoaders } from './useAdminIamLoaders'
import { useAdminIamOrgPolicyGroupActions } from './useAdminIamOrgPolicyGroupActions'
import { useAdminIamPrincipalTrustActions } from './useAdminIamPrincipalTrustActions'
import type { AdminIamState } from './useAdminIamState'

export function useAdminIamActions(
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
    ...useAdminIamOrgPolicyGroupActions(state, loaders, runAction),
    ...useAdminIamPrincipalTrustActions(state, loaders, runAction),
  }
}

export type AdminIamActions = ReturnType<typeof useAdminIamActions>
