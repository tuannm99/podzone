import { prettyJSON } from './presentation'
import type { AdminIamState } from './useAdminIamState'

export function createSimulationPresets(state: AdminIamState) {
  const applyServiceAssumePreset = () => {
    state.setSimScope('platform')
    state.setSimAction('order:update')
    state.setSimResource('*')
    state.setSimServicePrincipal('backoffice.podzone.internal')
    state.setSimAssumedRoleScope('platform')
    state.setSimAssumedRoleName('platform_admin')
    state.setSimAssumedRoleSourceIdentity('backoffice-admin')
    state.setSimAssumedRoleSessionName('service-assume')
    state.setSimAssumedRoleServicePrincipal('backoffice.podzone.internal')
    state.setSimAttributesJson(prettyJSON({ lane: 'priority' }))
    state.setSimSessionTagsJson(
      prettyJSON({ team: 'ops', path: 'service-assume' })
    )
  }

  const applyTenantAssumePreset = () => {
    state.setSimScope('tenant')
    state.setSimAction('order:update')
    state.setSimResource('*')
    state.setSimAssumedRoleScope('tenant')
    state.setSimAssumedRoleName('tenant_admin')
    state.setSimAssumedRoleTenantId(state.simTenantId().trim())
    state.setSimAssumedRoleSessionName('tenant-admin-review')
    state.setSimAssumedRoleSourceIdentity('store-ops')
    state.setSimAssumedRoleServicePrincipal('')
    state.setSimAttributesJson(prettyJSON({ lane: 'priority', region: 'us' }))
    state.setSimSessionTagsJson(
      prettyJSON({
        team: 'ops',
        store: state.simTenantId().trim() || 'tenant',
      })
    )
  }

  const applyScopeDownDenyPreset = () => {
    state.setSimAction('order:update')
    state.setSimResource('*')
    state.setSimSessionPolicyJson(
      prettyJSON([
        {
          effect: 'deny',
          actionPattern: 'order:update',
          resourcePattern: '*',
          conditions: [],
        },
      ])
    )
    state.setSimAttributesJson(prettyJSON({ lane: 'restricted' }))
    state.setSimSessionTagsJson(prettyJSON({ team: 'ops', mode: 'scope-down' }))
  }

  return {
    applyServiceAssumePreset,
    applyTenantAssumePreset,
    applyScopeDownDenyPreset,
  }
}
