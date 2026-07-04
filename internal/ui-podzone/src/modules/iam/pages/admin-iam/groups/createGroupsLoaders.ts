import type { AdminIamState } from '../createAdminIamState'

export function createGroupsLoaders(state: AdminIamState) {
  const loadGroupsForScope = async () => {
    if (!state.allowed() || !state.canManagePlatform()) return
    await state.reloadGroups()
  }

  const loadSelectedGroup = async () => {
    if (!state.canManagePlatform()) return
    const rawID = state.selectedGroupId().trim()
    if (!rawID) {
      state.clearGroupMembers()
      state.clearGroupPolicies()
      state.clearGroupInlinePolicies()
      return
    }
    const groupID = Number.parseInt(rawID, 10)
    if (!Number.isFinite(groupID) || groupID <= 0) return
    await Promise.all([
      state.reloadGroupMembers(),
      state.reloadGroupPolicies(),
      state.reloadGroupInlinePolicies(),
    ])
  }

  return { loadGroupsForScope, loadSelectedGroup }
}
