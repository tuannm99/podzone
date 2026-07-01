import {
  listGroupInlinePolicies,
  listGroupMembers,
  listGroupPolicies,
  listGroups,
} from '@/services/iam'
import type { AdminIamState } from '../createAdminIamState'

export function createGroupsLoaders(state: AdminIamState) {
  let scopeRequestID = 0
  let selectionRequestID = 0

  const loadGroupsForScope = async () => {
    if (!state.allowed()) return
    const currentRequest = ++scopeRequestID
    const result = await listGroups(
      state.groupScope(),
      state.groupScope() === 'tenant' ? state.groupTenantId().trim() : undefined
    )
    if (currentRequest !== scopeRequestID) return
    if (!result.success) {
      state.setPageError(result.message)
      return
    }
    state.setGroups(result.data)
  }

  const loadSelectedGroup = async () => {
    const currentRequest = ++selectionRequestID
    const rawID = state.selectedGroupId().trim()
    if (!rawID) {
      state.setGroupMembers([])
      state.setGroupPolicies([])
      state.setGroupInlinePolicies([])
      return
    }
    const groupID = Number.parseInt(rawID, 10)
    if (!Number.isFinite(groupID) || groupID <= 0) return
    const [membersResult, policiesResult, inlineResult] = await Promise.all([
      listGroupMembers(groupID),
      listGroupPolicies(groupID),
      listGroupInlinePolicies(groupID),
    ])
    if (
      currentRequest !== selectionRequestID ||
      rawID !== state.selectedGroupId().trim()
    )
      return
    if (membersResult.success) state.setGroupMembers(membersResult.data)
    else state.setPageError(membersResult.message)
    if (policiesResult.success) state.setGroupPolicies(policiesResult.data)
    else state.setPageError(policiesResult.message)
    if (inlineResult.success) state.setGroupInlinePolicies(inlineResult.data)
    else state.setPageError(inlineResult.message)
  }

  return { loadGroupsForScope, loadSelectedGroup }
}
