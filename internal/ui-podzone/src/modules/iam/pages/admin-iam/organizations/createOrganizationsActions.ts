import {
  attachServiceControlPolicy,
  attachTenantToOrganization,
  createOrganization,
  detachServiceControlPolicy,
  detachTenantFromOrganization,
} from '@/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import type { RunAction } from '../shared/actions'

export function createOrganizationsActions(
  state: AdminIamState,
  loaders: AdminIamLoaders,
  runAction: RunAction
) {
  const submitCreateOrganization = async (event: SubmitEvent) => {
    event.preventDefault()
    await runAction(async () => {
      const result = await createOrganization({
        name: state.orgName().trim(),
        slug: state.orgSlug().trim(),
      })
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(
        `Created organization ${result.data.organization?.slug || state.orgName()}.`
      )
      state.setOrgName('')
      state.setOrgSlug('')
      await loaders.loadBootstrap()
    })
  }

  const handleAttachTenantToOrg = () =>
    runAction(async () => {
      const tenantID = state.orgTenantId().trim()
      const result = await attachTenantToOrganization(
        state.selectedOrgId().trim(),
        tenantID
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Attached tenant ${tenantID} to organization.`)
      await loaders.loadBootstrap()
    })

  const handleDetachTenantFromOrg = (tenantID: string) =>
    runAction(async () => {
      const result = await detachTenantFromOrganization(
        state.selectedOrgId().trim(),
        tenantID
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Detached tenant ${tenantID} from organization.`)
      await loaders.loadBootstrap()
    })

  const handleAttachScp = () =>
    runAction(async () => {
      const policyName = state.orgPolicyName().trim()
      const result = await attachServiceControlPolicy(
        state.selectedOrgId().trim(),
        policyName
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Attached SCP ${policyName}.`)
      state.setOrgPolicyName('')
      await loaders.loadSelectedOrganization()
      await loaders.loadSelectedPolicy()
    })

  const handleDetachScp = (policyName: string) =>
    runAction(async () => {
      const result = await detachServiceControlPolicy(
        state.selectedOrgId().trim(),
        policyName
      )
      if (!result.success) throw new Error(result.message)
      state.setPageMessage(`Detached SCP ${policyName}.`)
      await loaders.loadSelectedOrganization()
      await loaders.loadSelectedPolicy()
    })

  return {
    submitCreateOrganization,
    handleAttachTenantToOrg,
    handleDetachTenantFromOrg,
    handleAttachScp,
    handleDetachScp,
  }
}
