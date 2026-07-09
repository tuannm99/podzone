import {
    addOrganizationMember,
    attachServiceControlPolicy,
    attachTenantToOrganization,
    createOrganization,
    detachServiceControlPolicy,
    detachTenantFromOrganization,
    removeOrganizationMember,
} from '@podzone/shared/services/iam'
import type { AdminIamLoaders } from '../createAdminIamLoaders'
import type { AdminIamState } from '../createAdminIamState'
import type { RunAction } from '../shared/actions'

export function createOrganizationsActions(state: AdminIamState, loaders: AdminIamLoaders, runAction: RunAction) {
    const submitCreateOrganization = async (event: SubmitEvent) => {
        event.preventDefault()
        await runAction(async () => {
            const result = await createOrganization({
                name: state.orgName().trim(),
                slug: state.orgSlug().trim(),
            })
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Created organization ${result.data.organization?.slug || state.orgName()}.`)
            state.setOrgName('')
            state.setOrgSlug('')
            await state.reloadOrganizations()
        })
    }

    const handleAttachTenantToOrg = () =>
        runAction(async () => {
            const tenantID = state.orgTenantId().trim()
            const result = await attachTenantToOrganization(state.selectedOrgId().trim(), tenantID)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Attached tenant ${tenantID} to organization.`)
            await state.reloadOrganizations()
        })

    const handleDetachTenantFromOrg = (tenantID: string) =>
        runAction(async () => {
            const result = await detachTenantFromOrganization(state.selectedOrgId().trim(), tenantID)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Detached tenant ${tenantID} from organization.`)
            await state.reloadOrganizations()
        })

    const handleAttachScp = () =>
        runAction(async () => {
            const policyName = state.orgPolicyName().trim()
            const result = await attachServiceControlPolicy(state.selectedOrgId().trim(), policyName)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Attached SCP ${policyName}.`)
            state.setOrgPolicyName('')
            await loaders.loadSelectedOrganization()
            await loaders.loadSelectedPolicy()
        })

    const handleDetachScp = (policyName: string) =>
        runAction(async () => {
            const result = await detachServiceControlPolicy(state.selectedOrgId().trim(), policyName)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Detached SCP ${policyName}.`)
            await loaders.loadSelectedOrganization()
            await loaders.loadSelectedPolicy()
        })

    const handleAddOrganizationMember = (userID: number, roleName: string) =>
        runAction(async () => {
            const result = await addOrganizationMember(state.selectedOrgId().trim(), userID, roleName)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Added user ${userID} to the organization.`)
            await state.reloadOrganizationMembers()
        })

    const handleRemoveOrganizationMember = (userID: string) =>
        runAction(async () => {
            const result = await removeOrganizationMember(state.selectedOrgId().trim(), userID)
            if (!result.success) throw new Error(result.message)
            state.setPageMessage(`Removed user ${userID} from the organization.`)
            await state.reloadOrganizationMembers()
        })

    return {
        submitCreateOrganization,
        handleAttachTenantToOrg,
        handleDetachTenantFromOrg,
        handleAttachScp,
        handleDetachScp,
        handleAddOrganizationMember,
        handleRemoveOrganizationMember,
    }
}
